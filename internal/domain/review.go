package domain

import "time"

// Review is a customer's rating of either a chef or a dish, tied to the order
// it came from (mirrors the reviews table,
// migrations/000008_create_reviews_table.up.sql). Exactly one of ChefID /
// MenuItemID is set. Aggregate ratings on chefs / menu_items are recomputed
// from these rows.
type Review struct {
	ID      int `json:"id"`
	UserID  int `json:"user_id"`
	OrderID int `json:"order_id"`

	ChefID     *int `json:"chef_id,omitempty"`
	MenuItemID *int `json:"menu_item_id,omitempty"`

	Rating  int     `json:"rating"`
	Comment *string `json:"comment,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Rating bounds.
const (
	MinRating = 1
	MaxRating = 5
)

// TargetsChef reports whether the review targets a chef.
func (r *Review) TargetsChef() bool { return r.ChefID != nil }

// TargetsMenuItem reports whether the review targets a dish.
func (r *Review) TargetsMenuItem() bool { return r.MenuItemID != nil }

// Validate checks the rating range and that exactly one target is set.
func (r *Review) Validate() error {
	if r.Rating < MinRating || r.Rating > MaxRating {
		return ErrInvalidRating
	}
	if r.TargetsChef() == r.TargetsMenuItem() {
		// both set or neither set
		return ErrInvalidReviewTarget
	}
	return nil
}
