package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeFavoriteRepo is an in-memory domain.FavoriteRepository for service tests.
type fakeFavoriteRepo struct {
	chefs *fakeChefRepo
	set   map[int][]int
}

func newFakeFavoriteRepo(chefs *fakeChefRepo) *fakeFavoriteRepo {
	return &fakeFavoriteRepo{chefs: chefs, set: map[int][]int{}}
}

func (f *fakeFavoriteRepo) Add(_ context.Context, userID, chefID int) error {
	for _, id := range f.set[userID] {
		if id == chefID {
			return nil
		}
	}
	f.set[userID] = append(f.set[userID], chefID)
	return nil
}
func (f *fakeFavoriteRepo) Remove(_ context.Context, userID, chefID int) error {
	out := f.set[userID][:0]
	for _, id := range f.set[userID] {
		if id != chefID {
			out = append(out, id)
		}
	}
	f.set[userID] = out
	return nil
}
func (f *fakeFavoriteRepo) ListChefs(ctx context.Context, userID, limit, offset int) ([]*domain.Chef, int, error) {
	out := make([]*domain.Chef, 0)
	for _, id := range f.set[userID] {
		if c, err := f.chefs.FindByID(ctx, id); err == nil {
			out = append(out, c)
		}
	}
	return out, len(out), nil
}

func favoriteFixture(t *testing.T) (*service.FavoriteService, *fakeChefRepo) {
	t.Helper()
	chefRepo := newFakeChefRepo()
	if err := chefRepo.Create(context.Background(), &domain.Chef{UserID: 1, IsActive: true}); err != nil {
		t.Fatalf("seed chef: %v", err)
	}
	svc := service.NewFavoriteService(newFakeFavoriteRepo(chefRepo), chefRepo)
	return svc, chefRepo
}

func TestFavoriteService_AddIsIdempotent(t *testing.T) {
	svc, _ := favoriteFixture(t)
	ctx := context.Background()

	if err := svc.Add(ctx, 100, 1); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := svc.Add(ctx, 100, 1); err != nil {
		t.Fatalf("second add should be a no-op: %v", err)
	}

	chefs, _, err := svc.List(ctx, 100, 20, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(chefs) != 1 {
		t.Errorf("favorites = %d, want 1 (idempotent)", len(chefs))
	}
}

func TestFavoriteService_AddUnknownChef(t *testing.T) {
	svc, _ := favoriteFixture(t)
	if err := svc.Add(context.Background(), 100, 999); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("err = %v, want ErrChefNotFound", err)
	}
}

func TestFavoriteService_Remove(t *testing.T) {
	svc, _ := favoriteFixture(t)
	ctx := context.Background()
	if err := svc.Add(ctx, 100, 1); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := svc.Remove(ctx, 100, 1); err != nil {
		t.Fatalf("remove: %v", err)
	}
	chefs, _, _ := svc.List(ctx, 100, 20, 0)
	if len(chefs) != 0 {
		t.Errorf("favorites after remove = %d, want 0", len(chefs))
	}
	// Removing again is a no-op.
	if err := svc.Remove(ctx, 100, 1); err != nil {
		t.Errorf("remove non-favorite = %v, want nil", err)
	}
}
