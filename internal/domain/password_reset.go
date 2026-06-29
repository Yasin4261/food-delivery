package domain

import "time"

// PasswordResetToken is a single-use, expiring token for resetting a forgotten
// password (mirrors password_reset_tokens). Only the hash of the raw token is
// stored; the raw token is delivered to the user out of band.
type PasswordResetToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Usable reports whether the token can still be redeemed at time now: it must
// be unused and not yet expired.
func (t *PasswordResetToken) Usable(now time.Time) bool {
	return t.UsedAt == nil && now.Before(t.ExpiresAt)
}
