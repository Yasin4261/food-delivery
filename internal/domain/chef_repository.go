package domain

import "context"

// ChefRepository is the port for chef persistence. Lookups return
// ErrChefNotFound when no row matches.
type ChefRepository interface {
	Create(ctx context.Context, chef *Chef) error
	FindByID(ctx context.Context, id int) (*Chef, error)
	FindByUserID(ctx context.Context, userID int) (*Chef, error)
	List(ctx context.Context, limit, offset int) ([]*Chef, error)
	// FindNearby returns active chefs whose delivery radius covers (lat, lng),
	// nearest first.
	FindNearby(ctx context.Context, lat, lng float64, limit int) ([]*Chef, error)
}
