package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Limiter is the throttling contract: Allow records a hit for key and reports
// whether it is within budget (and, when it is not, the seconds until the
// window resets). Implementations: the in-memory RateLimiter below and
// redisstore.RateLimiter (shared across instances).
type Limiter interface {
	Allow(key string) (ok bool, retryAfterSeconds int)
}

// RateLimit wraps a handler with per-client-IP throttling over any Limiter,
// replying 429 with a Retry-After header when the limit is exceeded.
func RateLimit(l Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ok, retryAfter := l.Allow(clientIP(r))
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
}

// RateLimiter is a simple in-memory fixed-window, per-key limiter. It is
// intended to throttle abusive bursts (e.g. credential stuffing on auth
// endpoints), not to be a precise quota system. Process-local; multi-instance
// deployments set REDIS_URL for the shared implementation.
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
// exceeded. It is RateLimit over this limiter.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return RateLimit(rl)(next)
}
