package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// EarningsRepository is the PostgreSQL adapter for domain.EarningsRepository.
type EarningsRepository struct {
	db *sql.DB
}

// NewEarningsRepository builds an EarningsRepository.
func NewEarningsRepository(db *sql.DB) *EarningsRepository {
	return &EarningsRepository{db: db}
}

// ChefEarnings sums a chef's line items from sub-orders they delivered on paid
// orders — the chef's own slice counts once *their* sub-order is delivered,
// regardless of other chefs in the same order. since (nullable) bounds the
// window by the order's created_at.
func (r *EarningsRepository) ChefEarnings(ctx context.Context, chefID int, since *time.Time) (*domain.Earnings, error) {
	const query = `
		SELECT
			COALESCE(SUM(oi.subtotal), 0)   AS total_earnings,
			COUNT(DISTINCT o.id)            AS delivered_orders,
			COALESCE(SUM(oi.quantity), 0)   AS items_sold
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN sub_orders s ON s.order_id = oi.order_id AND s.chef_id = oi.chef_id
		WHERE oi.chef_id = $1
		  AND s.status = 'delivered'
		  AND o.payment_status = 'paid'
		  AND ($2::timestamp IS NULL OR o.created_at >= $2)`

	e := &domain.Earnings{ChefID: chefID}
	err := r.db.QueryRowContext(ctx, query, chefID, since).
		Scan(&e.TotalEarnings, &e.DeliveredOrders, &e.ItemsSold)
	if err != nil {
		return nil, fmt.Errorf("chef earnings: %w", err)
	}
	return e, nil
}
