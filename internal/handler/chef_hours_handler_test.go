package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// closedWeekday returns a weekday guaranteed not to be "now" in the test
// fixture's evaluation zone (UTC): three days from today.
func closedWeekday() int {
	return (int(time.Now().UTC().Weekday()) + 3) % 7
}

// fakeChefHoursRepo is an in-memory domain.ChefHoursRepository for HTTP tests.
type fakeChefHoursRepo struct {
	byChef map[int][]*domain.ChefHours
}

func newFakeChefHoursRepo() *fakeChefHoursRepo {
	return &fakeChefHoursRepo{byChef: map[int][]*domain.ChefHours{}}
}

func (f *fakeChefHoursRepo) ReplaceForChef(_ context.Context, chefID int, hours []*domain.ChefHours) error {
	cp := make([]*domain.ChefHours, 0, len(hours))
	for _, h := range hours {
		hc := *h
		hc.ChefID = chefID
		cp = append(cp, &hc)
	}
	f.byChef[chefID] = cp
	return nil
}

func (f *fakeChefHoursRepo) ListByChef(_ context.Context, chefID int) ([]*domain.ChefHours, error) {
	out := make([]*domain.ChefHours, 0)
	for _, h := range f.byChef[chefID] {
		hc := *h
		out = append(out, &hc)
	}
	return out, nil
}

func (f *fakeChefHoursRepo) ListByChefs(_ context.Context, chefIDs []int) (map[int][]*domain.ChefHours, error) {
	out := map[int][]*domain.ChefHours{}
	for _, id := range chefIDs {
		if hours := f.byChef[id]; len(hours) > 0 {
			list, _ := f.ListByChef(context.Background(), id)
			out[id] = list
		}
	}
	return out, nil
}

func TestChefHoursHTTP_EditAndRead(t *testing.T) {
	srv := newTestServer()
	chefToken, _ := seedChefWithItem(t, srv, "chefa", "chefa@example.com")

	// Anonymous edit -> 401; customer -> 403.
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", "", `[]`); rec.Code != http.StatusUnauthorized {
		t.Errorf("anonymous = %d, want 401", rec.Code)
	}
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", cust, `[]`); rec.Code != http.StatusForbidden {
		t.Errorf("customer = %d, want 403", rec.Code)
	}

	// Bad clock string -> 400; bad weekday -> 400.
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken,
		`[{"weekday":1,"opens":"9am","closes":"17:00"}]`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad clock = %d, want 400", rec.Code)
	}
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken,
		`[{"weekday":9,"opens":"09:00","closes":"17:00"}]`); rec.Code != http.StatusBadRequest {
		t.Errorf("bad weekday = %d, want 400", rec.Code)
	}

	// Set a schedule; the public read returns it in HH:MM form.
	rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken,
		`[{"weekday":1,"opens":"09:00","closes":"17:00"},{"weekday":5,"opens":"18:00","closes":"02:00"}]`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set hours = %d (%s)", rec.Code, rec.Body)
	}
	pub := do(t, srv, http.MethodGet, "/api/v2/chefs/1/hours", "", "")
	var hours []map[string]any
	_ = json.Unmarshal(pub.Body.Bytes(), &hours)
	if pub.Code != http.StatusOK || len(hours) != 2 || hours[0]["opens"] != "09:00" || hours[1]["closes"] != "02:00" {
		t.Fatalf("public hours = %d/%v", pub.Code, hours)
	}

	// Full replace: an empty list clears the schedule (always open again).
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken, `[]`); rec.Code != http.StatusOK {
		t.Fatalf("clear hours = %d (%s)", rec.Code, rec.Body)
	}
	pub = do(t, srv, http.MethodGet, "/api/v2/chefs/1/hours", "", "")
	_ = json.Unmarshal(pub.Body.Bytes(), &hours)
	if len(hours) != 0 {
		t.Errorf("hours after clear = %v, want empty", hours)
	}
}

// A closed chef cannot receive orders (422); clearing the schedule reopens.
func TestChefHoursHTTP_OrderingGate(t *testing.T) {
	srv := newTestServer()
	chefToken, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	cust := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// A 1-minute window on some other weekday: effectively always closed.
	// (00:00-00:01 next weekday — "now" can only match one weekday at a time,
	// so pick two windows that cannot both be wrong... simplest: a window on
	// weekday (today+3)%7.)
	rec := do(t, srv, http.MethodGet, "/api/v2/chefs/1/hours", "", "")
	_ = rec
	closedBody := `[{"weekday":` + itoa(closedWeekday()) + `,"opens":"03:00","closes":"03:01"}]`
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken, closedBody); rec.Code != http.StatusOK {
		t.Fatalf("set closed hours = %d (%s)", rec.Code, rec.Body)
	}

	order := `{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", cust, order); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("order to closed chef = %d, want 422 (%s)", rec.Code, rec.Body)
	}

	// The browse list marks the chef closed.
	list := do(t, srv, http.MethodGet, "/api/v2/chefs", "", "")
	chefs := decodePage[domain.Chef](t, list.Body.Bytes()).Data
	if len(chefs) != 1 || chefs[0].IsOpenNow == nil || *chefs[0].IsOpenNow {
		t.Errorf("browse is_open_now = %+v, want false", chefs[0].IsOpenNow)
	}

	// Clearing the schedule reopens the kitchen.
	if rec := do(t, srv, http.MethodPut, "/api/v2/chefs/me/hours", chefToken, `[]`); rec.Code != http.StatusOK {
		t.Fatalf("clear hours = %d", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders", cust, order); rec.Code != http.StatusCreated {
		t.Errorf("order after clearing hours = %d, want 201 (%s)", rec.Code, rec.Body)
	}
}
