//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestSearchRepository_ChefsAndFood(t *testing.T) {
	resetDB(t)
	chefRepo := repository.NewChefRepository(testDB)
	itemRepo := repository.NewMenuItemRepository(testDB)
	search := repository.NewSearchRepository(testDB)

	// Two chefs; only one matches "pizza".
	pizza := domain.NewChef(seedUser(t, "p@e.com").ID, "Pizza Palace", "addr")
	if err := chefRepo.Create(ctx(), pizza); err != nil {
		t.Fatalf("create pizza chef: %v", err)
	}
	sushi := domain.NewChef(seedUser(t, "s@e.com").ID, "Sushi Spot", "addr")
	_ = chefRepo.Create(ctx(), sushi)

	// The adapter wraps the term in %…% itself; pass the raw query.
	got, _, err := search.SearchChefs(ctx(), "pizza", 20, 0)
	if err != nil {
		t.Fatalf("search chefs: %v", err)
	}
	if len(got) != 1 || got[0].ID != pizza.ID {
		t.Errorf("chef search returned %d, want only Pizza Palace", len(got))
	}
	// Case-insensitive.
	if up, _, _ := search.SearchChefs(ctx(), "PIZZA", 20, 0); len(up) != 1 {
		t.Errorf("case-insensitive search returned %d, want 1", len(up))
	}

	// Dishes.
	menu := seedMenu(t, pizza.ID)
	margherita := domain.NewMenuItem(menu.ID, pizza.ID, "Margherita Pizza", 8)
	if err := itemRepo.Create(ctx(), margherita); err != nil {
		t.Fatalf("create dish: %v", err)
	}
	_ = itemRepo.Create(ctx(), domain.NewMenuItem(menu.ID, pizza.ID, "Green Salad", 5))

	dishes, _, err := search.SearchMenuItems(ctx(), "pizza", 20, 0)
	if err != nil {
		t.Fatalf("search dishes: %v", err)
	}
	if len(dishes) != 1 || dishes[0].ID != margherita.ID {
		t.Errorf("dish search returned %d, want only Margherita Pizza", len(dishes))
	}
}

func TestSearchRepository_Users(t *testing.T) {
	resetDB(t)
	search := repository.NewSearchRepository(testDB)
	seedUser(t, "alice@example.com")
	seedUser(t, "bob@example.com")

	// Match by email substring.
	got, _, err := search.SearchUsers(ctx(), "alice", 20, 0)
	if err != nil {
		t.Fatalf("search users: %v", err)
	}
	if len(got) != 1 || got[0].Email != "alice@example.com" {
		t.Errorf("user search returned %d, want only alice", len(got))
	}
}
