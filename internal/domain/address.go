package domain

import (
	"strings"
	"time"
)

// Address is one saved delivery address in a customer's address book (mirrors
// the addresses table, migrations/000018). Orders never reference an address
// row — placement snapshots the text onto the order — so editing or deleting
// an address never rewrites history.
type Address struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`

	Label     string   `json:"label"`
	Address   string   `json:"address"`
	City      *string  `json:"city,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
	IsDefault bool     `json:"is_default"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Validate checks the invariants shared by create and update.
func (a *Address) Validate() error {
	a.Label = strings.TrimSpace(a.Label)
	a.Address = strings.TrimSpace(a.Address)
	if a.Label == "" {
		return ErrAddressLabelRequired
	}
	if len(a.Label) > 50 {
		return ErrAddressLabelTooLong
	}
	if a.Address == "" {
		return ErrAddressRequired
	}
	if (a.Latitude == nil) != (a.Longitude == nil) {
		return ErrCoordinatesIncomplete
	}
	return nil
}
