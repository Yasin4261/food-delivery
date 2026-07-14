package domain

import "context"

// Sort orders accepted by catalogue search/listing. Adapters must treat these
// as a whitelist — the value is mapped to a fixed ORDER BY expression, never
// interpolated from input. Every ordering ends with a stable id tiebreaker so
// pagination is deterministic.
const (
	SortDefault   = ""           // adapter's default ordering
	SortRating    = "rating"     // best rated first
	SortPopular   = "popular"    // most ordered first
	SortPriceAsc  = "price_asc"  // dishes only
	SortPriceDesc = "price_desc" // dishes only
)

// SearchFilters narrows catalogue search/listing results. Zero values mean
// "no constraint". Price and cuisine apply to dishes only.
type SearchFilters struct {
	MinRating float64
	MinPrice  float64
	MaxPrice  float64
	Cuisine   string
	Sort      string
}

// SearchRepository is the port for text search across the catalogue. Queries
// are case-insensitive substring matches; results are paginated.
type SearchRepository interface {
	// SearchChefs finds active chefs by business name, specialty or city,
	// narrowed and ordered by f (price/cuisine ignored).
	SearchChefs(ctx context.Context, q string, f SearchFilters, limit, offset int) ([]*Chef, int, error)
	// SearchMenuItems finds active & available dishes by name, description,
	// category or cuisine, narrowed and ordered by f.
	SearchMenuItems(ctx context.Context, q string, f SearchFilters, limit, offset int) ([]*MenuItem, int, error)
	// SearchUsers finds active users by username or email (admin-only feature).
	SearchUsers(ctx context.Context, q string, limit, offset int) ([]*User, int, error)
}
