package service

import (
	"context"
	"strings"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChefService implements the chef-profile use cases. It depends only on the
// domain.ChefRepository port.
type ChefService struct {
	chefs domain.ChefRepository
	hours domain.ChefHoursRepository // nil disables working-hours features
	loc   *time.Location             // platform TZ for open/closed evaluation
}

// NewChefService builds a ChefService. hours powers the working-hours
// schedule (nil disables it); loc is the platform time zone (nil = UTC).
func NewChefService(chefs domain.ChefRepository, hours domain.ChefHoursRepository, loc *time.Location) *ChefService {
	if loc == nil {
		loc = time.UTC
	}
	return &ChefService{chefs: chefs, hours: hours, loc: loc}
}

// CreateProfileInput is the data needed to open a chef profile.
type CreateProfileInput struct {
	BusinessName   string
	KitchenAddress string
	Bio            string
	Specialty      string
	KitchenCity    string
	Latitude       *float64
	Longitude      *float64
	DeliveryRadius int
}

// CreateProfile opens a chef profile for an authenticated user. Each user may
// own at most one profile.
func (s *ChefService) CreateProfile(ctx context.Context, userID int, in CreateProfileInput) (*domain.Chef, error) {
	in.BusinessName = strings.TrimSpace(in.BusinessName)
	in.KitchenAddress = strings.TrimSpace(in.KitchenAddress)

	if in.BusinessName == "" {
		return nil, ValidationError{Msg: "business_name is required"}
	}
	if in.KitchenAddress == "" {
		return nil, ValidationError{Msg: "kitchen_address is required"}
	}
	if in.DeliveryRadius < 0 {
		return nil, ValidationError{Msg: "delivery_radius cannot be negative"}
	}
	if (in.Latitude == nil) != (in.Longitude == nil) {
		return nil, ValidationError{Msg: "latitude and longitude must be provided together"}
	}

	if _, err := s.chefs.FindByUserID(ctx, userID); err == nil {
		return nil, domain.ErrChefProfileExists
	} else if err != domain.ErrChefNotFound {
		return nil, err
	}

	chef := domain.NewChef(userID, in.BusinessName, in.KitchenAddress)
	chef.Bio = optional(in.Bio)
	chef.Specialty = optional(in.Specialty)
	chef.KitchenCity = optional(in.KitchenCity)
	chef.KitchenLatitude = in.Latitude
	chef.KitchenLongitude = in.Longitude
	if in.DeliveryRadius > 0 {
		chef.DeliveryRadius = in.DeliveryRadius
	}

	if err := s.chefs.Create(ctx, chef); err != nil {
		return nil, err
	}
	return chef, nil
}

// UpdateProfile edits the caller's own kitchen profile (same validation as
// CreateProfile). A zero DeliveryRadius keeps the current one.
func (s *ChefService) UpdateProfile(ctx context.Context, userID int, in CreateProfileInput) (*domain.Chef, error) {
	in.BusinessName = strings.TrimSpace(in.BusinessName)
	in.KitchenAddress = strings.TrimSpace(in.KitchenAddress)

	if in.BusinessName == "" {
		return nil, ValidationError{Msg: "business_name is required"}
	}
	if in.KitchenAddress == "" {
		return nil, ValidationError{Msg: "kitchen_address is required"}
	}
	if in.DeliveryRadius < 0 {
		return nil, ValidationError{Msg: "delivery_radius cannot be negative"}
	}
	if (in.Latitude == nil) != (in.Longitude == nil) {
		return nil, ValidationError{Msg: "latitude and longitude must be provided together"}
	}

	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	chef.BusinessName = in.BusinessName
	chef.KitchenAddress = in.KitchenAddress
	chef.Bio = optional(in.Bio)
	chef.Specialty = optional(in.Specialty)
	chef.KitchenCity = optional(in.KitchenCity)
	chef.KitchenLatitude = in.Latitude
	chef.KitchenLongitude = in.Longitude
	if in.DeliveryRadius > 0 {
		chef.DeliveryRadius = in.DeliveryRadius
	}

	if err := s.chefs.Update(ctx, chef); err != nil {
		return nil, err
	}
	return chef, nil
}

// Get returns a chef by id, with the derived open/closed state.
func (s *ChefService) Get(ctx context.Context, id int) (*domain.Chef, error) {
	chef, err := s.chefs.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.decorateOpenNow(ctx, []*domain.Chef{chef}); err != nil {
		return nil, err
	}
	return chef, nil
}

// MyProfile returns the chef profile owned by userID, or ErrChefNotFound when
// the user has not opened one yet.
func (s *ChefService) MyProfile(ctx context.Context, userID int) (*domain.Chef, error) {
	return s.chefs.FindByUserID(ctx, userID)
}

// List returns a page of chefs narrowed/ordered by f. limit is clamped to a
// sane range; unknown sorts and out-of-range ratings are rejected.
func (s *ChefService) List(ctx context.Context, f domain.ChefListFilters, limit, offset int) ([]*domain.Chef, int, error) {
	if !chefSorts[f.Sort] {
		return nil, 0, ValidationError{Msg: "unknown sort: must be rating or popular"}
	}
	if f.MinRating < 0 || f.MinRating > 5 {
		return nil, 0, ValidationError{Msg: "min_rating must be between 0 and 5"}
	}
	limit, offset = normalisePaging(limit, offset)
	chefs, total, err := s.chefs.List(ctx, f, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	if err := s.decorateOpenNow(ctx, chefs); err != nil {
		return nil, 0, err
	}
	return chefs, total, nil
}

// Nearby returns chefs that can deliver to (lat, lng); onlineOnly restricts to
// chefs currently online.
func (s *ChefService) Nearby(ctx context.Context, lat, lng float64, limit int, onlineOnly bool) ([]*domain.Chef, error) {
	limit, _ = normalisePaging(limit, 0)
	chefs, err := s.chefs.FindNearby(ctx, lat, lng, limit, onlineOnly)
	if err != nil {
		return nil, err
	}
	if err := s.decorateOpenNow(ctx, chefs); err != nil {
		return nil, err
	}
	return chefs, nil
}

// SetOnline toggles the live presence of the caller's chef profile and returns
// the updated profile.
func (s *ChefService) SetOnline(ctx context.Context, userID int, online bool) (*domain.Chef, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.chefs.SetOnline(ctx, chef.ID, online); err != nil {
		return nil, err
	}
	chef.SetOnline(online)
	return chef, nil
}

// SetAcceptingOrders toggles the caller's availability (away / vacation mode)
// and returns the updated profile. When off, the chef is hidden from
// browse/search and PlaceOrder rejects new orders for them.
func (s *ChefService) SetAcceptingOrders(ctx context.Context, userID int, accepting bool) (*domain.Chef, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.chefs.SetAcceptingOrders(ctx, chef.ID, accepting); err != nil {
		return nil, err
	}
	chef.SetAcceptingOrders(accepting)
	return chef, nil
}

// optional turns a trimmed string into a pointer, or nil when empty.
func optional(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func normalisePaging(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// SetHours replaces the caller's whole weekly schedule. An empty list clears
// it (always open). Every window is validated before anything persists.
func (s *ChefService) SetHours(ctx context.Context, userID int, hours []*domain.ChefHours) ([]*domain.ChefHours, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, h := range hours {
		if err := h.Validate(); err != nil {
			return nil, err
		}
		h.ChefID = chef.ID
	}
	if err := s.hours.ReplaceForChef(ctx, chef.ID, hours); err != nil {
		return nil, err
	}
	return s.hours.ListByChef(ctx, chef.ID)
}

// HoursFor returns a chef's public weekly schedule.
func (s *ChefService) HoursFor(ctx context.Context, chefID int) ([]*domain.ChefHours, error) {
	return s.hours.ListByChef(ctx, chefID)
}

// decorateOpenNow stamps IsOpenNow on the chefs from their schedules in one
// batched query. Chefs without windows are open by definition.
func (s *ChefService) decorateOpenNow(ctx context.Context, chefs []*domain.Chef) error {
	if s.hours == nil || len(chefs) == 0 {
		return nil
	}
	ids := make([]int, 0, len(chefs))
	for _, c := range chefs {
		ids = append(ids, c.ID)
	}
	byChef, err := s.hours.ListByChefs(ctx, ids)
	if err != nil {
		return err
	}
	now := time.Now().In(s.loc)
	for _, c := range chefs {
		open := domain.IsOpenAt(byChef[c.ID], now)
		c.IsOpenNow = &open
	}
	return nil
}
