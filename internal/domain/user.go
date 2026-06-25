package domain

import "time"

// User is a platform account. The same entity backs customers, chefs and
// admins; the Role field distinguishes them. Fields mirror the users table
// (see migrations/000001_create_users_table.up.sql).
type User struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"` // never serialised
	PhoneNumber  *string `json:"phone_number,omitempty"`

	// Location
	Address   *string  `json:"address,omitempty"`
	City      *string  `json:"city,omitempty"`
	State     *string  `json:"state,omitempty"`
	ZipCode   *string  `json:"zip_code,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`

	// Role and status
	Role       string `json:"role"`
	IsVerified bool   `json:"is_verified"`
	IsActive   bool   `json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role constants.
const (
	RoleCustomer = "customer"
	RoleChef     = "chef"
	RoleAdmin    = "admin"
)

// ValidRole reports whether role is one the system recognises.
func ValidRole(role string) bool {
	switch role {
	case RoleCustomer, RoleChef, RoleAdmin:
		return true
	default:
		return false
	}
}

// NewUser builds a customer account with sensible defaults. The caller is
// responsible for hashing the password before passing it here.
func NewUser(username, email, passwordHash string) *User {
	now := time.Now()
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         RoleCustomer,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// IsCustomer reports whether the user is a customer.
func (u *User) IsCustomer() bool { return u.Role == RoleCustomer }

// IsChef reports whether the user is a chef.
func (u *User) IsChef() bool { return u.Role == RoleChef }

// IsAdmin reports whether the user is an admin.
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
