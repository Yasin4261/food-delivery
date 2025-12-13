package domain

import "time"

// Menu represents a chef's menu
type Menu struct {
	ID             int       `json:"id"`
	ChefID         int       `json:"chef_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	MenuType       string    `json:"menu_type"` // regular, daily_special, seasonal, weekend
	AvailableDays  *string   `json:"available_days,omitempty"`
	AvailableFrom  *string   `json:"available_from,omitempty"`  // TIME format
	AvailableUntil *string   `json:"available_until,omitempty"` // TIME format
	IsActive       bool      `json:"is_active"`
	IsFeatured     bool      `json:"is_featured"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// MenuType constants
const (
	MenuTypeRegular      = "regular"
	MenuTypeDailySpecial = "daily_special"
	MenuTypeSeasonal     = "seasonal"
	MenuTypeWeekend      = "weekend"
)

// NewMenu creates a new menu
func NewMenu(chefID int, name string) *Menu {
	return &Menu{
		ChefID:     chefID,
		Name:       name,
		MenuType:   MenuTypeRegular,
		IsActive:   true,
		IsFeatured: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Activate activates the menu
func (m *Menu) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// Deactivate deactivates the menu
func (m *Menu) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// Feature marks menu as featured
func (m *Menu) Feature() {
	m.IsFeatured = true
	m.UpdatedAt = time.Now()
}

// Unfeature removes featured status
func (m *Menu) Unfeature() {
	m.IsFeatured = false
	m.UpdatedAt = time.Now()
}
