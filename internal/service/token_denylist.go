package service

import (
	"sync"
	"time"
)

// TokenRevoker is the revocation contract shared by the logout handler
// (Revoke) and the auth middleware (IsRevoked). Implementations: the in-memory
// TokenDenylist below (single instance) and redisstore.Denylist (shared across
// instances).
type TokenRevoker interface {
	Revoke(jti string, exp time.Time)
	IsRevoked(jti string) bool
}

// TokenDenylist is an in-memory store of revoked JWT ids (jti). A logged-out or
// otherwise revoked token is rejected by the auth middleware until it would
// have expired anyway, so entries are kept only until their expiry.
//
// This is process-local; multi-instance deployments set REDIS_URL to use the
// Redis-backed implementation instead.
type TokenDenylist struct {
	mu      sync.Mutex
	revoked map[string]time.Time // jti -> expiry
	now     func() time.Time
}

// NewTokenDenylist builds an empty denylist.
func NewTokenDenylist() *TokenDenylist {
	return &TokenDenylist{revoked: map[string]time.Time{}, now: time.Now}
}

// Revoke marks a token id as revoked until exp.
func (d *TokenDenylist) Revoke(jti string, exp time.Time) {
	if jti == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.revoked[jti] = exp
}

// IsRevoked reports whether a token id is currently revoked, purging it once it
// has expired.
func (d *TokenDenylist) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	exp, ok := d.revoked[jti]
	if !ok {
		return false
	}
	if d.now().After(exp) {
		delete(d.revoked, jti)
		return false
	}
	return true
}
