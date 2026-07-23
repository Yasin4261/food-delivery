package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
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

// matchesQuery mirrors the adapter's case-insensitive substring match.
func matchesQuery(q string, fields ...string) bool {
	if q == "" {
		return true
	}
	q = strings.ToLower(q)
	for _, f := range fields {
		if strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}

func (f *fakeAdminRepo) ListUsers(_ context.Context, filters domain.AdminUserFilters, limit, offset int) ([]*domain.User, int, error) {
	all := make([]*domain.User, 0, len(f.users.users))
	for _, u := range f.users.users {
		if !matchesQuery(filters.Query, u.Email, u.Username) {
			continue
		}
		if filters.Role != "" && u.Role != filters.Role {
			continue
		}
		if filters.Active != nil && u.IsActive != *filters.Active {
			continue
		}
		cp := *u
		all = append(all, &cp)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	// total is the count of MATCHING rows, not the table size — the page
	// envelope must agree with the filter.
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

func (f *fakeAdminRepo) ListChefs(_ context.Context, filters domain.AdminChefFilters, limit, offset int) ([]*domain.Chef, int, error) {
	all := make([]*domain.Chef, 0, len(f.chefs.chefs))
	for _, c := range f.chefs.chefs {
		if !matchesQuery(filters.Query, c.BusinessName) {
			continue
		}
		if filters.Active != nil && c.IsActive != *filters.Active {
			continue
		}
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

func (f *fakeAdminRepo) ListOrders(_ context.Context, filters domain.AdminOrderFilters, limit, offset int) ([]*domain.Order, int, error) {
	all := make([]*domain.Order, 0, len(f.orders.orders))
	for _, o := range f.orders.orders {
		if filters.Status != "" && o.Status != filters.Status {
			continue
		}
		if filters.PaymentStatus != "" && o.PaymentStatus != filters.PaymentStatus {
			continue
		}
		if filters.UserID != 0 && o.UserID != filters.UserID {
			continue
		}
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

func TestAdminHTTP_PromoCodesAndCheckout(t *testing.T) {
	srv, _, users := newTestServerWithRepos()
	admin := registerCustomerToken(t, srv, "boss", "boss@example.com")
	promoteAdmin(t, users, "boss@example.com")
	admin = loginToken(t, srv, "boss@example.com")

	// Non-admin can't manage promos.
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodGet, "/api/v2/admin/promos", cust, ""); rec.Code != http.StatusForbidden {
		t.Errorf("customer list promos = %d, want 403", rec.Code)
	}

	// Create a 20% code.
	rec := do(t, srv, http.MethodPost, "/api/v2/admin/promos", admin,
		`{"code":"welcome20","discount_type":"percent","discount_value":20}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create promo = %d (%s)", rec.Code, rec.Body)
	}
	var promo map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &promo)
	if promo["code"] != "WELCOME20" {
		t.Errorf("code not normalised: %v", promo["code"])
	}

	// Invalid definition -> 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/admin/promos", admin,
		`{"code":"bad","discount_type":"half","discount_value":10}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad promo type = %d, want 400", rec.Code)
	}
	// Duplicate -> 409.
	if rec := do(t, srv, http.MethodPost, "/api/v2/admin/promos", admin,
		`{"code":"welcome20","discount_type":"fixed","discount_value":5}`); rec.Code != http.StatusConflict {
		t.Errorf("duplicate promo = %d, want 409", rec.Code)
	}

	// A customer applies it at checkout.
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	_ = chefToken
	body := `{"delivery_address":"x","payment_method":"cash","promo_code":"welcome20","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":2}]}`
	rec = do(t, srv, http.MethodPost, "/api/v2/orders", cust, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("order with promo = %d (%s)", rec.Code, rec.Body)
	}
	var order map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order["discount"] != float64(2) { // seedChefWithItem price 5 * 2 = 10, 20% = 2
		t.Errorf("discount = %v, want 2", order["discount"])
	}
	if order["promo_code"] != "WELCOME20" {
		t.Errorf("promo not snapshotted on order: %v", order["promo_code"])
	}

	// Deactivate it -> a new order can't use it (422).
	id := int(promo["id"].(float64))
	if rec := do(t, srv, http.MethodPatch, "/api/v2/admin/promos/"+itoa(id)+"/active", admin, `{"active":false}`); rec.Code != http.StatusOK {
		t.Fatalf("deactivate promo = %d", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", cust, body); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("order with deactivated promo = %d, want 422", rec.Code)
	}
}

// Search/filter/pagination on the admin lists (#118): the support console must
// be able to find one person or one class of order, and the page envelope's
// total must reflect the FILTER, not the table size.
func TestAdminHTTP_ListFilters(t *testing.T) {
	srv, _, users := newTestServerWithRepos()
	admin := registerCustomerToken(t, srv, "boss", "boss@example.com")
	promoteAdmin(t, users, "boss@example.com")
	admin = loginToken(t, srv, "boss@example.com")

	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	_ = chefToken
	cust := registerCustomerToken(t, srv, "alice", "alice@example.com")
	registerCustomerToken(t, srv, "bob", "bob@example.com")

	listUsers := func(qs string) pageResp[map[string]any] {
		t.Helper()
		rec := do(t, srv, http.MethodGet, "/api/v2/admin/users"+qs, admin, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET users%s = %d (%s)", qs, rec.Code, rec.Body)
		}
		return decodePage[map[string]any](t, rec.Body.Bytes())
	}

	// Free-text over email/username, case-insensitively.
	if p := listUsers("?q=alice"); p.Total != 1 || p.Data[0]["username"] != "alice" {
		t.Errorf("q=alice -> total=%d data=%v, want exactly alice", p.Total, p.Data)
	}
	if p := listUsers("?q=ALICE"); p.Total != 1 {
		t.Errorf("q=ALICE -> total=%d, want case-insensitive match", p.Total)
	}
	// Role filter.
	if p := listUsers("?role=chef"); p.Total != 1 || p.Data[0]["role"] != "chef" {
		t.Errorf("role=chef -> total=%d, want only the chef", p.Total)
	}
	// A filter that matches nothing yields an empty page, not everything.
	if p := listUsers("?q=nobody-here"); p.Total != 0 || len(p.Data) != 0 {
		t.Errorf("q=nobody-here -> total=%d len=%d, want 0/0", p.Total, len(p.Data))
	}
	// Unknown role is rejected rather than silently ignored.
	if rec := do(t, srv, http.MethodGet, "/api/v2/admin/users?role=wizard", admin, ""); rec.Code != http.StatusBadRequest {
		t.Errorf("role=wizard = %d, want 400", rec.Code)
	}

	// active= is tri-state: absent means both.
	both := listUsers("").Total
	if p := listUsers("?active=true"); p.Total != both {
		t.Errorf("active=true total=%d, want %d (all seeded users are active)", p.Total, both)
	}
	if p := listUsers("?active=false"); p.Total != 0 {
		t.Errorf("active=false total=%d, want 0", p.Total)
	}

	// Pagination reports the full matching total, not the page size.
	p := listUsers("?limit=1")
	if len(p.Data) != 1 || p.Total != both {
		t.Errorf("limit=1 -> len=%d total=%d, want 1/%d", len(p.Data), p.Total, both)
	}

	// Orders: filter by status, payment status and customer.
	body := `{"delivery_address":"1 St","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", cust, body); rec.Code != http.StatusCreated {
		t.Fatalf("place order = %d (%s)", rec.Code, rec.Body)
	}
	listOrders := func(qs string) pageResp[map[string]any] {
		t.Helper()
		rec := do(t, srv, http.MethodGet, "/api/v2/admin/orders"+qs, admin, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET orders%s = %d (%s)", qs, rec.Code, rec.Body)
		}
		return decodePage[map[string]any](t, rec.Body.Bytes())
	}
	if p := listOrders("?status=pending"); p.Total != 1 {
		t.Errorf("status=pending total=%d, want 1", p.Total)
	}
	if p := listOrders("?status=delivered"); p.Total != 0 {
		t.Errorf("status=delivered total=%d, want 0", p.Total)
	}
	if p := listOrders("?payment_status=paid"); p.Total != 0 {
		t.Errorf("payment_status=paid total=%d, want 0 (cash order is pending)", p.Total)
	}
	if p := listOrders("?user_id=9999"); p.Total != 0 {
		t.Errorf("user_id=9999 total=%d, want 0", p.Total)
	}
	// Unknown lifecycle values are rejected.
	for _, qs := range []string{"?status=teleported", "?payment_status=maybe"} {
		if rec := do(t, srv, http.MethodGet, "/api/v2/admin/orders"+qs, admin, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("orders%s = %d, want 400", qs, rec.Code)
		}
	}

	// Chefs: free-text over the business name.
	rec := do(t, srv, http.MethodGet, "/api/v2/admin/chefs?q=zzz-no-such-kitchen", admin, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET chefs = %d", rec.Code)
	}
	if p := decodePage[map[string]any](t, rec.Body.Bytes()); p.Total != 0 {
		t.Errorf("chef q=zzz total=%d, want 0", p.Total)
	}

	// Filtering stays admin-only.
	if rec := do(t, srv, http.MethodGet, "/api/v2/admin/users?q=alice", cust, ""); rec.Code != http.StatusForbidden {
		t.Errorf("customer filtering users = %d, want 403", rec.Code)
	}
}
