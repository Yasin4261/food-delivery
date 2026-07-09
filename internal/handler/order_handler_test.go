package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeOrderRepo is an in-memory domain.OrderRepository for HTTP tests.
type fakeOrderRepo struct {
	orders map[int]*domain.Order
	nextID int
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: map[int]*domain.Order{}, nextID: 1}
}

func cloneOrder(o *domain.Order, chefID int) *domain.Order {
	cp := *o
	cp.Items = make([]*domain.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		if chefID > 0 && it.ChefID != chefID {
			continue
		}
		ic := *it
		cp.Items = append(cp.Items, &ic)
	}
	return &cp
}

func (f *fakeOrderRepo) Create(_ context.Context, o *domain.Order) error {
	o.ID = f.nextID
	f.nextID++
	for i, it := range o.Items {
		it.ID = i + 1
		it.OrderID = o.ID
	}
	f.orders[o.ID] = cloneOrder(o, 0)
	return nil
}
func (f *fakeOrderRepo) FindByID(_ context.Context, id int) (*domain.Order, error) {
	if o, ok := f.orders[id]; ok {
		return cloneOrder(o, 0), nil
	}
	return nil, domain.ErrOrderNotFound
}
func (f *fakeOrderRepo) ListByUser(_ context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	out := make([]*domain.Order, 0)
	for _, o := range f.orders {
		if o.UserID == userID {
			out = append(out, cloneOrder(o, 0))
		}
	}
	return out, len(out), nil
}
func (f *fakeOrderRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Order, int, error) {
	out := make([]*domain.Order, 0)
	for _, o := range f.orders {
		if o.HasChef(chefID) {
			out = append(out, cloneOrder(o, chefID))
		}
	}
	return out, len(out), nil
}
func (f *fakeOrderRepo) UpdateStatus(_ context.Context, o *domain.Order) error {
	stored, ok := f.orders[o.ID]
	if !ok {
		return domain.ErrOrderNotFound
	}
	stored.Status = o.Status
	stored.PaymentStatus = o.PaymentStatus
	stored.ActualDeliveryTime = o.ActualDeliveryTime
	stored.CancelledAt = o.CancelledAt
	return nil
}

// seedChefWithItem registers a chef, opens a profile and publishes a menu with
// one limited-stock dish. It returns the chef's token and the dish id.
func seedChefWithItem(t *testing.T, srv http.Handler, username, email string) (token string, itemID int) {
	t.Helper()
	token = registerAndToken(t, srv, username, email)
	createChefProfile(t, srv, token)
	if rec := do(t, srv, http.MethodPost, "/api/v2/menus", token, `{"name":"Dinner"}`); rec.Code != http.StatusCreated {
		t.Fatalf("seed menu = %d (%s)", rec.Code, rec.Body)
	}
	rec := do(t, srv, http.MethodPost, "/api/v2/menu-items", token,
		`{"menu_id":1,"name":"Soup","price":5,"available_quantity":10}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("seed item = %d (%s)", rec.Code, rec.Body)
	}
	var item domain.MenuItem
	_ = json.Unmarshal(rec.Body.Bytes(), &item)
	return token, item.ID
}

func TestOrder_FullLifecycle(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Place an order for 2 units.
	body := `{"delivery_address":"123 St","payment_method":"cash","items":[{"menu_item_id":` +
		itoa(itemID) + `,"quantity":2}]}`
	rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("place order = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order.ID == 0 || order.Status != domain.OrderStatusPending {
		t.Fatalf("unexpected order: %+v", order)
	}
	if order.TotalPrice != 10 || len(order.Items) != 1 || order.Items[0].ChefID != 1 {
		t.Errorf("order totals/items wrong: %+v", order)
	}

	// Stock was decremented 10 -> 8.
	rec = do(t, srv, http.MethodGet, "/api/v2/menus/1/items", "", "")
	items := decodePage[domain.MenuItem](t, rec.Body.Bytes()).Data
	if len(items) != 1 || items[0].AvailableQuantity == nil || *items[0].AvailableQuantity != 8 {
		t.Errorf("stock not decremented: %+v", items)
	}

	// Customer can read it; a different customer cannot.
	if rec := do(t, srv, http.MethodGet, "/api/v2/orders/1", customer, ""); rec.Code != http.StatusOK {
		t.Errorf("owner get order = %d, want 200", rec.Code)
	}
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	if rec := do(t, srv, http.MethodGet, "/api/v2/orders/1", other, ""); rec.Code != http.StatusForbidden {
		t.Errorf("non-owner get order = %d, want 403", rec.Code)
	}

	// Chef sees the order scoped to their items and advances it.
	rec = do(t, srv, http.MethodGet, "/api/v2/chef/orders", chefToken, "")
	chefOrders := decodePage[domain.Order](t, rec.Body.Bytes()).Data
	if rec.Code != http.StatusOK || len(chefOrders) != 1 || len(chefOrders[0].Items) != 1 {
		t.Fatalf("chef orders = %d/%d, want 200/1", rec.Code, len(chefOrders))
	}

	for _, action := range []string{"confirm", "preparing", "ready", "delivering", "delivered"} {
		rec := do(t, srv, http.MethodPost, "/api/v2/chef/orders/1/status", chefToken, `{"action":"`+action+`"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("action %q = %d, want 200 (%s)", action, rec.Code, rec.Body)
		}
	}

	// Delivered cash order settles to paid (counts toward chef earnings).
	rec = do(t, srv, http.MethodGet, "/api/v2/orders/1", customer, "")
	var delivered domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &delivered)
	if delivered.PaymentStatus != "paid" {
		t.Errorf("delivered cash payment_status = %q, want paid", delivered.PaymentStatus)
	}

	// An illegal transition (confirm after delivered) is rejected.
	if rec := do(t, srv, http.MethodPost, "/api/v2/chef/orders/1/status", chefToken, `{"action":"confirm"}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("confirm after delivered = %d, want 422", rec.Code)
	}
	// Cancelling a delivered order is rejected.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/1/cancel", customer, ""); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("cancel delivered = %d, want 422", rec.Code)
	}
}

func TestOrder_CustomerCancelWhilePending(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	body := `{"delivery_address":"123 St","payment_method":"card","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer, body); rec.Code != http.StatusCreated {
		t.Fatalf("place = %d (%s)", rec.Code, rec.Body)
	}
	rec := do(t, srv, http.MethodPost, "/api/v2/orders/1/cancel", customer, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order.Status != domain.OrderStatusCancelled {
		t.Errorf("status = %q, want cancelled", order.Status)
	}
}

func TestOrder_GuardsAndValidation(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// No token -> 401.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", "", `{}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("place without token = %d, want 401", rec.Code)
	}
	// Customer cannot use chef-only endpoints -> 403.
	if rec := do(t, srv, http.MethodGet, "/api/v2/chef/orders", customer, ""); rec.Code != http.StatusForbidden {
		t.Errorf("customer chef orders = %d, want 403", rec.Code)
	}

	// Empty cart -> 422.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer,
		`{"delivery_address":"x","payment_method":"cash","items":[]}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty cart = %d, want 422", rec.Code)
	}
	// Bad payment method -> 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer,
		`{"delivery_address":"x","payment_method":"bitcoin","items":[{"menu_item_id":`+itoa(itemID)+`,"quantity":1}]}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad payment method = %d, want 400", rec.Code)
	}
	// Ordering more than in stock -> 422.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer,
		`{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":`+itoa(itemID)+`,"quantity":999}]}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("over-stock order = %d, want 422", rec.Code)
	}
}

func TestOrder_ChefCannotAdvanceUnrelatedOrder(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	body := `{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer, body); rec.Code != http.StatusCreated {
		t.Fatalf("place = %d (%s)", rec.Code, rec.Body)
	}

	// A second chef (no items in the order) cannot advance it.
	chefB := registerAndToken(t, srv, "chefb", "chefb@example.com")
	createChefProfile(t, srv, chefB)
	if rec := do(t, srv, http.MethodPost, "/api/v2/chef/orders/1/status", chefB, `{"action":"confirm"}`); rec.Code != http.StatusForbidden {
		t.Errorf("unrelated chef advance = %d, want 403", rec.Code)
	}
}

// itoa is a tiny local int-to-string to keep request bodies readable.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
