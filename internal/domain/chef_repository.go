package domain

import "context"

// ChefListFilters narrows and orders the public chef listing. Zero values
// mean "no constraint"; Sort follows the same whitelist as SearchFilters
// (rating / popular).
type ChefListFilters struct {
	OnlineOnly bool
	MinRating  float64
	Sort       string
}

// ChefRepository is the port for chef persistence. Lookups return
// ErrChefNotFound when no row matches.
type ChefRepository interface {
	Create(ctx context.Context, chef *Chef) error
	FindByID(ctx context.Context, id int) (*Chef, error)
	FindByUserID(ctx context.Context, userID int) (*Chef, error)
	// List returns a page of active chefs narrowed/ordered by f, plus the
	// total matching count.
	List(ctx context.Context, f ChefListFilters, limit, offset int) ([]*Chef, int, error)
	// FindNearby returns active chefs whose delivery radius covers (lat, lng),
	// nearest first; onlineOnly restricts to chefs currently online.
	FindNearby(ctx context.Context, lat, lng float64, limit int, onlineOnly bool) ([]*Chef, error)
	// SetOnline updates a chef's live presence flag.
	SetOnline(ctx context.Context, chefID int, online bool) error
	// SetAcceptingOrders updates the chef's availability (away / vacation mode).
	// When false the chef is hidden from browse/search and cannot take orders.
	SetAcceptingOrders(ctx context.Context, chefID int, accepting bool) error
	// SetImageURL updates the chef's kitchen photo URL.
	SetImageURL(ctx context.Context, chefID int, url string) error
	// Update persists the chef's editable profile fields (business name, bio,
	// specialty, kitchen address/city/coordinates, delivery radius) — never
	// verification, rating or status flags.
	Update(ctx context.Context, chef *Chef) error
}
