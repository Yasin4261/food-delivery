package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChefHoursRepository is the PostgreSQL adapter for
// domain.ChefHoursRepository.
type ChefHoursRepository struct {
	db *sql.DB
}

// NewChefHoursRepository builds a ChefHoursRepository.
func NewChefHoursRepository(db *sql.DB) *ChefHoursRepository {
	return &ChefHoursRepository{db: db}
}

// ReplaceForChef swaps the chef's whole weekly schedule in one transaction.
func (r *ChefHoursRepository) ReplaceForChef(ctx context.Context, chefID int, hours []*domain.ChefHours) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chef_hours WHERE chef_id = $1`, chefID); err != nil {
		return fmt.Errorf("clear chef hours: %w", err)
	}
	for _, h := range hours {
		err := tx.QueryRowContext(ctx, `
			INSERT INTO chef_hours (chef_id, weekday, opens_at, closes_at)
			VALUES ($1, $2, $3, $4)
			RETURNING id`, chefID, h.Weekday, h.OpensAt, h.ClosesAt).Scan(&h.ID)
		if err != nil {
			return fmt.Errorf("insert chef hours: %w", err)
		}
		h.ChefID = chefID
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit chef hours: %w", err)
	}
	return nil
}

// ListByChef returns a chef's windows ordered by weekday, opens_at.
func (r *ChefHoursRepository) ListByChef(ctx context.Context, chefID int) ([]*domain.ChefHours, error) {
	byChef, err := r.ListByChefs(ctx, []int{chefID})
	if err != nil {
		return nil, err
	}
	if hours, ok := byChef[chefID]; ok {
		return hours, nil
	}
	return []*domain.ChefHours{}, nil
}

// ListByChefs returns the windows of several chefs in one query, grouped by
// chef id.
func (r *ChefHoursRepository) ListByChefs(ctx context.Context, chefIDs []int) (map[int][]*domain.ChefHours, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, chef_id, weekday, opens_at, closes_at
		FROM chef_hours
		WHERE chef_id = ANY($1)
		ORDER BY chef_id, weekday, opens_at`, pq.Array(chefIDs))
	if err != nil {
		return nil, fmt.Errorf("list chef hours: %w", err)
	}
	defer rows.Close()

	out := make(map[int][]*domain.ChefHours)
	for rows.Next() {
		h := &domain.ChefHours{}
		if err := rows.Scan(&h.ID, &h.ChefID, &h.Weekday, &h.OpensAt, &h.ClosesAt); err != nil {
			return nil, fmt.Errorf("scan chef hours: %w", err)
		}
		out[h.ChefID] = append(out[h.ChefID], h)
	}
	return out, rows.Err()
}
