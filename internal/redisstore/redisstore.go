// Package redisstore holds Redis-backed adapters for the cross-cutting
// contracts that are in-memory by default: the token denylist
// (service.TokenRevoker / middleware.Denylist) and the auth rate limiter
// (middleware.Limiter). With Redis behind them, logout revocation and rate
// limits are shared across API instances.
//
// Availability trade-off: on Redis errors both adapters FAIL OPEN (the token
// is treated as not revoked; the request is allowed) with a logged warning —
// an outage must not lock every user out. Harden to fail-closed if the threat
// model demands it.
package redisstore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// opTimeout bounds every Redis call so a slow Redis cannot stall requests.
const opTimeout = 500 * time.Millisecond

// Denylist is a Redis-backed set of revoked JWT ids. Entries expire together
// with the token, so the set stays small.
type Denylist struct {
	rdb *redis.Client
}

// NewDenylist builds a Redis-backed denylist.
func NewDenylist(rdb *redis.Client) *Denylist {
	return &Denylist{rdb: rdb}
}

func denyKey(jti string) string { return "denylist:" + jti }

// Revoke marks a token id as revoked until its expiry.
func (d *Denylist) Revoke(jti string, exp time.Time) {
	if jti == "" {
		return
	}
	ttl := time.Until(exp)
	if ttl <= 0 {
		return // already expired; nothing to deny
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	if err := d.rdb.Set(ctx, denyKey(jti), "1", ttl).Err(); err != nil {
		slog.Warn("redis denylist: revoke failed", "error", err)
	}
}

// IsRevoked reports whether a token id is currently revoked.
func (d *Denylist) IsRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	n, err := d.rdb.Exists(ctx, denyKey(jti)).Result()
	if err != nil {
		slog.Warn("redis denylist: lookup failed; failing open", "error", err)
		return false
	}
	return n > 0
}

// RateLimiter is a Redis-backed fixed-window, per-key limiter (INCR + EXPIRE),
// so the window is shared across instances.
type RateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
}

// NewRateLimiter builds a limiter allowing limit hits per key per window.
func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, limit: limit, window: window}
}

// Allow records a hit for key and reports whether it is within the limit,
// returning the seconds until the window resets when it is not.
func (r *RateLimiter) Allow(key string) (bool, int) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	k := fmt.Sprintf("ratelimit:%s", key)
	n, err := r.rdb.Incr(ctx, k).Result()
	if err != nil {
		slog.Warn("redis rate limiter: incr failed; failing open", "error", err)
		return true, 0
	}
	if n == 1 {
		// First hit opens the window.
		if err := r.rdb.Expire(ctx, k, r.window).Err(); err != nil {
			slog.Warn("redis rate limiter: expire failed", "error", err)
		}
	}
	if int(n) > r.limit {
		ttl, err := r.rdb.TTL(ctx, k).Result()
		if err != nil || ttl < 0 {
			return false, int(r.window.Seconds())
		}
		return false, int(ttl.Seconds()) + 1
	}
	return true, 0
}
