package service_test

import (
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/service"
)

func TestTokenDenylist_RevokeAndExpiry(t *testing.T) {
	d := service.NewTokenDenylist()

	if d.IsRevoked("jti-1") {
		t.Error("unknown jti should not be revoked")
	}

	// Revoke until 1h from now → revoked.
	d.Revoke("jti-1", time.Now().Add(time.Hour))
	if !d.IsRevoked("jti-1") {
		t.Error("jti-1 should be revoked")
	}

	// Already-expired revocation is purged → not revoked.
	d.Revoke("jti-2", time.Now().Add(-time.Minute))
	if d.IsRevoked("jti-2") {
		t.Error("expired revocation should not block")
	}

	// Empty jti is a no-op.
	d.Revoke("", time.Now().Add(time.Hour))
	if d.IsRevoked("") {
		t.Error("empty jti should never be revoked")
	}
}
