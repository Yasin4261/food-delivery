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

// SearchChefs finds active chefs by business name, specialty or city.
func (r *SearchRepository) SearchChefs(ctx context.Context, q string, limit, offset int) ([]*domain.Chef, error) {
	query := `SELECT` + chefColumns + `
		FROM chefs
		WHERE is_active = true
		  AND (business_name ILIKE $1 OR specialty ILIKE $1 OR kitchen_city ILIKE $1)
		ORDER BY rating DESC, created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, like(q), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search chefs: %w", err)
	}
	defer rows.Close()
	return collectChefs(rows)
}

// SearchMenuItems finds active & available dishes.
func (r *SearchRepository) SearchMenuItems(ctx context.Context, q string, limit, offset int) ([]*domain.MenuItem, error) {
	query := `SELECT` + menuItemColumns + `
		FROM menu_items
		WHERE is_active = true AND is_available = true
		  AND (name ILIKE $1 OR description ILIKE $1 OR category ILIKE $1 OR cuisine ILIKE $1)
		ORDER BY is_featured DESC, rating DESC, created_at DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, like(q), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search menu items: %w", err)
	}
	defer rows.Close()
	return collectMenuItems(rows)
}

// SearchUsers finds active users by username or email.
func (r *SearchRepository) SearchUsers(ctx context.Context, q string, limit, offset int) ([]*domain.User, error) {
	query := `SELECT` + userColumns + `
		FROM users
		WHERE is_active = true AND (username ILIKE $1 OR email ILIKE $1)
		ORDER BY id
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, like(q), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
