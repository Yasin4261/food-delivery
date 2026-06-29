package domain

import "time"

// Favorite is a customer marking a chef as a favorite (mirrors the favorites
// table, migrations/000007_create_favorites_table.up.sql). The pair
// (UserID, ChefID) is unique, so favoriting is idempotent.
type Favorite struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ChefID    int       `json:"chef_id"`
	CreatedAt time.Time `json:"created_at"`
}
