package domain

import "context"

// AddressRepository is the port for address-book persistence. Lookups return
// ErrAddressNotFound when no row matches.
type AddressRepository interface {
	// Create persists an address. When the address is marked default, any
	// previous default of the same user is cleared in the same transaction.
	Create(ctx context.Context, address *Address) error
	// FindByID returns one address.
	FindByID(ctx context.Context, id int) (*Address, error)
	// ListByUser returns a user's addresses, default first, then newest.
	ListByUser(ctx context.Context, userID int) ([]*Address, error)
	// Update persists the editable fields. When the address becomes default,
	// any previous default of the same user is cleared in the same
	// transaction.
	Update(ctx context.Context, address *Address) error
	// Delete removes an address.
	Delete(ctx context.Context, id int) error
}
