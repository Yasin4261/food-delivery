package domain

import "context"

// MenuRepository is the port for menu persistence. Lookups return
// ErrMenuNotFound when no row matches.
type MenuRepository interface {
	Create(ctx context.Context, menu *Menu) error
	FindByID(ctx context.Context, id int) (*Menu, error)
	// ListByChef returns a page of a chef's active menus plus the total.
	ListByChef(ctx context.Context, chefID, limit, offset int) ([]*Menu, int, error)
	Update(ctx context.Context, menu *Menu) error
	// Deactivate soft-deletes a menu (is_active = false).
	Deactivate(ctx context.Context, id int) error
}

// MenuItemRepository is the port for dish persistence. Lookups return
// ErrMenuItemNotFound when no row matches.
type MenuItemRepository interface {
	Create(ctx context.Context, item *MenuItem) error
	FindByID(ctx context.Context, id int) (*MenuItem, error)
	// ListByMenu returns the active items in a menu.
	ListByMenu(ctx context.Context, menuID int) ([]*MenuItem, error)
	// ListByChef returns a page of a chef's active items across all menus plus the total.
	ListByChef(ctx context.Context, chefID, limit, offset int) ([]*MenuItem, int, error)
	Update(ctx context.Context, item *MenuItem) error
	// Deactivate soft-deletes an item (is_active = false).
	Deactivate(ctx context.Context, id int) error
	// DecrementStock atomically reduces a limited item's available_quantity by
	// qty, failing with ErrItemOutOfStock if not enough stock remains. It is a
	// no-op error (ErrItemOutOfStock) for unlimited items, which the caller
	// should skip.
	DecrementStock(ctx context.Context, id, qty int) error
	// SetImageURL updates the dish photo URL (the cover image).
	SetImageURL(ctx context.Context, id int, url string) error
	// SetImages persists the dish gallery (a JSON array of URLs, or nil to
	// clear it).
	SetImages(ctx context.Context, id int, images *string) error
}
