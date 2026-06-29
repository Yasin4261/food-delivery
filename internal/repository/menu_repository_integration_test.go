//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestMenuRepository_CRUD(t *testing.T) {
	resetDB(t)
	repo := repository.NewMenuRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)

	m := domain.NewMenu(chef.ID, "Dinner")
	desc := "evening menu"
	m.Description = &desc
	if err := repo.Create(ctx(), m); err != nil {
		t.Fatalf("create: %v", err)
	}
	if m.ID == 0 {
		t.Fatal("create did not back-fill id")
	}

	got, err := repo.FindByID(ctx(), m.ID)
	if err != nil || got.Name != "Dinner" || got.Description == nil || *got.Description != "evening menu" {
		t.Errorf("find = %+v, %v", got, err)
	}

	// Update editable fields and confirm updated_at advances.
	m.Name = "Supper"
	m.IsFeatured = true
	prev := got.UpdatedAt
	if err := repo.Update(ctx(), m); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ = repo.FindByID(ctx(), m.ID)
	if got.Name != "Supper" || !got.IsFeatured {
		t.Errorf("update not persisted: %+v", got)
	}
	if !got.UpdatedAt.After(prev) {
		t.Errorf("updated_at did not advance: %v !> %v", got.UpdatedAt, prev)
	}

	// Deactivate hides it from the chef listing.
	if err := repo.Deactivate(ctx(), m.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	list, err := repo.ListByChef(ctx(), chef.ID, 20, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("deactivated menu still listed: %d", len(list))
	}
}

func TestMenuRepository_NotFound(t *testing.T) {
	resetDB(t)
	repo := repository.NewMenuRepository(testDB)
	if _, err := repo.FindByID(ctx(), 999); !errors.Is(err, domain.ErrMenuNotFound) {
		t.Errorf("find missing = %v, want ErrMenuNotFound", err)
	}
	if err := repo.Deactivate(ctx(), 999); !errors.Is(err, domain.ErrMenuNotFound) {
		t.Errorf("deactivate missing = %v, want ErrMenuNotFound", err)
	}
}
