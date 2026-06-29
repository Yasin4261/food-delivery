package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/middleware"
)

func TestRateLimiter_ThrottlesPerIP(t *testing.T) {
	rl := middleware.NewRateLimiter(2, time.Minute)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	call := func(ip string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/v2/auth/login", nil)
		req.RemoteAddr = ip + ":12345"
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	// First two from the same IP pass, the third is throttled.
	if call("10.0.0.1") != http.StatusOK || call("10.0.0.1") != http.StatusOK {
		t.Fatal("first two requests should pass")
	}
	if code := call("10.0.0.1"); code != http.StatusTooManyRequests {
		t.Errorf("third request = %d, want 429", code)
	}

	// A different IP has its own budget.
	if code := call("10.0.0.2"); code != http.StatusOK {
		t.Errorf("other IP = %d, want 200", code)
	}
}

func TestRateLimiter_AllowReportsRetryAfter(t *testing.T) {
	rl := middleware.NewRateLimiter(1, time.Minute)
	if ok, _ := rl.Allow("k"); !ok {
		t.Fatal("first allow should pass")
	}
	ok, retry := rl.Allow("k")
	if ok {
		t.Error("second allow should be blocked")
	}
	if retry <= 0 {
		t.Errorf("retry-after = %d, want > 0", retry)
	}
}
