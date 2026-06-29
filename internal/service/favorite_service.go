package service

import (
	"context"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// FavoriteService implements the customer "favorite a chef" use cases. It
// depends only on domain ports.
type FavoriteService struct {
	favorites domain.FavoriteRepository
	chefs     domain.ChefRepository
}

// NewFavoriteService builds a FavoriteService.
func NewFavoriteService(favorites domain.FavoriteRepository, chefs domain.ChefRepository) *FavoriteService {
	return &FavoriteService{favorites: favorites, chefs: chefs}
}

// Add favorites a chef for the user. The chef must exist. Favoriting is
// idempotent.
func (s *FavoriteService) Add(ctx context.Context, userID, chefID int) error {
	if _, err := s.chefs.FindByID(ctx, chefID); err != nil {
		return err
	}
	return s.favorites.Add(ctx, userID, chefID)
}

// Remove unfavorites a chef for the user (a no-op if it was not favorited).
func (s *FavoriteService) Remove(ctx context.Context, userID, chefID int) error {
	return s.favorites.Remove(ctx, userID, chefID)
}

// List returns the chefs the user has favorited.
func (s *FavoriteService) List(ctx context.Context, userID, limit, offset int) ([]*domain.Chef, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.favorites.ListChefs(ctx, userID, limit, offset)
}
