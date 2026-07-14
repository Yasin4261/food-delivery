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
	got, _, err := search.SearchChefs(ctx(), "pizza", domain.SearchFilters{}, 20, 0)
	if err != nil {
		t.Fatalf("search chefs: %v", err)
	}
	if len(got) != 1 || got[0].ID != pizza.ID {
		t.Errorf("chef search returned %d, want only Pizza Palace", len(got))
	}
	// Case-insensitive.
	if up, _, _ := search.SearchChefs(ctx(), "PIZZA", domain.SearchFilters{}, 20, 0); len(up) != 1 {
		t.Errorf("case-insensitive search returned %d, want 1", len(up))
	}

	// Dishes.
	menu := seedMenu(t, pizza.ID)
	margherita := domain.NewMenuItem(menu.ID, pizza.ID, "Margherita Pizza", 8)
	if err := itemRepo.Create(ctx(), margherita); err != nil {
		t.Fatalf("create dish: %v", err)
	}
	_ = itemRepo.Create(ctx(), domain.NewMenuItem(menu.ID, pizza.ID, "Green Salad", 5))

	dishes, _, err := search.SearchMenuItems(ctx(), "pizza", domain.SearchFilters{}, 20, 0)
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

// setChefStats forces a chef's aggregate rating/order counters to set up
// filter/sort scenarios.
func setChefStats(t *testing.T, chefID int, rating float64, totalOrders int) {
	t.Helper()
	if _, err := testDB.Exec(`UPDATE chefs SET rating = $2, total_orders = $3 WHERE id = $1`, chefID, rating, totalOrders); err != nil {
		t.Fatalf("set chef stats: %v", err)
	}
}

// setItemStats forces a dish's rating/order counters.
func setItemStats(t *testing.T, itemID int, rating float64, totalOrders int) {
	t.Helper()
	if _, err := testDB.Exec(`UPDATE menu_items SET rating = $2, total_orders = $3 WHERE id = $1`, itemID, rating, totalOrders); err != nil {
		t.Fatalf("set item stats: %v", err)
	}
}

func TestSearchRepository_ChefFiltersAndSorts(t *testing.T) {
	resetDB(t)
	repo := repository.NewSearchRepository(testDB)

	// Three chefs matching "Kitchen": ratings 4.8 / 3.0 / 4.2, orders 5 / 50 / 20.
	a := seedChef(t, seedUser(t, "a@example.com").ID)
	b := seedChef(t, seedUser(t, "b@example.com").ID)
	c := seedChef(t, seedUser(t, "c@example.com").ID)
	setChefStats(t, a.ID, 4.8, 5)
	setChefStats(t, b.ID, 3.0, 50)
	setChefStats(t, c.ID, 4.2, 20)

	f := domain.SearchFilters{MinRating: 4}
	got, total, err := repo.SearchChefs(ctx(), "Kitchen", f, 20, 0)
	if err != nil {
		t.Fatalf("min rating: %v", err)
	}
	if total != 2 || len(got) != 2 {
		t.Fatalf("min_rating=4 -> %d chefs, want 2", total)
	}
	for _, ch := range got {
		if ch.Rating < 4 {
			t.Errorf("chef %d rating %v below filter", ch.ID, ch.Rating)
		}
	}

	pop, _, err := repo.SearchChefs(ctx(), "Kitchen", domain.SearchFilters{Sort: domain.SortPopular}, 20, 0)
	if err != nil {
		t.Fatalf("popular: %v", err)
	}
	if pop[0].ID != b.ID || pop[1].ID != c.ID || pop[2].ID != a.ID {
		t.Errorf("popular order wrong: %d,%d,%d", pop[0].ID, pop[1].ID, pop[2].ID)
	}

	rated, _, err := repo.SearchChefs(ctx(), "Kitchen", domain.SearchFilters{Sort: domain.SortRating}, 20, 0)
	if err != nil {
		t.Fatalf("rating: %v", err)
	}
	if rated[0].ID != a.ID || rated[2].ID != b.ID {
		t.Errorf("rating order wrong: %d..%d", rated[0].ID, rated[2].ID)
	}
}

func TestSearchRepository_DishFiltersAndSorts(t *testing.T) {
	resetDB(t)
	repo := repository.NewSearchRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)

	// Dishes named "Soup N": prices 5 / 12 / 30, ratings 4.5 / 3 / 5.
	cheap := seedItem(t, menu.ID, chef.ID, 5, 10)
	mid := seedItem(t, menu.ID, chef.ID, 12, 10)
	pricey := seedItem(t, menu.ID, chef.ID, 30, 10)
	setItemStats(t, cheap.ID, 4.5, 100)
	setItemStats(t, mid.ID, 3.0, 5)
	setItemStats(t, pricey.ID, 5.0, 50)
	// Distinct cuisine on one dish for the cuisine filter.
	if _, err := testDB.Exec(`UPDATE menu_items SET cuisine = 'Anatolian' WHERE id = $1`, pricey.ID); err != nil {
		t.Fatalf("set cuisine: %v", err)
	}

	q := "Soup" // seedItem names dishes "Soup"

	// Price window 6..20 -> only the mid dish.
	got, total, err := repo.SearchMenuItems(ctx(), q, domain.SearchFilters{MinPrice: 6, MaxPrice: 20}, 20, 0)
	if err != nil {
		t.Fatalf("price window: %v", err)
	}
	if total != 1 || got[0].ID != mid.ID {
		t.Fatalf("price window -> %d/%v, want just the mid dish", total, got)
	}

	// min_rating 4 -> cheap + pricey.
	if _, total, err = repo.SearchMenuItems(ctx(), q, domain.SearchFilters{MinRating: 4}, 20, 0); err != nil || total != 2 {
		t.Fatalf("min rating -> %d (%v), want 2", total, err)
	}

	// Cuisine (substring, case-insensitive) -> the Anatolian dish.
	byCuisine, total, err := repo.SearchMenuItems(ctx(), q, domain.SearchFilters{Cuisine: "anatol"}, 20, 0)
	if err != nil || total != 1 || byCuisine[0].ID != pricey.ID {
		t.Fatalf("cuisine -> %d (%v), want the Anatolian dish", total, err)
	}

	// Sorts.
	asc, _, _ := repo.SearchMenuItems(ctx(), q, domain.SearchFilters{Sort: domain.SortPriceAsc}, 20, 0)
	if asc[0].ID != cheap.ID || asc[2].ID != pricey.ID {
		t.Errorf("price_asc order wrong")
	}
	desc, _, _ := repo.SearchMenuItems(ctx(), q, domain.SearchFilters{Sort: domain.SortPriceDesc}, 20, 0)
	if desc[0].ID != pricey.ID {
		t.Errorf("price_desc order wrong")
	}
	pop, _, _ := repo.SearchMenuItems(ctx(), q, domain.SearchFilters{Sort: domain.SortPopular}, 20, 0)
	if pop[0].ID != cheap.ID {
		t.Errorf("popular order wrong")
	}
}

func TestChefRepository_ListFiltersAndSorts(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefRepository(testDB)
	a := seedChef(t, seedUser(t, "a@example.com").ID)
	b := seedChef(t, seedUser(t, "b@example.com").ID)
	setChefStats(t, a.ID, 4.5, 5)
	setChefStats(t, b.ID, 3.5, 50)

	got, total, err := repo.List(ctx(), domain.ChefListFilters{MinRating: 4}, 20, 0)
	if err != nil || total != 1 || got[0].ID != a.ID {
		t.Fatalf("min rating list -> %d (%v), want just chef a", total, err)
	}
	pop, _, err := repo.List(ctx(), domain.ChefListFilters{Sort: domain.SortPopular}, 20, 0)
	if err != nil || pop[0].ID != b.ID {
		t.Fatalf("popular list wrong (%v)", err)
	}
}
