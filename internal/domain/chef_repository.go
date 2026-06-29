package domain

import "context"

// ChefRepository is the port for chef persistence. Lookups return
// ErrChefNotFound when no row matches.
type ChefRepository interface {
	Create(ctx context.Context, chef *Chef) error
	FindByID(ctx context.Context, id int) (*Chef, error)
	FindByUserID(ctx context.Context, userID int) (*Chef, error)
	// List returns a page of active chefs (onlineOnly restricts to chefs
	// currently online) plus the total matching count.
	List(ctx context.Context, limit, offset int, onlineOnly bool) ([]*Chef, int, error)
	// FindNearby returns active chefs whose delivery radius covers (lat, lng),
	// nearest first; onlineOnly restricts to chefs currently online.
	FindNearby(ctx context.Context, lat, lng float64, limit int, onlineOnly bool) ([]*Chef, error)
	// SetOnline updates a chef's live presence flag.
	SetOnline(ctx context.Context, chefID int, online bool) error
}
