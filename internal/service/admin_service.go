package service

import (
	"context"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AdminService implements the admin/moderation use cases. Access is gated to
// the admin role at the router; this layer clears secrets and normalises
// paging.
type AdminService struct {
	admin  domain.AdminRepository
	promos domain.PromoRepository
}

// NewAdminService builds an AdminService. promos powers promo-code management.
func NewAdminService(admin domain.AdminRepository, promos domain.PromoRepository) *AdminService {
	return &AdminService{admin: admin, promos: promos}
}

// PromoInput is the data needed to create a promo code.
type PromoInput struct {
	Code          string
	DiscountType  string
	DiscountValue float64
	MinOrder      float64
	ValidFrom     *time.Time
	ValidUntil    *time.Time
	UsageLimit    int
}

// CreatePromo validates and persists a new promo code.
func (s *AdminService) CreatePromo(ctx context.Context, in PromoInput) (*domain.PromoCode, error) {
	p := &domain.PromoCode{
		Code:          in.Code,
		DiscountType:  in.DiscountType,
		DiscountValue: in.DiscountValue,
		MinOrder:      in.MinOrder,
		ValidFrom:     in.ValidFrom,
		ValidUntil:    in.ValidUntil,
		UsageLimit:    in.UsageLimit,
		IsActive:      true,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.promos.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPromos returns a page of all promo codes and the total.
func (s *AdminService) ListPromos(ctx context.Context, limit, offset int) ([]*domain.PromoCode, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.promos.List(ctx, limit, offset)
}

// SetPromoActive activates or deactivates a promo code.
func (s *AdminService) SetPromoActive(ctx context.Context, id int, active bool) error {
	return s.promos.SetActive(ctx, id, active)
}

// ListUsers returns a page of all users (password hashes cleared) and the total.
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, int, error) {
	limit, offset = normalisePaging(limit, offset)
	users, total, err := s.admin.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	return users, total, nil
}

// SetUserActive activates or deactivates a user (deactivation blocks login).
func (s *AdminService) SetUserActive(ctx context.Context, userID int, active bool) error {
	return s.admin.SetUserActive(ctx, userID, active)
}

// ListChefs returns a page of all chefs (including inactive) and the total.
func (s *AdminService) ListChefs(ctx context.Context, limit, offset int) ([]*domain.Chef, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListChefs(ctx, limit, offset)
}

// SetChefActive activates or deactivates a chef (deactivation hides them from
// browse/search and blocks new orders).
func (s *AdminService) SetChefActive(ctx context.Context, chefID int, active bool) error {
	return s.admin.SetChefActive(ctx, chefID, active)
}

// ListOrders returns a page of all orders (the platform overview) and the total.
func (s *AdminService) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.admin.ListOrders(ctx, limit, offset)
}

// Stats returns the aggregated platform dashboard figures.
func (s *AdminService) Stats(ctx context.Context) (*domain.PlatformStats, error) {
	return s.admin.Stats(ctx)
}
