package domain

import "time"

// OrderItem is a single line in an order (mirrors the order_items table,
// migrations/000006_create_order_items_table.up.sql). It snapshots the dish's
// name and price at order time so later menu edits don't rewrite history, and
// carries chef_id so an order can span multiple chefs and be split per chef for
// chef-facing views and earnings.
type OrderItem struct {
	ID         int `json:"id"`
	OrderID    int `json:"order_id"`
	MenuItemID int `json:"menu_item_id"`
	ChefID     int `json:"chef_id"`

	ItemName  string  `json:"item_name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Subtotal  float64 `json:"subtotal"`

	SpecialInstructions *string `json:"special_instructions,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// NewOrderItem builds a line item from a menu item snapshot, computing the
// line subtotal as quantity * unitPrice.
func NewOrderItem(menuItemID, chefID int, name string, quantity int, unitPrice float64) *OrderItem {
	return &OrderItem{
		MenuItemID: menuItemID,
		ChefID:     chefID,
		ItemName:   name,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		Subtotal:   unitPrice * float64(quantity),
		CreatedAt:  time.Now(),
	}
}
