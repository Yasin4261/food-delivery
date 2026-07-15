package service

import (
	"context"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AdminService implements the admin/moderation use cases. Access is gated to
// the admin role at the router; this layer clears secrets and normalises
// paging.
type AdminService struct {
	admin domain.AdminRepository
}

// NewAdminService builds an AdminService.
func NewAdminService(admin domain.AdminRepository) *AdminService {
	return &AdminService{admin: admin}
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
