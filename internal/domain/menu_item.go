package domain

import "time"

// MenuItem represents a dish/meal in a menu
type MenuItem struct {
	ID     int `json:"id"`
	MenuID int `json:"menu_id"`
	ChefID int `json:"chef_id"`
	
	// Meal information
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"` // appetizer, main_course, dessert, beverage, soup
	Cuisine     *string `json:"cuisine,omitempty"`  // turkish, italian, chinese, etc.
	
	// Pricing
	Price         float64  `json:"price"`
	OriginalPrice *float64 `json:"original_price,omitempty"`
	
	// Portion and preparation
	PortionSize     *string `json:"portion_size,omitempty"`
	PreparationTime *int    `json:"preparation_time,omitempty"` // in minutes
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
	
	// Nutritional values
	Calories *int     `json:"calories,omitempty"`
	Protein  *float64 `json:"protein,omitempty"`
	Carbs    *float64 `json:"carbs,omitempty"`
	Fat      *float64 `json:"fat,omitempty"`
	
	// Media
	ImageURL *string `json:"image_url,omitempty"`
	Images   *string `json:"images,omitempty"` // JSON array
	
	// Statistics
	Rating       float64 `json:"rating"`
	TotalReviews int     `json:"total_reviews"`
	TotalOrders  int     `json:"total_orders"`
	Views        int     `json:"views"`
	
	// Status
	IsActive    bool `json:"is_active"`
	IsFeatured  bool `json:"is_featured"`
	IsAvailable bool `json:"is_available"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Category constants
const (
	CategoryAppetizer  = "appetizer"
	CategoryMainCourse = "main_course"
	CategoryDessert    = "dessert"
	CategoryBeverage   = "beverage"
	CategorySoup       = "soup"
)

// NewMenuItem creates a new menu item
func NewMenuItem(menuID, chefID int, name string, price float64) *MenuItem {
	return &MenuItem{
		MenuID:       menuID,
		ChefID:       chefID,
		Name:         name,
		Price:        price,
		ServingSize:  1,
		IsUnlimited:  false,
		IsVegetarian: false,
		IsVegan:      false,
		IsGlutenFree: false,
		IsHalal:      false,
		IsSpicy:      false,
		Rating:       0.0,
		IsActive:     true,
		IsFeatured:   false,
		IsAvailable:  true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// DecrementStock decrements available quantity
func (m *MenuItem) DecrementStock(quantity int) error {
	if m.IsUnlimited {
		return nil
	}
	
	if m.AvailableQuantity == nil {
		return ErrStockNotManaged
	}
	
	if *m.AvailableQuantity < quantity {
		return ErrInsufficientStock
	}
	
	*m.AvailableQuantity -= quantity
	m.UpdatedAt = time.Now()
	return nil
}

// IncrementStock increments available quantity
func (m *MenuItem) IncrementStock(quantity int) {
	if m.IsUnlimited {
		return
	}
	
	if m.AvailableQuantity == nil {
		qty := quantity
		m.AvailableQuantity = &qty
	} else {
		*m.AvailableQuantity += quantity
	}
	m.UpdatedAt = time.Now()
}

// UpdateRating updates menu item rating
func (m *MenuItem) UpdateRating(newRating float64) {
	totalRating := m.Rating * float64(m.TotalReviews)
	m.TotalReviews++
	m.Rating = (totalRating + newRating) / float64(m.TotalReviews)
	m.UpdatedAt = time.Now()
}

// IncrementViews increments view count
func (m *MenuItem) IncrementViews() {
	m.Views++
	m.UpdatedAt = time.Now()
}

// IncrementOrders increments order count
func (m *MenuItem) IncrementOrders() {
	m.TotalOrders++
	m.UpdatedAt = time.Now()
}

// Activate activates the menu item
func (m *MenuItem) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate deactivates the menu item
func (m *MenuItem) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// MarkAvailable marks item as available
func (m *MenuItem) MarkAvailable() {
	m.IsAvailable = true
	m.UpdatedAt = time.Now()
}

// MarkUnavailable marks item as unavailable
func (m *MenuItem) MarkUnavailable() {
	m.IsAvailable = false
	m.UpdatedAt = time.Now()
}

// Feature marks item as featured
func (m *MenuItem) Feature() {
	m.IsFeatured = true
	m.UpdatedAt = time.Now()
}

// Unfeature removes featured status
func (m *MenuItem) Unfeature() {
	m.IsFeatured = false
	m.UpdatedAt = time.Now()
}
