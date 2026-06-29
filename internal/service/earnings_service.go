package service

import (
	"context"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// EarningsService implements the "chef sees their earnings" use case.
type EarningsService struct {
	earnings domain.EarningsRepository
	chefs    domain.ChefRepository
}

// NewEarningsService builds an EarningsService.
func NewEarningsService(earnings domain.EarningsRepository, chefs domain.ChefRepository) *EarningsService {
	return &EarningsService{earnings: earnings, chefs: chefs}
}

// ForChef returns the earnings of the caller's chef profile. days > 0 limits
// the window to the last N days; days <= 0 means all-time.
func (s *EarningsService) ForChef(ctx context.Context, userID, days int) (*domain.Earnings, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var since *time.Time
	if days > 0 {
		t := time.Now().AddDate(0, 0, -days)
		since = &t
	}
	return s.earnings.ChefEarnings(ctx, chef.ID, since)
}
