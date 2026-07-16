package domain

import "time"

// EmailVerificationToken is a single-use, expiring token proving ownership of an
// account's email address (mirrors email_verification_tokens). Only the hash of
// the raw token is stored; the raw token is delivered to the user out of band
// and redeeming it flips users.is_verified.
type EmailVerificationToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Usable reports whether the token can still be redeemed at time now: it must be
// unused and not yet expired.
func (t *EmailVerificationToken) Usable(now time.Time) bool {
	return t.UsedAt == nil && now.Before(t.ExpiresAt)
}
