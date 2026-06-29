package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// --- in-memory fakes for the menu ports ---

type fakeMenuRepo struct {
	menus  map[int]*domain.Menu
	nextID int
}

func newFakeMenuRepo() *fakeMenuRepo {
	return &fakeMenuRepo{menus: map[int]*domain.Menu{}, nextID: 1}
}

func (f *fakeMenuRepo) Create(_ context.Context, m *domain.Menu) error {
	m.ID = f.nextID
	f.nextID++
	cp := *m
	f.menus[m.ID] = &cp
	return nil
}
func (f *fakeMenuRepo) FindByID(_ context.Context, id int) (*domain.Menu, error) {
	if m, ok := f.menus[id]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, domain.ErrMenuNotFound
}
func (f *fakeMenuRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Menu, int, error) {
	out := make([]*domain.Menu, 0)
	for _, m := range f.menus {
		if m.ChefID == chefID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}
func (f *fakeMenuRepo) Update(_ context.Context, m *domain.Menu) error {
	if _, ok := f.menus[m.ID]; !ok {
		return domain.ErrMenuNotFound
	}
	cp := *m
	f.menus[m.ID] = &cp
	return nil
}
func (f *fakeMenuRepo) Deactivate(_ context.Context, id int) error {
	m, ok := f.menus[id]
	if !ok {
		return domain.ErrMenuNotFound
	}
	m.IsActive = false
	return nil
}

type fakeMenuItemRepo struct {
	items  map[int]*domain.MenuItem
	nextID int
}

func newFakeMenuItemRepo() *fakeMenuItemRepo {
	return &fakeMenuItemRepo{items: map[int]*domain.MenuItem{}, nextID: 1}
}

func (f *fakeMenuItemRepo) Create(_ context.Context, m *domain.MenuItem) error {
	m.ID = f.nextID
	f.nextID++
	cp := *m
	f.items[m.ID] = &cp
	return nil
}
func (f *fakeMenuItemRepo) FindByID(_ context.Context, id int) (*domain.MenuItem, error) {
	if m, ok := f.items[id]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, domain.ErrMenuItemNotFound
}
func (f *fakeMenuItemRepo) ListByMenu(_ context.Context, menuID int) ([]*domain.MenuItem, error) {
	out := make([]*domain.MenuItem, 0)
	for _, m := range f.items {
		if m.MenuID == menuID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeMenuItemRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.MenuItem, int, error) {
	out := make([]*domain.MenuItem, 0)
	for _, m := range f.items {
		if m.ChefID == chefID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}
func (f *fakeMenuItemRepo) Update(_ context.Context, m *domain.MenuItem) error {
	if _, ok := f.items[m.ID]; !ok {
		return domain.ErrMenuItemNotFound
	}
	cp := *m
	f.items[m.ID] = &cp
	return nil
}
func (f *fakeMenuItemRepo) Deactivate(_ context.Context, id int) error {
	m, ok := f.items[id]
	if !ok {
		return domain.ErrMenuItemNotFound
	}
	m.IsActive = false
	return nil
}
func (f *fakeMenuItemRepo) DecrementStock(_ context.Context, id, qty int) error {
	m, ok := f.items[id]
	if !ok || m.IsUnlimited || m.AvailableQuantity == nil || *m.AvailableQuantity < qty {
		return domain.ErrItemOutOfStock
	}
	*m.AvailableQuantity -= qty
	return nil
}

// registerCustomerToken registers a customer (non-chef) and returns its token.
func registerCustomerToken(t *testing.T, srv http.Handler, username, email string) string {
	t.Helper()
	body := `{"username":"` + username + `","email":"` + email + `","password":"secret123","role":"customer"}`
	rec := do(t, srv, http.MethodPost, "/api/v2/auth/register", "", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup register customer failed: %d (%s)", rec.Code, rec.Body)
	}
	var reg struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &reg); err != nil {
		t.Fatalf("decode register: %v", err)
	}
	return reg.Token
}

// createChefProfile creates a chef profile for the token's user.
func createChefProfile(t *testing.T, srv http.Handler, token string) {
	t.Helper()
	rec := do(t, srv, http.MethodPost, "/api/v2/chefs", token,
		`{"business_name":"Kitchen","kitchen_address":"addr"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup create chef failed: %d (%s)", rec.Code, rec.Body)
	}
}

func TestMenu_FullFlow(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")
	createChefProfile(t, srv, token)

	// Create a menu.
	rec := do(t, srv, http.MethodPost, "/api/v2/menus", token,
		`{"name":"Dinner","menu_type":"regular","is_featured":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create menu = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	var menu domain.Menu
	if err := json.Unmarshal(rec.Body.Bytes(), &menu); err != nil {
		t.Fatalf("decode menu: %v", err)
	}
	if menu.ID == 0 || menu.ChefID != 1 || menu.MenuType != "regular" {
		t.Errorf("unexpected menu: %+v", menu)
	}

	// Public read of the menu.
	if rec := do(t, srv, http.MethodGet, "/api/v2/menus/1", "", ""); rec.Code != http.StatusOK {
		t.Errorf("get menu = %d, want 200", rec.Code)
	}
	// List the chef's menus (public).
	rec = do(t, srv, http.MethodGet, "/api/v2/chefs/1/menus", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list chef menus = %d, want 200", rec.Code)
	}
	if p := decodePage[domain.Menu](t, rec.Body.Bytes()); len(p.Data) != 1 || p.Total != 1 {
		t.Errorf("chef menus = %+v, want one / total 1", p)
	}

	// Add a dish.
	rec = do(t, srv, http.MethodPost, "/api/v2/menu-items", token,
		`{"menu_id":1,"name":"Soup","price":4.5,"is_vegan":true,"available_quantity":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create item = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	var item domain.MenuItem
	_ = json.Unmarshal(rec.Body.Bytes(), &item)
	if item.ID == 0 || item.ChefID != 1 || item.MenuID != 1 || !item.IsVegan {
		t.Errorf("unexpected item: %+v", item)
	}

	// List items in the menu and across the chef (public).
	rec = do(t, srv, http.MethodGet, "/api/v2/menus/1/items", "", "")
	if p := decodePage[domain.MenuItem](t, rec.Body.Bytes()); rec.Code != http.StatusOK || len(p.Data) != 1 {
		t.Errorf("menu items = %d/%+v, want 200 with one", rec.Code, p)
	}
	rec = do(t, srv, http.MethodGet, "/api/v2/chefs/1/menu-items", "", "")
	if p := decodePage[domain.MenuItem](t, rec.Body.Bytes()); rec.Code != http.StatusOK || len(p.Data) != 1 {
		t.Errorf("chef items = %d/%+v, want 200 with one", rec.Code, p)
	}

	// Update the menu.
	if rec := do(t, srv, http.MethodPut, "/api/v2/menus/1", token,
		`{"name":"Supper"}`); rec.Code != http.StatusOK {
		t.Errorf("update menu = %d, want 200 (%s)", rec.Code, rec.Body)
	}

	// Delete the dish, then the menu.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/menu-items/1", token, ""); rec.Code != http.StatusNoContent {
		t.Errorf("delete item = %d, want 204", rec.Code)
	}
	if rec := do(t, srv, http.MethodDelete, "/api/v2/menus/1", token, ""); rec.Code != http.StatusNoContent {
		t.Errorf("delete menu = %d, want 204", rec.Code)
	}
	// A deactivated menu is hidden from public reads.
	if rec := do(t, srv, http.MethodGet, "/api/v2/menus/1", "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("get deactivated menu = %d, want 404", rec.Code)
	}
}

func TestMenu_AuthAndRoleGuards(t *testing.T) {
	srv := newTestServer()

	// No token -> 401.
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", "", `{"name":"X"}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("create menu without token = %d, want 401", rec.Code)
	}

	// Customer role -> 403 (chef-only endpoint).
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", customer, `{"name":"X"}`); rec.Code != http.StatusForbidden {
		t.Errorf("customer create menu = %d, want 403", rec.Code)
	}

	// Chef with no profile yet -> 404 (no chef to attach the menu to).
	chef := registerAndToken(t, srv, "chefa", "chefa@example.com")
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", chef, `{"name":"X"}`); rec.Code != http.StatusNotFound {
		t.Errorf("chef without profile create menu = %d, want 404", rec.Code)
	}
}

func TestMenu_OwnershipEnforced(t *testing.T) {
	srv := newTestServer()

	// Chef A creates a profile and a menu (chef id 1, menu id 1).
	chefA := registerAndToken(t, srv, "chefa", "chefa@example.com")
	createChefProfile(t, srv, chefA)
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", chefA, `{"name":"A"}`); rec.Code != http.StatusCreated {
		t.Fatalf("chef A create menu = %d, want 201 (%s)", rec.Code, rec.Body)
	}

	// Chef B creates a profile and tries to edit chef A's menu -> 403.
	chefB := registerAndToken(t, srv, "chefb", "chefb@example.com")
	createChefProfile(t, srv, chefB)
	if rec := do(t, srv, http.MethodPut, "/api/v2/menus/1", chefB, `{"name":"hijack"}`); rec.Code != http.StatusForbidden {
		t.Errorf("chef B update chef A menu = %d, want 403", rec.Code)
	}
	if rec := do(t, srv, http.MethodDelete, "/api/v2/menus/1", chefB, ""); rec.Code != http.StatusForbidden {
		t.Errorf("chef B delete chef A menu = %d, want 403", rec.Code)
	}
}

func TestMenu_Validation(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")
	createChefProfile(t, srv, token)

	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", token, `{"name":"  "}`); rec.Code != http.StatusBadRequest {
		t.Errorf("blank menu name = %d, want 400", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", token,
		`{"name":"X","menu_type":"brunch"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad menu_type = %d, want 400", rec.Code)
	}

	// Need a menu for item validation.
	_ = do(t, srv, http.MethodPost, "/api/v2/menus", token, `{"name":"M"}`)
	if rec := do(t, srv, http.MethodPost, "/api/v2/menu-items", token,
		`{"menu_id":1,"name":"Free","price":0}`); rec.Code != http.StatusBadRequest {
		t.Errorf("zero price = %d, want 400", rec.Code)
	}
}
