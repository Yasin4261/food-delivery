package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// FavoriteRepository is the PostgreSQL adapter for domain.FavoriteRepository.
type FavoriteRepository struct {
	db *sql.DB
}

// NewFavoriteRepository builds a FavoriteRepository.
func NewFavoriteRepository(db *sql.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Add favorites a chef, relying on the unique(user_id, chef_id) constraint to
// stay idempotent.
func (r *FavoriteRepository) Add(ctx context.Context, userID, chefID int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO favorites (user_id, chef_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, chef_id) DO NOTHING`, userID, chefID)
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}

// Remove unfavorites a chef.
func (r *FavoriteRepository) Remove(ctx context.Context, userID, chefID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM favorites WHERE user_id = $1 AND chef_id = $2`, userID, chefID)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

// ListChefs returns the active chefs a user has favorited, most recently
// favorited first. It reuses the shared chef column list and scanner; keeping
// favorites in correlated subqueries (not a join) leaves chefs as the only
// table in the main scope, so the unqualified chef columns stay unambiguous.
func (r *FavoriteRepository) ListChefs(ctx context.Context, userID, limit, offset int) ([]*domain.Chef, int, error) {
	query := `SELECT` + chefColumns + `
		FROM chefs
		WHERE chefs.is_active = true
		  AND chefs.id IN (SELECT chef_id FROM favorites WHERE user_id = $1)
		ORDER BY (SELECT created_at FROM favorites WHERE favorites.chef_id = chefs.id AND favorites.user_id = $1) DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	chefs, err := collectChefs(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM favorites f JOIN chefs c ON c.id = f.chef_id
		WHERE f.user_id = $1 AND c.is_active = true`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}
	return chefs, total, nil
}
