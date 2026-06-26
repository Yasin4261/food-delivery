package domain

import "time"

// Menu is a chef's collection of dishes (mirrors the menus table,
// migrations/000004_create_menus_table.up.sql). A menu belongs to exactly one
// chef and groups MenuItems.
type Menu struct {
	ID     int `json:"id"`
	ChefID int `json:"chef_id"`

	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	MenuType    string  `json:"menu_type"`

	// AvailableDays is a comma-separated list, e.g. "monday,tuesday,friday".
	AvailableDays *string `json:"available_days,omitempty"`

	IsActive   bool `json:"is_active"`
	IsFeatured bool `json:"is_featured"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Menu types the product recognises.
const (
	MenuTypeRegular      = "regular"
	MenuTypeDailySpecial = "daily_special"
	MenuTypeSeasonal     = "seasonal"
	MenuTypeWeekend      = "weekend"

	defaultMenuType = MenuTypeRegular
)

// ValidMenuType reports whether t is a recognised menu type.
func ValidMenuType(t string) bool {
	switch t {
	case MenuTypeRegular, MenuTypeDailySpecial, MenuTypeSeasonal, MenuTypeWeekend:
		return true
	default:
		return false
	}
}

// NewMenu builds a menu for a chef with sensible defaults.
func NewMenu(chefID int, name string) *Menu {
	now := time.Now()
	return &Menu{
		ChefID:    chefID,
		Name:      name,
		MenuType:  defaultMenuType,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
