package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// SearchRepository is the PostgreSQL adapter for domain.SearchRepository. It
// reuses the column lists and scanners of the other adapters.
type SearchRepository struct {
	db *sql.DB
}

// NewSearchRepository builds a SearchRepository.
func NewSearchRepository(db *sql.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// like wraps a query term for a case-insensitive substring match.
func like(q string) string { return "%" + q + "%" }

// Sort whitelists: the client's sort value selects one of these fixed ORDER BY
// expressions — it is never interpolated. Every ordering ends with the id so
// pagination is stable when the sort key ties.
var chefOrder = map[string]string{
	domain.SortDefault: `rating DESC, created_at DESC, id DESC`,
	domain.SortRating:  `rating DESC, total_reviews DESC, id DESC`,
	domain.SortPopular: `total_orders DESC, rating DESC, id DESC`,
}

var dishOrder = map[string]string{
	domain.SortDefault:   `is_featured DESC, rating DESC, created_at DESC, id DESC`,
	domain.SortRating:    `rating DESC, total_reviews DESC, id DESC`,
	domain.SortPopular:   `total_orders DESC, rating DESC, id DESC`,
	domain.SortPriceAsc:  `price ASC, id DESC`,
	domain.SortPriceDesc: `price DESC, id DESC`,
}

// SearchChefs finds active chefs by business name, specialty or city,
// narrowed by min rating and ordered by the whitelisted sort.
func (r *SearchRepository) SearchChefs(ctx context.Context, q string, f domain.SearchFilters, limit, offset int) ([]*domain.Chef, int, error) {
	where := ` WHERE is_active = true AND is_accepting_orders = true
		  AND (business_name ILIKE $1 OR specialty ILIKE $1 OR kitchen_city ILIKE $1)
		  AND rating >= $2`
	args := []any{like(q), f.MinRating}

	rows, err := r.db.QueryContext(ctx,
		`SELECT`+chefColumns+` FROM chefs`+where+`
		ORDER BY `+chefOrder[f.Sort]+` LIMIT $3 OFFSET $4`,
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("search chefs: %w", err)
	}
	defer rows.Close()
	chefs, err := collectChefs(rows)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.count(ctx, `SELECT count(*) FROM chefs`+where, args...)
	return chefs, total, err
}

// SearchMenuItems finds active & available dishes, narrowed by rating, price
// range and cuisine, ordered by the whitelisted sort.
func (r *SearchRepository) SearchMenuItems(ctx context.Context, q string, f domain.SearchFilters, limit, offset int) ([]*domain.MenuItem, int, error) {
	// Each dietary flag is a bound bool: when true it requires the attribute;
	// when false it adds no constraint. Cheap and index-friendly.
	where := ` WHERE is_active = true AND is_available = true
		  AND (name ILIKE $1 OR description ILIKE $1 OR category ILIKE $1 OR cuisine ILIKE $1)
		  AND rating >= $2
		  AND price >= $3
		  AND ($4 = 0 OR price <= $4)
		  AND ($5 = '' OR cuisine ILIKE $5)
		  AND (NOT $6 OR is_vegetarian)
		  AND (NOT $7 OR is_vegan)
		  AND (NOT $8 OR is_gluten_free)
		  AND (NOT $9 OR is_halal)`
	cuisineArg := ""
	if f.Cuisine != "" {
		cuisineArg = like(f.Cuisine)
	}
	args := []any{like(q), f.MinRating, f.MinPrice, f.MaxPrice, cuisineArg,
		f.Vegetarian, f.Vegan, f.GlutenFree, f.Halal}

	rows, err := r.db.QueryContext(ctx,
		`SELECT`+menuItemColumns+` FROM menu_items`+where+`
		ORDER BY `+dishOrder[f.Sort]+` LIMIT $10 OFFSET $11`,
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("search menu items: %w", err)
	}
	defer rows.Close()
	items, err := collectMenuItems(rows)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.count(ctx, `SELECT count(*) FROM menu_items`+where, args...)
	return items, total, err
}

// SearchUsers finds active users by username or email.
func (r *SearchRepository) SearchUsers(ctx context.Context, q string, limit, offset int) ([]*domain.User, int, error) {
	const where = ` WHERE is_active = true AND (username ILIKE $1 OR email ILIKE $1)`
	rows, err := r.db.QueryContext(ctx, `SELECT`+userColumns+` FROM users`+where+`
		ORDER BY id LIMIT $2 OFFSET $3`, like(q), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	total, err := r.count(ctx, `SELECT count(*) FROM users`+where, like(q))
	return users, total, err
}

func (r *SearchRepository) count(ctx context.Context, query string, args ...any) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count search: %w", err)
	}
	return total, nil
}
