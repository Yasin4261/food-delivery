package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeChefRepo is an in-memory domain.ChefRepository for HTTP tests.
type fakeChefRepo struct {
	chefs  map[int]*domain.Chef
	nextID int
}

func newFakeChefRepo() *fakeChefRepo {
	return &fakeChefRepo{chefs: map[int]*domain.Chef{}, nextID: 1}
}

func (f *fakeChefRepo) Create(_ context.Context, c *domain.Chef) error {
	c.ID = f.nextID
	f.nextID++
	cp := *c
	f.chefs[c.ID] = &cp
	return nil
}
func (f *fakeChefRepo) FindByID(_ context.Context, id int) (*domain.Chef, error) {
	if c, ok := f.chefs[id]; ok && c.IsActive {
		cp := *c
		return &cp, nil
	}
	return nil, domain.ErrChefNotFound
}
func (f *fakeChefRepo) FindByUserID(_ context.Context, userID int) (*domain.Chef, error) {
	for _, c := range f.chefs {
		if c.UserID == userID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrChefNotFound
}
func (f *fakeChefRepo) List(_ context.Context, limit, offset int) ([]*domain.Chef, error) {
	out := make([]*domain.Chef, 0)
	for _, c := range f.chefs {
		if c.IsActive {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeChefRepo) FindNearby(_ context.Context, lat, lng float64, limit int) ([]*domain.Chef, error) {
	out := make([]*domain.Chef, 0)
	for _, c := range f.chefs {
		if c.IsActive && c.CanDeliverTo(lat, lng) {
			cp := *c
			out = append(out, &cp)
		}
	}
	return out, nil
}

func TestChef_CreateRequiresAuth(t *testing.T) {
	srv := newTestServer()
	rec := do(t, srv, http.MethodPost, "/api/v2/chefs", "",
		`{"business_name":"K","kitchen_address":"addr"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("create without token = %d, want 401", rec.Code)
	}
}

func TestChef_CreateGetListFlow(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")

	// Create the profile.
	rec := do(t, srv, http.MethodPost, "/api/v2/chefs", token,
		`{"business_name":"Yasin's Kitchen","kitchen_address":"123 Main St","kitchen_city":"Istanbul"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create chef = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	var chef domain.Chef
	if err := json.Unmarshal(rec.Body.Bytes(), &chef); err != nil {
		t.Fatalf("decode chef: %v", err)
	}
	if chef.ID == 0 || chef.DeliveryRadius != 5 {
		t.Errorf("unexpected chef: %+v", chef)
	}

	// One profile per user -> 409.
	if rec := do(t, srv, http.MethodPost, "/api/v2/chefs", token,
		`{"business_name":"Another","kitchen_address":"x"}`); rec.Code != http.StatusConflict {
		t.Errorf("second create = %d, want 409", rec.Code)
	}

	// Get by id.
	if rec := do(t, srv, http.MethodGet, "/api/v2/chefs/1", "", ""); rec.Code != http.StatusOK {
		t.Errorf("get chef = %d, want 200", rec.Code)
	}
	// Unknown id -> 404.
	if rec := do(t, srv, http.MethodGet, "/api/v2/chefs/999", "", ""); rec.Code != http.StatusNotFound {
		t.Errorf("get unknown chef = %d, want 404", rec.Code)
	}

	// List.
	rec = do(t, srv, http.MethodGet, "/api/v2/chefs", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list = %d, want 200", rec.Code)
	}
	var list []domain.Chef
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("list returned %d chefs, want 1", len(list))
	}
}

func TestChef_CreateValidation(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")

	if rec := do(t, srv, http.MethodPost, "/api/v2/chefs", token,
		`{"kitchen_address":"addr"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("missing business_name = %d, want 400", rec.Code)
	}
}

func TestChef_Nearby(t *testing.T) {
	srv := newTestServer()
	token := registerAndToken(t, srv, "yasin", "yasin@example.com")

	_ = do(t, srv, http.MethodPost, "/api/v2/chefs", token,
		`{"business_name":"K","kitchen_address":"addr","latitude":41.0082,"longitude":28.9784,"delivery_radius":10}`)

	// Missing coordinates -> 400.
	if rec := do(t, srv, http.MethodGet, "/api/v2/chefs/nearby", "", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("nearby without coords = %d, want 400", rec.Code)
	}

	// Nearby query at the kitchen location -> finds the chef.
	rec := do(t, srv, http.MethodGet, "/api/v2/chefs/nearby?lat=41.0082&lng=28.9784", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("nearby = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var list []domain.Chef
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode nearby: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("nearby returned %d chefs, want 1", len(list))
	}
}
