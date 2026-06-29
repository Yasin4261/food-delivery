package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeEarningsRepo returns a canned summary; the real SQL aggregation is
// exercised by the repository integration tests.
type fakeEarningsRepo struct{}

func newFakeEarningsRepo() *fakeEarningsRepo { return &fakeEarningsRepo{} }

func (fakeEarningsRepo) ChefEarnings(_ context.Context, chefID int, _ *time.Time) (*domain.Earnings, error) {
	return &domain.Earnings{ChefID: chefID, TotalEarnings: 42.50, DeliveredOrders: 3, ItemsSold: 7}, nil
}

func TestEarnings_RequiresChefAndReturnsSummary(t *testing.T) {
	srv := newTestServer()

	// Customer (non-chef) is rejected.
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodGet, "/api/v2/chefs/me/earnings", customer, ""); rec.Code != http.StatusForbidden {
		t.Errorf("customer earnings = %d, want 403", rec.Code)
	}

	// Chef without a profile -> 404 (no chef to attribute earnings to).
	chefNoProfile := registerAndToken(t, srv, "chefx", "chefx@example.com")
	if rec := do(t, srv, http.MethodGet, "/api/v2/chefs/me/earnings", chefNoProfile, ""); rec.Code != http.StatusNotFound {
		t.Errorf("chef without profile earnings = %d, want 404", rec.Code)
	}

	// Chef with a profile -> 200 with the summary.
	chefToken := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefToken)
	rec := do(t, srv, http.MethodGet, "/api/v2/chefs/me/earnings", chefToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("earnings = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var e domain.Earnings
	_ = json.Unmarshal(rec.Body.Bytes(), &e)
	if e.TotalEarnings != 42.50 || e.DeliveredOrders != 3 || e.ItemsSold != 7 {
		t.Errorf("unexpected earnings: %+v", e)
	}
}
