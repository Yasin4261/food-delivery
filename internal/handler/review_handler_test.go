package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakeReviewRepo is an in-memory domain.ReviewRepository for HTTP tests
// (no aggregate recompute; that is covered by the repository integration tests).
type fakeReviewRepo struct {
	reviews []*domain.Review
	nextID  int
}

func newFakeReviewRepo() *fakeReviewRepo { return &fakeReviewRepo{nextID: 1} }

func sameTarget(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (f *fakeReviewRepo) Create(_ context.Context, rv *domain.Review) error {
	for _, ex := range f.reviews {
		if ex.UserID == rv.UserID && ex.OrderID == rv.OrderID &&
			sameTarget(ex.ChefID, rv.ChefID) && sameTarget(ex.MenuItemID, rv.MenuItemID) {
			return domain.ErrReviewExists
		}
	}
	rv.ID = f.nextID
	f.nextID++
	cp := *rv
	f.reviews = append(f.reviews, &cp)
	return nil
}
func (f *fakeReviewRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Review, int, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.ChefID != nil && *rv.ChefID == chefID {
			out = append(out, rv)
		}
	}
	return out, len(out), nil
}
func (f *fakeReviewRepo) ListByMenuItem(_ context.Context, menuItemID, limit, offset int) ([]*domain.Review, int, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.MenuItemID != nil && *rv.MenuItemID == menuItemID {
			out = append(out, rv)
		}
	}
	return out, len(out), nil
}

// placeAndDeliverOrder places an order for the customer containing the chef's
// item, then advances it to delivered via the chef. Returns the order id.
func placeAndDeliverOrder(t *testing.T, srv http.Handler, chefToken, customerToken string, itemID int) int {
	t.Helper()
	body := `{"delivery_address":"123 St","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	rec := do(t, srv, http.MethodPost, "/api/v2/orders", customerToken, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("place order = %d (%s)", rec.Code, rec.Body)
	}
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	for _, action := range []string{"confirm", "preparing", "ready", "delivering", "delivered"} {
		if rec := do(t, srv, http.MethodPost, "/api/v2/chef/orders/"+itoa(order.ID)+"/status", chefToken, `{"action":"`+action+`"}`); rec.Code != http.StatusOK {
			t.Fatalf("advance %q = %d (%s)", action, rec.Code, rec.Body)
		}
	}
	return order.ID
}

func TestReview_Flow(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeAndDeliverOrder(t, srv, chefToken, customer, itemID)

	// No token -> 401.
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", "", `{}`); rec.Code != http.StatusUnauthorized {
		t.Errorf("review without token = %d, want 401", rec.Code)
	}

	// Chef review, then product review.
	chefReview := `{"order_id":` + itoa(orderID) + `,"chef_id":1,"rating":5,"comment":"great"}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer, chefReview); rec.Code != http.StatusCreated {
		t.Fatalf("chef review = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	productReview := `{"order_id":` + itoa(orderID) + `,"menu_item_id":` + itoa(itemID) + `,"rating":4}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer, productReview); rec.Code != http.StatusCreated {
		t.Fatalf("product review = %d, want 201 (%s)", rec.Code, rec.Body)
	}

	// Duplicate chef review -> 409.
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer, chefReview); rec.Code != http.StatusConflict {
		t.Errorf("duplicate review = %d, want 409", rec.Code)
	}

	// Public reads.
	rec := do(t, srv, http.MethodGet, "/api/v2/chefs/1/reviews", "", "")
	if p := decodePage[domain.Review](t, rec.Body.Bytes()); rec.Code != http.StatusOK || len(p.Data) != 1 || p.Total != 1 {
		t.Errorf("chef reviews = %d/%+v, want 200 with one", rec.Code, p)
	}
	rec = do(t, srv, http.MethodGet, "/api/v2/menu-items/"+itoa(itemID)+"/reviews", "", "")
	if p := decodePage[domain.Review](t, rec.Body.Bytes()); rec.Code != http.StatusOK || len(p.Data) != 1 {
		t.Errorf("item reviews = %d/%+v, want 200 with one", rec.Code, p)
	}
}

func TestReview_Guards(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeAndDeliverOrder(t, srv, chefToken, customer, itemID)

	// Another customer cannot review someone else's order -> 403.
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", other,
		`{"order_id":`+itoa(orderID)+`,"chef_id":1,"rating":5}`); rec.Code != http.StatusForbidden {
		t.Errorf("non-owner review = %d, want 403", rec.Code)
	}

	// Bad rating -> 400.
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer,
		`{"order_id":`+itoa(orderID)+`,"chef_id":1,"rating":9}`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad rating = %d, want 400", rec.Code)
	}

	// Reviewing a chef not in the order -> 422.
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer,
		`{"order_id":`+itoa(orderID)+`,"chef_id":999,"rating":5}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("foreign chef review = %d, want 422", rec.Code)
	}

	// A pending (undelivered) order cannot be reviewed -> 422.
	body := `{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer, body)
	var pending domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &pending)
	if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer,
		`{"order_id":`+itoa(pending.ID)+`,"chef_id":1,"rating":5}`); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("pending order review = %d, want 422", rec.Code)
	}
}

func (f *fakeReviewRepo) ListByUserOrder(_ context.Context, userID, orderID int) ([]*domain.Review, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.UserID == userID && rv.OrderID == orderID {
			cp := *rv
			out = append(out, &cp)
		}
	}
	return out, nil
}

// The rating history endpoint: the caller sees their own reviews for an
// order; other users get an empty list, anonymous callers a 401.
func TestReview_HistoryForOrder(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeAndDeliverOrder(t, srv, chefToken, customer, itemID)
	path := "/api/v2/orders/" + itoa(orderID) + "/reviews"

	if rec := do(t, srv, http.MethodGet, path, "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("history without token = %d, want 401", rec.Code)
	}

	// Before rating: empty history.
	rec := do(t, srv, http.MethodGet, path, customer, "")
	var history []domain.Review
	_ = json.Unmarshal(rec.Body.Bytes(), &history)
	if rec.Code != http.StatusOK || len(history) != 0 {
		t.Fatalf("pre-rating history = %d/%d reviews, want 200/0", rec.Code, len(history))
	}

	// Rate the chef and the dish, then the history shows both.
	for _, body := range []string{
		`{"order_id":` + itoa(orderID) + `,"chef_id":1,"rating":5,"comment":"great"}`,
		`{"order_id":` + itoa(orderID) + `,"menu_item_id":` + itoa(itemID) + `,"rating":4}`,
	} {
		if rec := do(t, srv, http.MethodPost, "/api/v2/reviews", customer, body); rec.Code != http.StatusCreated {
			t.Fatalf("review = %d (%s)", rec.Code, rec.Body)
		}
	}
	rec = do(t, srv, http.MethodGet, path, customer, "")
	_ = json.Unmarshal(rec.Body.Bytes(), &history)
	if len(history) != 2 {
		t.Fatalf("history = %d reviews, want 2", len(history))
	}
	if history[0].ChefID == nil || *history[0].ChefID != 1 || history[0].Rating != 5 {
		t.Errorf("chef review wrong in history: %+v", history[0])
	}

	// Another user asking about this order sees nothing (no leak, no 404
	// probe signal beyond emptiness).
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	rec = do(t, srv, http.MethodGet, path, other, "")
	_ = json.Unmarshal(rec.Body.Bytes(), &history)
	if rec.Code != http.StatusOK || len(history) != 0 {
		t.Errorf("foreign history = %d/%d, want 200/0", rec.Code, len(history))
	}
}
