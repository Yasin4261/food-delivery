package domain

import (
	"context"
	"time"
)

// EarningsRepository is the port for the chef earnings read-model. It derives
// totals from order_items joined to delivered & paid orders.
type EarningsRepository interface {
	// ChefEarnings sums a chef's earnings from delivered & paid orders. When
	// since is non-nil, only orders created at or after it are counted.
	ChefEarnings(ctx context.Context, chefID int, since *time.Time) (*Earnings, error)
}
