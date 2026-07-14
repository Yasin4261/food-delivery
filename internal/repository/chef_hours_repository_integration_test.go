//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestChefHoursRepository_ReplaceAndList(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefHoursRepository(testDB)
	a := seedChef(t, seedUser(t, "a@example.com").ID)
	b := seedChef(t, seedUser(t, "b@example.com").ID)

	monday := &domain.ChefHours{Weekday: 1, OpensAt: 9 * 60, ClosesAt: 17 * 60}
	friday := &domain.ChefHours{Weekday: 5, OpensAt: 18 * 60, ClosesAt: 2 * 60} // overnight
	if err := repo.ReplaceForChef(ctx(), a.ID, []*domain.ChefHours{friday, monday}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if monday.ID == 0 {
		t.Error("ids not back-filled")
	}

	got, err := repo.ListByChef(ctx(), a.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// Ordered by weekday, opens_at regardless of insert order.
	if len(got) != 2 || got[0].Weekday != 1 || got[1].Weekday != 5 || got[1].ClosesAt != 2*60 {
		t.Errorf("list wrong: %+v", got)
	}

	// Replace swaps the whole schedule atomically.
	if err := repo.ReplaceForChef(ctx(), a.ID, []*domain.ChefHours{{Weekday: 2, OpensAt: 600, ClosesAt: 900}}); err != nil {
		t.Fatalf("replace 2: %v", err)
	}
	got, _ = repo.ListByChef(ctx(), a.ID)
	if len(got) != 1 || got[0].Weekday != 2 {
		t.Errorf("replace did not swap: %+v", got)
	}

	// Empty replace clears; chef b never had rows.
	if err := repo.ReplaceForChef(ctx(), a.ID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	byChef, err := repo.ListByChefs(ctx(), []int{a.ID, b.ID})
	if err != nil {
		t.Fatalf("list by chefs: %v", err)
	}
	if len(byChef) != 0 {
		t.Errorf("cleared chefs still present: %v", byChef)
	}

	// Batched query groups per chef.
	_ = repo.ReplaceForChef(ctx(), a.ID, []*domain.ChefHours{monday})
	_ = repo.ReplaceForChef(ctx(), b.ID, []*domain.ChefHours{{Weekday: 3, OpensAt: 60, ClosesAt: 120}})
	byChef, _ = repo.ListByChefs(ctx(), []int{a.ID, b.ID})
	if len(byChef[a.ID]) != 1 || len(byChef[b.ID]) != 1 {
		t.Errorf("grouping wrong: %v", byChef)
	}
}
