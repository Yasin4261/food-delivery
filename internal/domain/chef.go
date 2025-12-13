package domain

import "time"

// Chef represents a chef/home cook profile
type Chef struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
	
	// Profile information
	BusinessName    string  `json:"business_name"`
	Bio             *string `json:"bio,omitempty"`
	Specialty       *string `json:"specialty,omitempty"`
	ExperienceYears *int    `json:"experience_years,omitempty"`
	
	// Location
	KitchenAddress   string   `json:"kitchen_address"`
	KitchenCity      *string  `json:"kitchen_city,omitempty"`
	KitchenLatitude  *float64 `json:"kitchen_latitude,omitempty"`
	KitchenLongitude *float64 `json:"kitchen_longitude,omitempty"`
	DeliveryRadius   int      `json:"delivery_radius"` // in kilometers
	
	// Certificates and verification
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
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewChef creates a new chef profile
func NewChef(userID int, businessName, kitchenAddress string) *Chef {
	return &Chef{
		UserID:            userID,
		BusinessName:      businessName,
		KitchenAddress:    kitchenAddress,
		DeliveryRadius:    5, // default 5km
		IsVerified:        false,
		IsActive:          true,
		IsAcceptingOrders: true,
		Rating:            0.0,
		TotalReviews:      0,
		TotalOrders:       0,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// Verify verifies the chef (admin action)
func (c *Chef) Verify() {
	c.IsVerified = true
	now := time.Now()
	c.VerifiedAt = &now
	c.UpdatedAt = now
}

// Activate activates the chef profile
func (c *Chef) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// Deactivate deactivates the chef profile
func (c *Chef) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// StartAcceptingOrders allows chef to accept new orders
func (c *Chef) StartAcceptingOrders() {
	c.IsAcceptingOrders = true
	c.UpdatedAt = time.Now()
}

// StopAcceptingOrders stops chef from accepting new orders
func (c *Chef) StopAcceptingOrders() {
	c.IsAcceptingOrders = false
	c.UpdatedAt = time.Now()
}

// UpdateRating updates chef rating
func (c *Chef) UpdateRating(newRating float64) {
	totalRating := c.Rating * float64(c.TotalReviews)
	c.TotalReviews++
	c.Rating = (totalRating + newRating) / float64(c.TotalReviews)
	c.UpdatedAt = time.Now()
}

// IncrementOrders increments total orders count
func (c *Chef) IncrementOrders() {
	c.TotalOrders++
	c.UpdatedAt = time.Now()
}

// CanDeliver checks if chef can deliver to given location
func (c *Chef) CanDeliver(lat, lng float64) bool {
	if c.KitchenLatitude == nil || c.KitchenLongitude == nil {
		return false
	}
	
	distance := CalculateDistance(*c.KitchenLatitude, *c.KitchenLongitude, lat, lng)
	return distance <= float64(c.DeliveryRadius)
}
