//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestChefRepository_CreateAndFind(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefRepository(testDB)
	u := seedUser(t, "chef@example.com")

	c := domain.NewChef(u.ID, "Yasin's Kitchen", "123 Main St")
	if err := repo.Create(ctx(), c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID == 0 {
		t.Fatal("create did not back-fill id")
	}

	byID, err := repo.FindByID(ctx(), c.ID)
	if err != nil || byID.BusinessName != "Yasin's Kitchen" {
		t.Errorf("find by id = %+v, %v", byID, err)
	}
	byUser, err := repo.FindByUserID(ctx(), u.ID)
	if err != nil || byUser.ID != c.ID {
		t.Errorf("find by user = %+v, %v", byUser, err)
	}
	if _, err := repo.FindByID(ctx(), 999); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("find missing = %v, want ErrChefNotFound", err)
	}
}

// TestChefRepository_FindNearby exercises the SQL Haversine filtering, which no
// unit test can cover.
func TestChefRepository_FindNearby(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefRepository(testDB)

	// Istanbul chef, 10km radius.
	istLat, istLng := 41.0082, 28.9784
	near := domain.NewChef(seedUser(t, "ist@example.com").ID, "Istanbul Kitchen", "addr")
	near.KitchenLatitude, near.KitchenLongitude, near.DeliveryRadius = &istLat, &istLng, 10
	if err := repo.Create(ctx(), near); err != nil {
		t.Fatalf("create near: %v", err)
	}

	// Ankara chef (~350km away), 10km radius — must be excluded.
	ankLat, ankLng := 39.9334, 32.8597
	far := domain.NewChef(seedUser(t, "ank@example.com").ID, "Ankara Kitchen", "addr")
	far.KitchenLatitude, far.KitchenLongitude, far.DeliveryRadius = &ankLat, &ankLng, 10
	if err := repo.Create(ctx(), far); err != nil {
		t.Fatalf("create far: %v", err)
	}

	got, err := repo.FindNearby(ctx(), istLat, istLng, 20, false)
	if err != nil {
		t.Fatalf("find nearby: %v", err)
	}
	if len(got) != 1 || got[0].ID != near.ID {
		t.Errorf("nearby returned %d chefs, want only the Istanbul kitchen", len(got))
	}
}

func TestChefRepository_SetOnlineAndFilter(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefRepository(testDB)
	lat, lng := 41.0082, 28.9784

	c := domain.NewChef(seedUser(t, "chef@example.com").ID, "Kitchen", "addr")
	c.KitchenLatitude, c.KitchenLongitude, c.DeliveryRadius = &lat, &lng, 10
	if err := repo.Create(ctx(), c); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Offline by default → excluded from online-only nearby.
	if got, _ := repo.FindNearby(ctx(), lat, lng, 20, true); len(got) != 0 {
		t.Errorf("online-only nearby before toggle = %d, want 0", len(got))
	}

	if err := repo.SetOnline(ctx(), c.ID, true); err != nil {
		t.Fatalf("set online: %v", err)
	}
	got, _ := repo.FindByID(ctx(), c.ID)
	if !got.IsOnline {
		t.Error("is_online not persisted")
	}
	if near, _ := repo.FindNearby(ctx(), lat, lng, 20, true); len(near) != 1 {
		t.Errorf("online-only nearby after toggle = %d, want 1", len(near))
	}
}

func TestChefRepository_List(t *testing.T) {
	resetDB(t)
	repo := repository.NewChefRepository(testDB)
	for i, email := range []string{"a@e.com", "b@e.com", "c@e.com"} {
		c := domain.NewChef(seedUser(t, email).ID, "Kitchen", "addr")
		c.Rating = float64(i) // ascending; List orders by rating DESC
		if err := repo.Create(ctx(), c); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	chefs, err := repo.List(ctx(), 2, 0, false)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(chefs) != 2 {
		t.Fatalf("list limit not applied: got %d, want 2", len(chefs))
	}
	if chefs[0].Rating < chefs[1].Rating {
		t.Errorf("list not ordered by rating desc: %v, %v", chefs[0].Rating, chefs[1].Rating)
	}
}
