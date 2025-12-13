package domain

import "time"

// User represents a user in the system (customer, chef, or admin)
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose password hash
	PhoneNumber  *string   `json:"phone_number,omitempty"`
	
	// Location
	Address   *string  `json:"address,omitempty"`
	City      *string  `json:"city,omitempty"`
	State     *string  `json:"state,omitempty"`
	ZipCode   *string  `json:"zip_code,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	
	// Role and status
	Role       string `json:"role"` // customer, chef, admin
	IsVerified bool   `json:"is_verified"`
	IsActive   bool   `json:"is_active"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole constants
const (
	RoleCustomer = "customer"
	RoleChef     = "chef"
	RoleAdmin    = "admin"
)

// NewUser creates a new user with default values
func NewUser(username, email, passwordHash string) *User {
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleCustomer,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// IsCustomer checks if user is a customer
func (u *User) IsCustomer() bool {
	return u.Role == RoleCustomer
}

// IsChef checks if user is a chef
func (u *User) IsChef() bool {
	return u.Role == RoleChef
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// Activate activates the user account
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}

// Verify verifies the user email
func (u *User) Verify() {
	u.IsVerified = true
	u.UpdatedAt = time.Now()
}
