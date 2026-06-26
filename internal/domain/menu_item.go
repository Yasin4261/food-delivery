package domain

import "time"

// MenuItem is an individual dish within a menu (mirrors the menu_items table,
// migrations/000005_create_menu_items_table.up.sql). It carries chef_id as well
// as menu_id so order lines and chef-facing views can be filtered by chef
// without joining through menus.
type MenuItem struct {
	ID     int `json:"id"`
	MenuID int `json:"menu_id"`
	ChefID int `json:"chef_id"`

	// Description
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	Cuisine     *string `json:"cuisine,omitempty"`

	// Pricing
	Price         float64  `json:"price"`
	OriginalPrice *float64 `json:"original_price,omitempty"`

	// Portion and preparation
	PortionSize     *string `json:"portion_size,omitempty"`
	PreparationTime *int    `json:"preparation_time,omitempty"` // minutes
	ServingSize     int     `json:"serving_size"`

	// Stock
	AvailableQuantity *int `json:"available_quantity,omitempty"`
	IsUnlimited       bool `json:"is_unlimited"`
	DailyLimit        *int `json:"daily_limit,omitempty"`

	// Dietary features
	IsVegetarian bool `json:"is_vegetarian"`
	IsVegan      bool `json:"is_vegan"`
	IsGlutenFree bool `json:"is_gluten_free"`
	IsHalal      bool `json:"is_halal"`
	IsSpicy      bool `json:"is_spicy"`
	SpiceLevel   *int `json:"spice_level,omitempty"` // 0-5

	// Nutrition
	Calories *int     `json:"calories,omitempty"`
	Protein  *float64 `json:"protein,omitempty"`
	Carbs    *float64 `json:"carbs,omitempty"`
	Fat      *float64 `json:"fat,omitempty"`

	// Media
	ImageURL *string `json:"image_url,omitempty"`
	Images   *string `json:"images,omitempty"` // JSON array of URLs

	// Statistics
	Rating       float64 `json:"rating"`
	TotalReviews int     `json:"total_reviews"`
	TotalOrders  int     `json:"total_orders"`
	Views        int     `json:"views"`

	// Status
	IsActive    bool `json:"is_active"`
	IsFeatured  bool `json:"is_featured"`
	IsAvailable bool `json:"is_available"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMenuItem builds a dish with sensible defaults.
func NewMenuItem(menuID, chefID int, name string, price float64) *MenuItem {
	now := time.Now()
	return &MenuItem{
		MenuID:      menuID,
		ChefID:      chefID,
		Name:        name,
		Price:       price,
		ServingSize: 1,
		IsActive:    true,
		IsAvailable: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// InStock reports whether at least qty units can be ordered. An unlimited item
// is always in stock; a limited item without a tracked quantity is treated as
// out of stock.
func (m *MenuItem) InStock(qty int) bool {
	if qty <= 0 {
		return false
	}
	if m.IsUnlimited {
		return true
	}
	return m.AvailableQuantity != nil && *m.AvailableQuantity >= qty
}

// IsOrderable reports whether a customer can currently order this item.
func (m *MenuItem) IsOrderable() bool {
	return m.IsActive && m.IsAvailable && m.InStock(1)
}
