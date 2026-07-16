//go:build integration

package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestEmailVerificationRepository_RoundTrip(t *testing.T) {
	resetDB(t)
	repo := repository.NewEmailVerificationRepository(testDB)
	users := repository.NewUserRepository(testDB)
	user := seedUser(t, "verify@example.com")

	token := &domain.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: "cafebabe",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := repo.Create(ctx(), token); err != nil {
		t.Fatalf("create: %v", err)
	}
	if token.ID == 0 || token.CreatedAt.IsZero() {
		t.Fatalf("create did not back-fill id/created_at: %+v", token)
	}

	got, err := repo.FindByHash(ctx(), "cafebabe")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.UserID != user.ID || got.UsedAt != nil || !got.Usable(time.Now()) {
		t.Errorf("unexpected token: %+v", got)
	}

	// Redeeming: mark the user verified and consume the token.
	if err := users.MarkVerified(ctx(), user.ID); err != nil {
		t.Fatalf("mark verified: %v", err)
	}
	verified, _ := users.FindByID(ctx(), user.ID)
	if !verified.IsVerified {
		t.Error("user should be verified after MarkVerified")
	}

	if err := repo.MarkUsed(ctx(), token.ID); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	got, _ = repo.FindByHash(ctx(), "cafebabe")
	if got.UsedAt == nil || got.Usable(time.Now()) {
		t.Errorf("token should be consumed: %+v", got)
	}

	if _, err := repo.FindByHash(ctx(), "missing"); !errors.Is(err, domain.ErrVerificationTokenNotFound) {
		t.Errorf("find missing = %v, want ErrVerificationTokenNotFound", err)
	}
}
