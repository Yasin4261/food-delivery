//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func buildAddress(userID int, label string, isDefault bool) *domain.Address {
	city := "Istanbul"
	return &domain.Address{
		UserID: userID, Label: label, Address: label + " street 1", City: &city, IsDefault: isDefault,
	}
}

func TestAddressRepository_CRUDAndDefaultSwap(t *testing.T) {
	resetDB(t)
	repo := repository.NewAddressRepository(testDB)
	user := seedUser(t, "addr@example.com")

	home := buildAddress(user.ID, "Home", true)
	if err := repo.Create(ctx(), home); err != nil {
		t.Fatalf("create home: %v", err)
	}
	if home.ID == 0 {
		t.Fatal("id not back-filled")
	}

	// Creating a new default clears the old one atomically — the partial
	// unique index would reject two defaults, so this proves the tx works.
	work := buildAddress(user.ID, "Work", true)
	if err := repo.Create(ctx(), work); err != nil {
		t.Fatalf("create work: %v", err)
	}
	list, err := repo.ListByUser(ctx(), user.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].Label != "Work" || !list[0].IsDefault || list[1].IsDefault {
		t.Errorf("default swap wrong: %+v / %+v", list[0], list[1])
	}

	// Update flips the default back.
	fresh, _ := repo.FindByID(ctx(), home.ID)
	fresh.Label = "Home base"
	fresh.IsDefault = true
	if err := repo.Update(ctx(), fresh); err != nil {
		t.Fatalf("update: %v", err)
	}
	list, _ = repo.ListByUser(ctx(), user.ID)
	if list[0].Label != "Home base" || !list[0].IsDefault {
		t.Errorf("update/default flip wrong: %+v", list[0])
	}

	// Delete.
	if err := repo.Delete(ctx(), work.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(ctx(), work.ID); !errors.Is(err, domain.ErrAddressNotFound) {
		t.Errorf("find deleted = %v, want ErrAddressNotFound", err)
	}
	if err := repo.Delete(ctx(), work.ID); !errors.Is(err, domain.ErrAddressNotFound) {
		t.Errorf("double delete = %v, want ErrAddressNotFound", err)
	}
}

func TestAddressRepository_ScopedToUser(t *testing.T) {
	resetDB(t)
	repo := repository.NewAddressRepository(testDB)
	alice := seedUser(t, "alice@example.com")
	bob := seedUser(t, "bob@example.com")

	if err := repo.Create(ctx(), buildAddress(alice.ID, "Home", true)); err != nil {
		t.Fatalf("create alice: %v", err)
	}
	// Bob's default does not disturb Alice's.
	if err := repo.Create(ctx(), buildAddress(bob.ID, "Home", true)); err != nil {
		t.Fatalf("create bob: %v", err)
	}
	aliceList, _ := repo.ListByUser(ctx(), alice.ID)
	bobList, _ := repo.ListByUser(ctx(), bob.ID)
	if len(aliceList) != 1 || len(bobList) != 1 || !aliceList[0].IsDefault || !bobList[0].IsDefault {
		t.Errorf("per-user defaults wrong: alice=%+v bob=%+v", aliceList, bobList)
	}
}

// Deleting a user cascades to their address book (FK on delete cascade).
func TestAddressRepository_CascadeOnUserDelete(t *testing.T) {
	resetDB(t)
	repo := repository.NewAddressRepository(testDB)
	user := seedUser(t, "gone@example.com")
	a := buildAddress(user.ID, "Home", true)
	if err := repo.Create(ctx(), a); err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := testDB.Exec(`DELETE FROM users WHERE id = $1`, user.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if _, err := repo.FindByID(ctx(), a.ID); !errors.Is(err, domain.ErrAddressNotFound) {
		t.Errorf("address survived user deletion: %v", err)
	}
}
