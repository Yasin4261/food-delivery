//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// ctx is a convenience background context for the integration tests.
func ctx() context.Context { return context.Background() }

// resetDB truncates every application table and restarts identity sequences so
// each test starts from a known-empty schema. schema_migrations is left intact.
func resetDB(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec(`
		TRUNCATE favorites, order_items, orders, menu_items, menus, chefs, users
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("reset db: %v", err)
	}
}

// --- seed helpers: build prerequisite rows through the real adapters so the
// foreign keys line up and the seeding path is itself exercised. ---

func seedUser(t *testing.T, email string) *domain.User {
	t.Helper()
	repo := repository.NewUserRepository(testDB)
	u := domain.NewUser("user_"+email, email, "hash")
	if err := repo.Create(ctx(), u); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func seedChef(t *testing.T, userID int) *domain.Chef {
	t.Helper()
	repo := repository.NewChefRepository(testDB)
	c := domain.NewChef(userID, "Kitchen", "123 Main St")
	if err := repo.Create(ctx(), c); err != nil {
		t.Fatalf("seed chef: %v", err)
	}
	return c
}

func seedMenu(t *testing.T, chefID int) *domain.Menu {
	t.Helper()
	repo := repository.NewMenuRepository(testDB)
	m := domain.NewMenu(chefID, "Dinner")
	if err := repo.Create(ctx(), m); err != nil {
		t.Fatalf("seed menu: %v", err)
	}
	return m
}

// seedItem inserts a limited-stock dish with the given price and quantity.
func seedItem(t *testing.T, menuID, chefID int, price float64, qty int) *domain.MenuItem {
	t.Helper()
	repo := repository.NewMenuItemRepository(testDB)
	it := domain.NewMenuItem(menuID, chefID, "Soup", price)
	it.AvailableQuantity = &qty
	if err := repo.Create(ctx(), it); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	return it
}
