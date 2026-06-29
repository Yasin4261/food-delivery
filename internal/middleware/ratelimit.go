package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter is a simple fixed-window, per-key limiter. It is intended to
// throttle abusive bursts (e.g. credential stuffing on auth endpoints), not to
// be a precise quota system.
type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*windowCount
	now    func() time.Time // injectable for tests
}

type windowCount struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter builds a limiter allowing limit requests per key per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:  limit,
		window: window,
		hits:   map[string]*windowCount{},
		now:    time.Now,
	}
}

// Allow records a hit for key and reports whether it is within the limit,
// returning the seconds until the window resets.
func (rl *RateLimiter) Allow(key string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	w := rl.hits[key]
	if w == nil || now.After(w.resetAt) {
		rl.hits[key] = &windowCount{count: 1, resetAt: now.Add(rl.window)}
		return true, 0
	}
	w.count++
	if w.count > rl.limit {
		return false, int(time.Until(w.resetAt).Seconds()) + 1
	}
	return true, 0
}

// Middleware throttles requests per client IP, replying 429 when the limit is
// exceeded.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, retryAfter := rl.Allow(clientIP(r))
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "too many requests; slow down"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
