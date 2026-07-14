package domain

import "time"

// Chef is a freelance home-cook profile. It belongs 1:1 to a User (the account)
// and mirrors the chefs table (migrations/000003_create_chefs_table.up.sql).
type Chef struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`

	// Profile
	BusinessName    string  `json:"business_name"`
	Bio             *string `json:"bio,omitempty"`
	Specialty       *string `json:"specialty,omitempty"`
	ExperienceYears *int    `json:"experience_years,omitempty"`
	ImageURL        *string `json:"image_url,omitempty"` // kitchen photo

	// Location
	KitchenAddress   string   `json:"kitchen_address"`
	KitchenCity      *string  `json:"kitchen_city,omitempty"`
	KitchenLatitude  *float64 `json:"kitchen_latitude,omitempty"`
	KitchenLongitude *float64 `json:"kitchen_longitude,omitempty"`
	DeliveryRadius   int      `json:"delivery_radius"` // kilometres

	// Verification
	FoodLicenseNumber    *string    `json:"food_license_number,omitempty"`
	HealthCertificateURL *string    `json:"health_certificate_url,omitempty"`
	IsVerified           bool       `json:"is_verified"`
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`

	// Statistics
	Rating       float64 `json:"rating"`
	TotalReviews int     `json:"total_reviews"`
	TotalOrders  int     `json:"total_orders"`

	// Status
	IsActive          bool `json:"is_active"`
	IsAcceptingOrders bool `json:"is_accepting_orders"`
	IsOnline          bool `json:"is_online"` // live presence, distinct from accepting orders

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

const defaultDeliveryRadiusKm = 5

// NewChef builds a chef profile with sensible defaults.
func NewChef(userID int, businessName, kitchenAddress string) *Chef {
	now := time.Now()
	return &Chef{
		UserID:            userID,
		BusinessName:      businessName,
		KitchenAddress:    kitchenAddress,
		DeliveryRadius:    defaultDeliveryRadiusKm,
		IsActive:          true,
		IsAcceptingOrders: true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// HasLocation reports whether the kitchen has coordinates set.
func (c *Chef) HasLocation() bool {
	return c.KitchenLatitude != nil && c.KitchenLongitude != nil
}

// CanDeliverTo reports whether (lat, lng) falls within the chef's delivery
// radius. A chef without coordinates can deliver nowhere.
func (c *Chef) CanDeliverTo(lat, lng float64) bool {
	if !c.HasLocation() {
		return false
	}
	return CalculateDistance(*c.KitchenLatitude, *c.KitchenLongitude, lat, lng) <= float64(c.DeliveryRadius)
}

// SetAcceptingOrders toggles whether the chef takes new orders.
func (c *Chef) SetAcceptingOrders(accepting bool) {
	c.IsAcceptingOrders = accepting
	c.UpdatedAt = time.Now()
}

// SetOnline toggles the chef's live presence.
func (c *Chef) SetOnline(online bool) {
	c.IsOnline = online
	c.UpdatedAt = time.Now()
}
