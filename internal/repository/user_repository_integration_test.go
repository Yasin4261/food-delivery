//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestUserRepository_CreateAndFind(t *testing.T) {
	resetDB(t)
	repo := repository.NewUserRepository(testDB)

	u := domain.NewUser("yasin", "yasin@example.com", "hashed")
	if err := repo.Create(ctx(), u); err != nil {
		t.Fatalf("create: %v", err)
	}
	if u.ID == 0 || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Fatalf("create did not back-fill id/timestamps: %+v", u)
	}

	byID, err := repo.FindByID(ctx(), u.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if byID.Email != "yasin@example.com" || byID.Role != domain.RoleCustomer {
		t.Errorf("unexpected user: %+v", byID)
	}
	if byID.PasswordHash != "hashed" {
		t.Errorf("password hash not persisted: %q", byID.PasswordHash)
	}

	byEmail, err := repo.FindByEmail(ctx(), "yasin@example.com")
	if err != nil || byEmail.ID != u.ID {
		t.Errorf("find by email = %+v, %v", byEmail, err)
	}
	byUsername, err := repo.FindByUsername(ctx(), "yasin")
	if err != nil || byUsername.ID != u.ID {
		t.Errorf("find by username = %+v, %v", byUsername, err)
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	resetDB(t)
	repo := repository.NewUserRepository(testDB)
	if _, err := repo.FindByID(ctx(), 999); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("find missing = %v, want ErrUserNotFound", err)
	}
}

func TestUserRepository_UniqueEmail(t *testing.T) {
	resetDB(t)
	repo := repository.NewUserRepository(testDB)
	_ = repo.Create(ctx(), domain.NewUser("a", "dup@example.com", "h"))

	// The DB unique constraint must reject the duplicate (the service checks
	// first, but the constraint is the real guard).
	err := repo.Create(ctx(), domain.NewUser("b", "dup@example.com", "h"))
	if err == nil {
		t.Error("expected a unique-violation error on duplicate email")
	}
}
