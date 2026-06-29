//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestFavoriteRepository_AddIdempotentAndList(t *testing.T) {
	resetDB(t)
	repo := repository.NewFavoriteRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)

	// Favoriting twice must not error and must not duplicate (ON CONFLICT).
	if err := repo.Add(ctx(), customer.ID, chef.ID); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := repo.Add(ctx(), customer.ID, chef.ID); err != nil {
		t.Fatalf("re-add (should be idempotent): %v", err)
	}

	chefs, err := repo.ListChefs(ctx(), customer.ID, 20, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(chefs) != 1 || chefs[0].ID != chef.ID {
		t.Fatalf("favorites = %d, want exactly one (id %d)", len(chefs), chef.ID)
	}

	// Remove unfavorites; removing again is a no-op.
	if err := repo.Remove(ctx(), customer.ID, chef.ID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := repo.Remove(ctx(), customer.ID, chef.ID); err != nil {
		t.Fatalf("re-remove (no-op): %v", err)
	}
	chefs, _ = repo.ListChefs(ctx(), customer.ID, 20, 0)
	if len(chefs) != 0 {
		t.Errorf("favorites after remove = %d, want 0", len(chefs))
	}
}

// TestFavoriteRepository_ListExcludesInactiveChef confirms the ListChefs query
// (chef join via correlated subquery) honours chefs.is_active.
func TestFavoriteRepository_ListExcludesInactiveChef(t *testing.T) {
	resetDB(t)
	repo := repository.NewFavoriteRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	_ = repo.Add(ctx(), customer.ID, chef.ID)

	// Soft-delete the chef directly.
	if _, err := testDB.Exec(`UPDATE chefs SET is_active = false WHERE id = $1`, chef.ID); err != nil {
		t.Fatalf("deactivate chef: %v", err)
	}
	chefs, err := repo.ListChefs(ctx(), customer.ID, 20, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(chefs) != 0 {
		t.Errorf("inactive chef still listed as favorite: %d", len(chefs))
	}
}
