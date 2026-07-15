package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeAdminRepo implements domain.AdminRepository over the shared in-memory
// fakes, so admin mutations really affect login/browse in the same test.
type fakeAdminRepo struct {
	users  *fakeUserRepo
	chefs  *fakeChefRepo
	orders *fakeOrderRepo
}

func newFakeAdminRepo(u *fakeUserRepo, c *fakeChefRepo, o *fakeOrderRepo) *fakeAdminRepo {
	return &fakeAdminRepo{users: u, chefs: c, orders: o}
}

func (f *fakeAdminRepo) ListUsers(_ context.Context, limit, offset int) ([]*domain.User, int, error) {
	all := make([]*domain.User, 0, len(f.users.users))
	for _, u := range f.users.users {
		cp := *u
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (f *fakeAdminRepo) SetUserActive(_ context.Context, userID int, active bool) error {
	u, ok := f.users.users[userID]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.IsActive = active
	return nil
}

func (f *fakeAdminRepo) ListChefs(_ context.Context, limit, offset int) ([]*domain.Chef, int, error) {
	all := make([]*domain.Chef, 0, len(f.chefs.chefs))
	for _, c := range f.chefs.chefs {
		cp := *c
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	return all, len(all), nil
}

func (f *fakeAdminRepo) SetChefActive(_ context.Context, chefID int, active bool) error {
	c, ok := f.chefs.chefs[chefID]
	if !ok {
		return domain.ErrChefNotFound
	}
	c.IsActive = active
	return nil
}

func (f *fakeAdminRepo) ListOrders(_ context.Context, limit, offset int) ([]*domain.Order, int, error) {
	all := make([]*domain.Order, 0, len(f.orders.orders))
	for _, o := range f.orders.orders {
		all = append(all, cloneOrder(o, 0))
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID > all[j].ID })
	return all, len(all), nil
}

func (f *fakeAdminRepo) Stats(_ context.Context) (*domain.PlatformStats, error) {
	s := &domain.PlatformStats{TopChefs: []domain.TopChef{}}
	s.TotalUsers = len(f.users.users)
	s.TotalChefs = len(f.chefs.chefs)
	for _, c := range f.chefs.chefs {
		if c.IsActive {
			s.ActiveChefs++
		}
	}
	s.TotalOrders = len(f.orders.orders)
	return s, nil
}

// promoteAdmin flips a registered user to the admin role directly (admin is not
// self-assignable via the API — the first admin is seeded out of band).
func promoteAdmin(t *testing.T, users *fakeUserRepo, email string) {
	t.Helper()
	for _, u := range users.users {
		if u.Email == email {
			u.Role = domain.RoleAdmin
			return
		}
	}
	t.Fatalf("user %s not found to promote", email)
}

func TestAdminHTTP_RoleGuard(t *testing.T) {
	srv, _, users := newTestServerWithRepos()
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	chefToken, _ := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	_ = users

	for _, path := range []string{"/api/v2/admin/stats", "/api/v2/admin/users", "/api/v2/admin/orders"} {
		if rec := do(t, srv, http.MethodGet, path, "", ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s anon = %d, want 401", path, rec.Code)
		}
		if rec := do(t, srv, http.MethodGet, path, customer, ""); rec.Code != http.StatusForbidden {
			t.Errorf("GET %s customer = %d, want 403", path, rec.Code)
		}
		if rec := do(t, srv, http.MethodGet, path, chefToken, ""); rec.Code != http.StatusForbidden {
			t.Errorf("GET %s chef = %d, want 403", path, rec.Code)
		}
	}
}

func TestAdminHTTP_ModerationAndStats(t *testing.T) {
	srv, _, users := newTestServerWithRepos()
	admin := registerCustomerToken(t, srv, "boss", "boss@example.com")
	promoteAdmin(t, users, "boss@example.com")
	// A fresh token so the promoted role is in the claims.
	admin = loginToken(t, srv, "boss@example.com")

	chefToken, _ := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	victim := registerCustomerToken(t, srv, "victim", "victim@example.com")
	_ = chefToken

	// Stats reachable for admin; no password hash in the user list.
	rec := do(t, srv, http.MethodGet, "/api/v2/admin/stats", admin, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("stats = %d (%s)", rec.Code, rec.Body)
	}
	rec = do(t, srv, http.MethodGet, "/api/v2/admin/users", admin, "")
	var page pageResp[map[string]any]
	_ = json.Unmarshal(rec.Body.Bytes(), &page)
	if page.Total < 3 {
		t.Errorf("users total = %d, want >= 3", page.Total)
	}
	for _, u := range page.Data {
		if _, leaked := u["password_hash"]; leaked {
			t.Fatal("password_hash leaked in admin user list")
		}
	}

	// Deactivate the victim -> their login is blocked.
	if rec := do(t, srv, http.MethodPatch, "/api/v2/admin/users/3/active", admin, `{"active":false}`); rec.Code != http.StatusOK {
		t.Fatalf("deactivate = %d (%s)", rec.Code, rec.Body)
	}
	_ = victim
	if rec := do(t, srv, http.MethodPost, "/api/v2/auth/login", "",
		`{"email":"victim@example.com","password":"secret123"}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("deactivated login = %d, want 401", rec.Code)
	}

	// Admin cannot deactivate their own account.
	if rec := do(t, srv, http.MethodPatch, "/api/v2/admin/users/1/active", admin, `{"active":false}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("self-deactivate = %d, want 422", rec.Code)
	}

	// Unknown user -> 404.
	if rec := do(t, srv, http.MethodPatch, "/api/v2/admin/users/9999/active", admin, `{"active":false}`); rec.Code != http.StatusNotFound {
		t.Errorf("unknown user = %d, want 404", rec.Code)
	}
}

// A deactivated chef disappears from browse and can't receive orders.
func TestAdminHTTP_DeactivateChefHidesAndBlocks(t *testing.T) {
	srv, _, users := newTestServerWithRepos()
	admin := registerCustomerToken(t, srv, "boss", "boss@example.com")
	promoteAdmin(t, users, "boss@example.com")
	admin = loginToken(t, srv, "boss@example.com")

	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	_ = chefToken
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Chef is visible on browse before deactivation.
	rec := do(t, srv, http.MethodGet, "/api/v2/chefs", "", "")
	if p := decodePage[domain.Chef](t, rec.Body.Bytes()); len(p.Data) != 1 {
		t.Fatalf("browse before = %d chefs, want 1", len(p.Data))
	}

	if rec := do(t, srv, http.MethodPatch, "/api/v2/admin/chefs/1/active", admin, `{"active":false}`); rec.Code != http.StatusOK {
		t.Fatalf("deactivate chef = %d (%s)", rec.Code, rec.Body)
	}

	// Hidden from browse.
	rec = do(t, srv, http.MethodGet, "/api/v2/chefs", "", "")
	if p := decodePage[domain.Chef](t, rec.Body.Bytes()); len(p.Data) != 0 {
		t.Errorf("browse after = %d chefs, want 0 (deactivated hidden)", len(p.Data))
	}

	// New order to the deactivated chef is rejected: their dishes are hidden
	// and the chef lookup (active-only) fails, so a stale cart can't order.
	body := `{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", cust, body); rec.Code != http.StatusNotFound {
		t.Errorf("order to deactivated chef = %d, want 404 (%s)", rec.Code, rec.Body)
	}
}
