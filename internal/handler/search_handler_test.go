package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeSearchRepo returns one canned result per type (real ILIKE matching is
// covered by the repository integration tests).
type fakeSearchRepo struct{}

func newFakeSearchRepo() *fakeSearchRepo { return &fakeSearchRepo{} }

func (fakeSearchRepo) SearchChefs(_ context.Context, q string, limit, offset int) ([]*domain.Chef, error) {
	return []*domain.Chef{{ID: 1, BusinessName: "Match " + q}}, nil
}
func (fakeSearchRepo) SearchMenuItems(_ context.Context, q string, limit, offset int) ([]*domain.MenuItem, error) {
	return []*domain.MenuItem{{ID: 1, Name: "Match " + q}}, nil
}
func (fakeSearchRepo) SearchUsers(_ context.Context, q string, limit, offset int) ([]*domain.User, error) {
	return []*domain.User{{ID: 1, Username: "Match " + q, PasswordHash: "secret"}}, nil
}

// adminToken mints a valid admin JWT signed with the test secret.
func adminToken(t *testing.T) string {
	t.Helper()
	claims := service.Claims{
		UserID: 1, Email: "admin@example.com", Role: domain.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign admin token: %v", err)
	}
	return signed
}

func TestSearch_TypesAndGuards(t *testing.T) {
	srv := newTestServer()
	token := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Auth required.
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=x&type=chef", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("search without token = %d, want 401", rec.Code)
	}

	// Chef and food search work for any authenticated user.
	rec := do(t, srv, http.MethodGet, "/api/v2/search?q=pizza&type=chef", token, "")
	var chefs []domain.Chef
	_ = json.Unmarshal(rec.Body.Bytes(), &chefs)
	if rec.Code != http.StatusOK || len(chefs) != 1 {
		t.Errorf("chef search = %d/%d, want 200/1", rec.Code, len(chefs))
	}
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=soup&type=food", token, ""); rec.Code != http.StatusOK {
		t.Errorf("food search = %d, want 200", rec.Code)
	}

	// Empty query -> 400; unknown type -> 400.
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=&type=chef", token, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("empty q = %d, want 400", rec.Code)
	}
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=x&type=planet", token, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("bad type = %d, want 400", rec.Code)
	}

	// User search is admin-only.
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=x&type=user", token, ""); rec.Code != http.StatusForbidden {
		t.Errorf("customer user-search = %d, want 403", rec.Code)
	}
	if rec := do(t, srv, http.MethodGet, "/api/v2/search?q=x&type=user", adminToken(t), ""); rec.Code != http.StatusOK {
		t.Errorf("admin user-search = %d, want 200 (%s)", rec.Code, rec.Body)
	}
}
