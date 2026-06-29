package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ReviewRepository is the PostgreSQL adapter for domain.ReviewRepository.
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository builds a ReviewRepository.
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

const reviewColumns = `
	id, user_id, order_id, chef_id, menu_item_id, rating, comment, created_at, updated_at`

func scanReview(s interface{ Scan(...any) error }) (*domain.Review, error) {
	r := &domain.Review{}
	err := s.Scan(
		&r.ID, &r.UserID, &r.OrderID, &r.ChefID, &r.MenuItemID, &r.Rating, &r.Comment,
		&r.CreatedAt, &r.UpdatedAt,
	)
	return r, err
}

// Create inserts a review and recomputes the reviewed chef's or dish's
// aggregate rating in the same transaction, so the derived value never drifts
// from the review rows.
func (r *ReviewRepository) Create(ctx context.Context, rv *domain.Review) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	insert := `
		INSERT INTO reviews (user_id, order_id, chef_id, menu_item_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`
	err = tx.QueryRowContext(
		ctx, insert,
		rv.UserID, rv.OrderID, rv.ChefID, rv.MenuItemID, rv.Rating, rv.Comment,
	).Scan(&rv.ID, &rv.CreatedAt, &rv.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return domain.ErrReviewExists
		}
		return fmt.Errorf("create review: %w", err)
	}

	if err := recomputeRating(ctx, tx, rv); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit review: %w", err)
	}
	return nil
}

// recomputeRating updates the aggregate rating/total_reviews of whichever
// target the review points at, from the reviews table.
func recomputeRating(ctx context.Context, tx *sql.Tx, rv *domain.Review) error {
	switch {
	case rv.ChefID != nil:
		_, err := tx.ExecContext(ctx, `
			UPDATE chefs SET
				rating = COALESCE((SELECT ROUND(AVG(rating), 2) FROM reviews WHERE chef_id = $1), 0),
				total_reviews = (SELECT COUNT(*) FROM reviews WHERE chef_id = $1),
				updated_at = now()
			WHERE id = $1`, *rv.ChefID)
		if err != nil {
			return fmt.Errorf("recompute chef rating: %w", err)
		}
	case rv.MenuItemID != nil:
		_, err := tx.ExecContext(ctx, `
			UPDATE menu_items SET
				rating = COALESCE((SELECT ROUND(AVG(rating), 2) FROM reviews WHERE menu_item_id = $1), 0),
				total_reviews = (SELECT COUNT(*) FROM reviews WHERE menu_item_id = $1),
				updated_at = now()
			WHERE id = $1`, *rv.MenuItemID)
		if err != nil {
			return fmt.Errorf("recompute menu item rating: %w", err)
		}
	}
	return nil
}

// ListByChef returns a page of a chef's reviews (newest first) and the total.
func (r *ReviewRepository) ListByChef(ctx context.Context, chefID, limit, offset int) ([]*domain.Review, int, error) {
	reviews, err := r.list(ctx, `SELECT`+reviewColumns+`
		FROM reviews WHERE chef_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, chefID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.count(ctx, `SELECT count(*) FROM reviews WHERE chef_id = $1`, chefID)
	return reviews, total, err
}

// ListByMenuItem returns a page of a dish's reviews (newest first) and the total.
func (r *ReviewRepository) ListByMenuItem(ctx context.Context, menuItemID, limit, offset int) ([]*domain.Review, int, error) {
	reviews, err := r.list(ctx, `SELECT`+reviewColumns+`
		FROM reviews WHERE menu_item_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, menuItemID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.count(ctx, `SELECT count(*) FROM reviews WHERE menu_item_id = $1`, menuItemID)
	return reviews, total, err
}

func (r *ReviewRepository) count(ctx context.Context, query string, arg any) (int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, query, arg).Scan(&total); err != nil {
		return 0, fmt.Errorf("count reviews: %w", err)
	}
	return total, nil
}

func (r *ReviewRepository) list(ctx context.Context, query string, args ...any) ([]*domain.Review, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]*domain.Review, 0)
	for rows.Next() {
		rv, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, rv)
	}
	return reviews, rows.Err()
}
