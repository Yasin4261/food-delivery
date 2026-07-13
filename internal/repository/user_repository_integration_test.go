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

func TestUserRepository_UpdateProfile(t *testing.T) {
	resetDB(t)
	repo := repository.NewUserRepository(testDB)
	user := seedUser(t, "profile@example.com")

	phone, city := "+90 555 111 22 33", "Istanbul"
	lat, lng := 41.0082, 28.9784
	user.PhoneNumber = &phone
	user.City = &city
	user.Latitude = &lat
	user.Longitude = &lng
	if err := repo.UpdateProfile(ctx(), user); err != nil {
		t.Fatalf("update profile: %v", err)
	}

	got, err := repo.FindByID(ctx(), user.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.PhoneNumber == nil || *got.PhoneNumber != phone || got.City == nil || *got.City != city {
		t.Errorf("contact fields not persisted: %+v", got)
	}
	if got.Latitude == nil || *got.Latitude != lat || got.Longitude == nil || *got.Longitude != lng {
		t.Errorf("location not persisted: lat=%v lng=%v", got.Latitude, got.Longitude)
	}
	// Identity untouched.
	if got.Email != "profile@example.com" {
		t.Errorf("email changed: %q", got.Email)
	}

	// Clearing works too (nil pointers -> NULL).
	user.PhoneNumber, user.City, user.Latitude, user.Longitude = nil, nil, nil, nil
	if err := repo.UpdateProfile(ctx(), user); err != nil {
		t.Fatalf("clear profile: %v", err)
	}
	got, _ = repo.FindByID(ctx(), user.ID)
	if got.PhoneNumber != nil || got.Latitude != nil {
		t.Errorf("fields not cleared: %+v", got)
	}

	ghost := *user
	ghost.ID = 9999
	if err := repo.UpdateProfile(ctx(), &ghost); err != domain.ErrUserNotFound {
		t.Errorf("unknown user = %v, want ErrUserNotFound", err)
	}
}
