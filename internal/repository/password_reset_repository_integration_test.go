//go:build integration

package repository_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestPasswordResetRepository_RoundTrip(t *testing.T) {
	resetDB(t)
	repo := repository.NewPasswordResetRepository(testDB)
	user := seedUser(t, "reset@example.com")

	token := &domain.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: "deadbeef",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := repo.Create(ctx(), token); err != nil {
		t.Fatalf("create: %v", err)
	}
	if token.ID == 0 || token.CreatedAt.IsZero() {
		t.Fatalf("create did not back-fill id/created_at: %+v", token)
	}

	got, err := repo.FindByHash(ctx(), "deadbeef")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.UserID != user.ID || got.UsedAt != nil || !got.Usable(time.Now()) {
		t.Errorf("unexpected token: %+v", got)
	}

	// MarkUsed consumes it.
	if err := repo.MarkUsed(ctx(), token.ID); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	got, _ = repo.FindByHash(ctx(), "deadbeef")
	if got.UsedAt == nil || got.Usable(time.Now()) {
		t.Errorf("token should be consumed: %+v", got)
	}

	if _, err := repo.FindByHash(ctx(), "missing"); !errors.Is(err, domain.ErrResetTokenNotFound) {
		t.Errorf("find missing = %v, want ErrResetTokenNotFound", err)
	}
}
