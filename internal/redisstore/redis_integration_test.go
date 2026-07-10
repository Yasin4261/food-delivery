//go:build integration

package redisstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// client connects to the throwaway Redis from docker-compose.test.yml via
// TEST_REDIS_URL, skipping when it is not set.
func client(t *testing.T) *redis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set; skipping redis integration tests")
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse TEST_REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

// TestDenylist_SharedAcrossInstances is the multi-instance acceptance test:
// a revocation written by one adapter instance is visible to another.
func TestDenylist_SharedAcrossInstances(t *testing.T) {
	rdb := client(t)
	jti := fmt.Sprintf("jti-%d", time.Now().UnixNano())

	instanceA := NewDenylist(rdb)
	instanceB := NewDenylist(rdb)

	if instanceB.IsRevoked(jti) {
		t.Fatal("fresh jti must not be revoked")
	}
	instanceA.Revoke(jti, time.Now().Add(time.Hour))
	if !instanceB.IsRevoked(jti) {
		t.Error("revocation on instance A must be visible on instance B")
	}
}

func TestDenylist_ExpiresWithToken(t *testing.T) {
	rdb := client(t)
	d := NewDenylist(rdb)
	jti := fmt.Sprintf("jti-exp-%d", time.Now().UnixNano())

	d.Revoke(jti, time.Now().Add(1*time.Second))
	if !d.IsRevoked(jti) {
		t.Fatal("should be revoked immediately after Revoke")
	}
	time.Sleep(1200 * time.Millisecond)
	if d.IsRevoked(jti) {
		t.Error("revocation should expire with the token")
	}

	// Revoking an already-expired token is a no-op.
	d.Revoke("expired-"+jti, time.Now().Add(-time.Minute))
	if d.IsRevoked("expired-" + jti) {
		t.Error("expired revocation must not be stored")
	}
}

// TestRateLimiter_SharedWindow: hits from two limiter instances count against
// the same per-key budget.
func TestRateLimiter_SharedWindow(t *testing.T) {
	rdb := client(t)
	key := fmt.Sprintf("ip-%d", time.Now().UnixNano())

	a := NewRateLimiter(rdb, 3, time.Minute)
	b := NewRateLimiter(rdb, 3, time.Minute)

	for i := 1; i <= 3; i++ {
		l := a
		if i%2 == 0 {
			l = b // alternate instances — shared budget
		}
		if ok, _ := l.Allow(key); !ok {
			t.Fatalf("hit %d should be allowed", i)
		}
	}
	ok, retry := b.Allow(key)
	if ok {
		t.Error("4th hit should be blocked across instances")
	}
	if retry <= 0 || retry > 61 {
		t.Errorf("retry-after = %d, want within the window", retry)
	}

	// A different key has its own budget.
	if ok, _ := a.Allow(key + "-other"); !ok {
		t.Error("independent keys must not share budgets")
	}
}

func TestRateLimiter_WindowResets(t *testing.T) {
	rdb := client(t)
	key := fmt.Sprintf("ip-reset-%d", time.Now().UnixNano())
	rl := NewRateLimiter(rdb, 1, 1*time.Second)

	if ok, _ := rl.Allow(key); !ok {
		t.Fatal("first hit allowed")
	}
	if ok, _ := rl.Allow(key); ok {
		t.Fatal("second hit within the window must block")
	}
	time.Sleep(1200 * time.Millisecond)
	if ok, _ := rl.Allow(key); !ok {
		t.Error("a new window should open after expiry")
	}
}
