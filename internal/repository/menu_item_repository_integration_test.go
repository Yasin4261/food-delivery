//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestMenuItemRepository_CRUDAndScan(t *testing.T) {
	resetDB(t)
	repo := repository.NewMenuItemRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)

	it := domain.NewMenuItem(menu.ID, chef.ID, "Soup", 4.50)
	qty := 10
	it.AvailableQuantity = &qty
	it.IsVegan = true
	cat := "soup"
	it.Category = &cat
	if err := repo.Create(ctx(), it); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Round-trip every scanned field of interest (decimal price, pointers).
	got, err := repo.FindByID(ctx(), it.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Price != 4.50 {
		t.Errorf("price = %v, want 4.50 (decimal scan)", got.Price)
	}
	if !got.IsVegan || got.Category == nil || *got.Category != "soup" {
		t.Errorf("flags/category not persisted: %+v", got)
	}
	if got.AvailableQuantity == nil || *got.AvailableQuantity != 10 {
		t.Errorf("available_quantity = %v, want 10", got.AvailableQuantity)
	}
	// DB defaults applied.
	if !got.IsActive || !got.IsAvailable || got.ServingSize != 1 {
		t.Errorf("defaults wrong: %+v", got)
	}

	list, err := repo.ListByMenu(ctx(), menu.ID)
	if err != nil || len(list) != 1 {
		t.Errorf("list by menu = %d, %v", len(list), err)
	}
	chefList, err := repo.ListByChef(ctx(), chef.ID, 20, 0)
	if err != nil || len(chefList) != 1 {
		t.Errorf("list by chef = %d, %v", len(chefList), err)
	}
}

func TestMenuItemRepository_DecrementStock(t *testing.T) {
	resetDB(t)
	repo := repository.NewMenuItemRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	it := seedItem(t, menu.ID, chef.ID, 5, 3)

	// Decrement within stock.
	if err := repo.DecrementStock(ctx(), it.ID, 2); err != nil {
		t.Fatalf("decrement: %v", err)
	}
	got, _ := repo.FindByID(ctx(), it.ID)
	if *got.AvailableQuantity != 1 {
		t.Errorf("stock = %d, want 1", *got.AvailableQuantity)
	}

	// Over-decrement is refused atomically and leaves stock untouched.
	if err := repo.DecrementStock(ctx(), it.ID, 5); !errors.Is(err, domain.ErrItemOutOfStock) {
		t.Errorf("over-decrement = %v, want ErrItemOutOfStock", err)
	}
	got, _ = repo.FindByID(ctx(), it.ID)
	if *got.AvailableQuantity != 1 {
		t.Errorf("stock changed after failed decrement: %d", *got.AvailableQuantity)
	}
}

func TestMenuItemRepository_DecrementUnlimited(t *testing.T) {
	resetDB(t)
	repo := repository.NewMenuItemRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)

	it := domain.NewMenuItem(menu.ID, chef.ID, "Bread", 2)
	it.IsUnlimited = true
	if err := repo.Create(ctx(), it); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Unlimited items have no tracked quantity, so the guarded UPDATE matches
	// no row → ErrItemOutOfStock (the service skips unlimited items).
	if err := repo.DecrementStock(ctx(), it.ID, 1); !errors.Is(err, domain.ErrItemOutOfStock) {
		t.Errorf("decrement unlimited = %v, want ErrItemOutOfStock", err)
	}
}
