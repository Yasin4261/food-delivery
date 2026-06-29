package domain

import "context"

// SearchRepository is the port for text search across the catalogue. Queries
// are case-insensitive substring matches; results are paginated.
type SearchRepository interface {
	// SearchChefs finds active chefs by business name, specialty or city.
	SearchChefs(ctx context.Context, q string, limit, offset int) ([]*Chef, error)
	// SearchMenuItems finds active & available dishes by name, description,
	// category or cuisine.
	SearchMenuItems(ctx context.Context, q string, limit, offset int) ([]*MenuItem, error)
	// SearchUsers finds active users by username or email (admin-only feature).
	SearchUsers(ctx context.Context, q string, limit, offset int) ([]*User, error)
}
