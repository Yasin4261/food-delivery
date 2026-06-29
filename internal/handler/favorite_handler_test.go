package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeFavoriteRepo is an in-memory domain.FavoriteRepository for HTTP tests. It
// resolves favorited chefs through the shared chef repo.
type fakeFavoriteRepo struct {
	chefs *fakeChefRepo
	// set[userID] is the ordered list of favorited chef ids.
	set map[int][]int
}

func newFakeFavoriteRepo(chefs *fakeChefRepo) *fakeFavoriteRepo {
	return &fakeFavoriteRepo{chefs: chefs, set: map[int][]int{}}
}

func (f *fakeFavoriteRepo) Add(_ context.Context, userID, chefID int) error {
	for _, id := range f.set[userID] {
		if id == chefID {
			return nil // idempotent
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

func TestFavorites_Flow(t *testing.T) {
	srv := newTestServer()
	// A chef to favorite (chef id 1).
	chefToken := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefToken)
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// No token -> 401.
	if rec := do(t, srv, http.MethodGet, "/api/v2/favorites", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("list without token = %d, want 401", rec.Code)
	}

	// Favorite the chef -> 204; favoriting again is idempotent -> 204.
	if rec := do(t, srv, http.MethodPost, "/api/v2/favorites/1", customer, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("favorite = %d, want 204 (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/favorites/1", customer, ""); rec.Code != http.StatusNoContent {
		t.Errorf("re-favorite = %d, want 204", rec.Code)
	}

	// List shows exactly one favorite chef, with paging metadata.
	rec := do(t, srv, http.MethodGet, "/api/v2/favorites", customer, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d, want 200", rec.Code)
	}
	p := decodePage[domain.Chef](t, rec.Body.Bytes())
	if len(p.Data) != 1 || p.Data[0].ID != 1 || p.Total != 1 || p.Limit != 20 {
		t.Errorf("favorites page = %+v, want one chef id 1 / total 1", p)
	}

	// Unfavorite -> 204; list is empty.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/favorites/1", customer, ""); rec.Code != http.StatusNoContent {
		t.Errorf("unfavorite = %d, want 204", rec.Code)
	}
	rec = do(t, srv, http.MethodGet, "/api/v2/favorites", customer, "")
	if p := decodePage[domain.Chef](t, rec.Body.Bytes()); len(p.Data) != 0 || p.Total != 0 {
		t.Errorf("favorites after remove = %+v, want empty", p)
	}
}

func TestFavorites_UnknownChef(t *testing.T) {
	srv := newTestServer()
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodPost, "/api/v2/favorites/999", customer, ""); rec.Code != http.StatusNotFound {
		t.Errorf("favorite unknown chef = %d, want 404", rec.Code)
	}
}
