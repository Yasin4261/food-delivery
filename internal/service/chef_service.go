package service

import (
	"context"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChefService implements the chef-profile use cases. It depends only on the
// domain.ChefRepository port.
type ChefService struct {
	chefs domain.ChefRepository
}

// NewChefService builds a ChefService.
func NewChefService(chefs domain.ChefRepository) *ChefService {
	return &ChefService{chefs: chefs}
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

// Get returns a chef by id.
func (s *ChefService) Get(ctx context.Context, id int) (*domain.Chef, error) {
	return s.chefs.FindByID(ctx, id)
}

// List returns a page of chefs. limit is clamped to a sane range.
func (s *ChefService) List(ctx context.Context, limit, offset int) ([]*domain.Chef, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.chefs.List(ctx, limit, offset)
}

// Nearby returns chefs that can deliver to (lat, lng).
func (s *ChefService) Nearby(ctx context.Context, lat, lng float64, limit int) ([]*domain.Chef, error) {
	limit, _ = normalisePaging(limit, 0)
	return s.chefs.FindNearby(ctx, lat, lng, limit)
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
