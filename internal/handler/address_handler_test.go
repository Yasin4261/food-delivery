package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeAddressRepo is an in-memory domain.AddressRepository for HTTP tests.
type fakeAddressRepo struct {
	addresses map[int]*domain.Address
	nextID    int
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{addresses: map[int]*domain.Address{}, nextID: 1}
}

func (f *fakeAddressRepo) clearDefault(userID int) {
	for _, a := range f.addresses {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
}

func (f *fakeAddressRepo) Create(_ context.Context, a *domain.Address) error {
	if a.IsDefault {
		f.clearDefault(a.UserID)
	}
	a.ID = f.nextID
	f.nextID++
	cp := *a
	f.addresses[a.ID] = &cp
	return nil
}

func (f *fakeAddressRepo) FindByID(_ context.Context, id int) (*domain.Address, error) {
	if a, ok := f.addresses[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, domain.ErrAddressNotFound
}

func (f *fakeAddressRepo) ListByUser(_ context.Context, userID int) ([]*domain.Address, error) {
	out := make([]*domain.Address, 0)
	for _, a := range f.addresses {
		if a.UserID == userID && a.IsDefault {
			cp := *a
			out = append(out, &cp)
		}
	}
	for _, a := range f.addresses {
		if a.UserID == userID && !a.IsDefault {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeAddressRepo) Update(_ context.Context, a *domain.Address) error {
	if _, ok := f.addresses[a.ID]; !ok {
		return domain.ErrAddressNotFound
	}
	if a.IsDefault {
		f.clearDefault(a.UserID)
	}
	cp := *a
	f.addresses[a.ID] = &cp
	return nil
}

func (f *fakeAddressRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.addresses[id]; !ok {
		return domain.ErrAddressNotFound
	}
	delete(f.addresses, id)
	return nil
}

func TestAddressHTTP_RequiresAuth(t *testing.T) {
	srv := newTestServer()
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v2/addresses"},
		{http.MethodPost, "/api/v2/addresses"},
		{http.MethodPut, "/api/v2/addresses/1"},
		{http.MethodDelete, "/api/v2/addresses/1"},
	} {
		if rec := do(t, srv, tc.method, tc.path, "", `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without token = %d, want 401", tc.method, tc.path, rec.Code)
		}
	}
}

func TestAddressHTTP_CRUDAndOwnership(t *testing.T) {
	srv := newTestServer()
	alice := registerCustomer(t, srv, "alice", "alice@example.com")
	bob := registerCustomer(t, srv, "bob", "bob@example.com")

	// Create (first one becomes the default).
	rec := do(t, srv, http.MethodPost, "/api/v2/addresses", alice,
		`{"label":"Home","address":"1 Main St","city":"Istanbul","latitude":41.0,"longitude":29.0}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create = %d (%s)", rec.Code, rec.Body)
	}
	var home map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &home)
	if home["is_default"] != true {
		t.Error("first address should be the default")
	}
	homeID := int(home["id"].(float64))

	// Validation -> 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/addresses", alice, `{"label":"","address":"x"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("missing label = %d, want 400", rec.Code)
	}

	// Second address as default takes over.
	rec = do(t, srv, http.MethodPost, "/api/v2/addresses", alice,
		`{"label":"Work","address":"2 Office Rd","is_default":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create work = %d (%s)", rec.Code, rec.Body)
	}

	rec = do(t, srv, http.MethodGet, "/api/v2/addresses", alice, "")
	var list []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 2 || list[0]["label"] != "Work" || list[0]["is_default"] != true || list[1]["is_default"] != false {
		t.Errorf("list wrong: %v", list)
	}

	// Bob sees an empty book and cannot touch Alice's rows.
	rec = do(t, srv, http.MethodGet, "/api/v2/addresses", bob, "")
	var bobList []map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &bobList)
	if len(bobList) != 0 {
		t.Errorf("bob's list = %v, want empty", bobList)
	}
	path := fmt.Sprintf("/api/v2/addresses/%d", homeID)
	if rec := do(t, srv, http.MethodPut, path, bob, `{"label":"Hacked","address":"x"}`); rec.Code != http.StatusForbidden {
		t.Errorf("foreign update = %d, want 403", rec.Code)
	}
	if rec := do(t, srv, http.MethodDelete, path, bob, ""); rec.Code != http.StatusForbidden {
		t.Errorf("foreign delete = %d, want 403", rec.Code)
	}

	// Owner update + delete.
	rec = do(t, srv, http.MethodPut, path, alice, `{"label":"Home 2","address":"3 New St"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update = %d (%s)", rec.Code, rec.Body)
	}
	if rec := do(t, srv, http.MethodDelete, path, alice, ""); rec.Code != http.StatusNoContent {
		t.Errorf("delete = %d, want 204", rec.Code)
	}
	if rec := do(t, srv, http.MethodDelete, path, alice, ""); rec.Code != http.StatusNotFound {
		t.Errorf("double delete = %d, want 404", rec.Code)
	}
}

// Ordering with address_id: the saved address fills the delivery fields; a
// foreign address is forbidden.
func TestAddressHTTP_OrderWithSavedAddress(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chef", "chef@example.com")
	_ = chefToken
	cust := registerCustomer(t, srv, "cust", "cust@example.com")
	thief := registerCustomer(t, srv, "thief", "thief@example.com")

	rec := do(t, srv, http.MethodPost, "/api/v2/addresses", cust,
		`{"label":"Home","address":"5 Saved St","city":"Istanbul"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create address = %d (%s)", rec.Code, rec.Body)
	}
	var addr map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &addr)
	addrID := int(addr["id"].(float64))

	body := fmt.Sprintf(`{"address_id":%d,"payment_method":"cash","items":[{"menu_item_id":%d,"quantity":1}]}`, addrID, itemID)
	rec = do(t, srv, http.MethodPost, "/api/v2/orders", cust, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("order with saved address = %d (%s)", rec.Code, rec.Body)
	}
	var order map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order["delivery_address"] != "5 Saved St" {
		t.Errorf("delivery_address = %v, want the snapshot", order["delivery_address"])
	}

	// Someone else's address id -> 403.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", thief, body); rec.Code != http.StatusForbidden {
		t.Errorf("foreign address order = %d, want 403", rec.Code)
	}
}
