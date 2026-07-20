package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// fakePaymentSessionRepo is an in-memory domain.PaymentSessionRepository.
type fakePaymentSessionRepo struct {
	mu       sync.Mutex
	sessions map[int]*domain.PaymentSession
	nextID   int
}

func newFakePaymentSessionRepo() *fakePaymentSessionRepo {
	return &fakePaymentSessionRepo{sessions: map[int]*domain.PaymentSession{}, nextID: 1}
}

func (f *fakePaymentSessionRepo) Create(_ context.Context, s *domain.PaymentSession) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s.ID = f.nextID
	f.nextID++
	if s.Status == "" {
		s.Status = domain.PaymentSessionInitiated
	}
	s.CreatedAt = time.Now()
	cp := *s
	f.sessions[s.ID] = &cp
	return nil
}
func (f *fakePaymentSessionRepo) FindByToken(_ context.Context, token string) (*domain.PaymentSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.Token == token {
			cp := *s
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentSessionNotFound
}
func (f *fakePaymentSessionRepo) FindPaidByOrder(_ context.Context, orderID int) (*domain.PaymentSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.sessions {
		if s.OrderID == orderID && s.Status == domain.PaymentSessionPaid {
			cp := *s
			return &cp, nil
		}
	}
	return nil, domain.ErrPaymentSessionNotFound
}
func (f *fakePaymentSessionRepo) UpdateStatus(_ context.Context, id int, status string, paymentID *string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.sessions[id]
	if !ok {
		return domain.ErrPaymentSessionNotFound
	}
	s.Status = status
	if paymentID != nil {
		s.PaymentID = paymentID
	}
	return nil
}

// fakePaymentMethodRepo is an in-memory domain.PaymentMethodRepository for the
// saved-card HTTP tests.
type fakePaymentMethodRepo struct {
	mu     sync.Mutex
	cards  []*domain.SavedCard
	nextID int
}

func newFakePaymentMethodRepo() *fakePaymentMethodRepo {
	return &fakePaymentMethodRepo{nextID: 1}
}

func (f *fakePaymentMethodRepo) Add(_ context.Context, c *domain.SavedCard) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, e := range f.cards {
		if e.UserID == c.UserID && e.CardToken == c.CardToken {
			c.ID, c.CreatedAt = e.ID, e.CreatedAt
			return nil
		}
	}
	c.ID = f.nextID
	f.nextID++
	c.CreatedAt = time.Now()
	cp := *c
	f.cards = append(f.cards, &cp)
	return nil
}
func (f *fakePaymentMethodRepo) ListByUser(_ context.Context, userID int) ([]*domain.SavedCard, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []*domain.SavedCard
	for i := len(f.cards) - 1; i >= 0; i-- {
		if f.cards[i].UserID == userID {
			cp := *f.cards[i]
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakePaymentMethodRepo) FindByToken(_ context.Context, userID int, token string) (*domain.SavedCard, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.cards {
		if c.UserID == userID && c.CardToken == token {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrCardNotFound
}
func (f *fakePaymentMethodRepo) CardUserKey(_ context.Context, userID int) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.cards) - 1; i >= 0; i-- {
		if f.cards[i].UserID == userID {
			return f.cards[i].CardUserKey, nil
		}
	}
	return "", nil
}
func (f *fakePaymentMethodRepo) Delete(_ context.Context, userID int, token string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, c := range f.cards {
		if c.UserID == userID && c.CardToken == token {
			f.cards = append(f.cards[:i], f.cards[i+1:]...)
			return nil
		}
	}
	return domain.ErrCardNotFound
}

// placeCardOrder places a card order for the customer and returns its id.
func placeCardOrder(t *testing.T, srv http.Handler, customerToken string, itemID int) int {
	t.Helper()
	body := `{"delivery_address":"1 Pay St","payment_method":"card","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	rec := do(t, srv, http.MethodPost, "/api/v2/orders", customerToken, body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("place card order = %d (%s)", rec.Code, rec.Body)
	}
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	return order.ID
}

// payToken starts the checkout and extracts the mock token from the returned
// payment page URL.
func payToken(t *testing.T, srv http.Handler, customerToken string, orderID int) string {
	t.Helper()
	rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(orderID)+"/pay", customerToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("pay = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var resp struct {
		PaymentPageURL string `json:"payment_page_url"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	u, err := url.Parse(resp.PaymentPageURL)
	if err != nil || u.Query().Get("token") == "" {
		t.Fatalf("unexpected payment page url: %q", resp.PaymentPageURL)
	}
	if !strings.HasPrefix(resp.PaymentPageURL, "http://app.test/mock-pay?token=") {
		t.Fatalf("payment page should be the mock pay screen, got %q", resp.PaymentPageURL)
	}
	return u.Query().Get("token")
}

// callback posts the token form-encoded, exactly like the gateway's browser
// redirect does.
func callback(t *testing.T, srv http.Handler, token string) string {
	t.Helper()
	r := doForm(t, srv, "/api/v2/payments/callback", "token="+url.QueryEscape(token))
	if r.Code != http.StatusSeeOther {
		t.Fatalf("callback = %d, want 303 (%s)", r.Code, r.Body)
	}
	return r.Header().Get("Location")
}

func TestPayment_CardCheckoutFlow(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeCardOrder(t, srv, customer, itemID)

	// Success path: pay -> mock page token -> callback -> order paid.
	token := payToken(t, srv, customer, orderID)
	loc := callback(t, srv, token)
	if !strings.Contains(loc, "payment=success") || !strings.Contains(loc, "order="+itoa(orderID)) {
		t.Errorf("redirect = %q, want payment=success for order %d", loc, orderID)
	}

	rec := do(t, srv, http.MethodGet, "/api/v2/orders/"+itoa(orderID), customer, "")
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order.PaymentStatus != "paid" {
		t.Errorf("payment_status = %q, want paid", order.PaymentStatus)
	}

	// Replayed callback stays success (idempotent) and does not corrupt state.
	if loc := callback(t, srv, token); !strings.Contains(loc, "payment=success") {
		t.Errorf("replayed callback = %q, want success", loc)
	}

	// A paid order cannot be paid again.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(orderID)+"/pay", customer, ""); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("re-pay = %d, want 422", rec.Code)
	}
}

func TestPayment_FailedCheckout(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeCardOrder(t, srv, customer, itemID)

	token := payToken(t, srv, customer, orderID)
	loc := callback(t, srv, token+":fail")
	if !strings.Contains(loc, "payment=failed") {
		t.Errorf("redirect = %q, want payment=failed", loc)
	}

	rec := do(t, srv, http.MethodGet, "/api/v2/orders/"+itoa(orderID), customer, "")
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order.PaymentStatus != "pending" {
		t.Errorf("payment_status after failure = %q, want pending (retryable)", order.PaymentStatus)
	}

	// Retry after failure works.
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(orderID)+"/pay", customer, ""); rec.Code != http.StatusOK {
		t.Errorf("retry pay = %d, want 200", rec.Code)
	}
}

func TestPayment_Guards(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Cash orders cannot be paid online.
	body := `{"delivery_address":"x","payment_method":"cash","items":[{"menu_item_id":` + itoa(itemID) + `,"quantity":1}]}`
	rec := do(t, srv, http.MethodPost, "/api/v2/orders", customer, body)
	var cash domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &cash)
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(cash.ID)+"/pay", customer, ""); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("pay cash order = %d, want 422", rec.Code)
	}

	// Only the owner may pay.
	cardID := placeCardOrder(t, srv, customer, itemID)
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(cardID)+"/pay", other, ""); rec.Code != http.StatusForbidden {
		t.Errorf("non-owner pay = %d, want 403", rec.Code)
	}
	if rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(cardID)+"/pay", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("anonymous pay = %d, want 401", rec.Code)
	}

	// Unknown callback token -> error redirect, never a 500.
	loc := callback(t, srv, "no-such-token")
	if !strings.Contains(loc, "payment=error") {
		t.Errorf("unknown token redirect = %q, want payment=error", loc)
	}
}

func TestPayment_SavedCards(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")

	// Initially no saved cards.
	if cards := listCards(t, srv, customer); len(cards) != 0 {
		t.Fatalf("initial cards = %d, want 0", len(cards))
	}

	// Pay opting to save the card, then complete via the callback.
	orderID := placeCardOrder(t, srv, customer, itemID)
	rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(orderID)+"/pay", customer, `{"save_card":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("pay (save) = %d (%s)", rec.Code, rec.Body)
	}
	var payResp struct {
		PaymentPageURL string `json:"payment_page_url"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &payResp)
	u, _ := url.Parse(payResp.PaymentPageURL)
	callback(t, srv, u.Query().Get("token"))

	// The card is now saved (masked, tokenised — no PAN).
	cards := listCards(t, srv, customer)
	if len(cards) != 1 {
		t.Fatalf("after save cards = %d, want 1", len(cards))
	}
	if cards[0].CardToken == "" || cards[0].MaskedNumber == "" {
		t.Errorf("saved card missing token/mask: %+v", cards[0])
	}

	// Owner scoping: another customer sees none, and cannot delete this card.
	other := registerCustomerToken(t, srv, "other", "other@example.com")
	if c := listCards(t, srv, other); len(c) != 0 {
		t.Errorf("other user sees %d cards, want 0", len(c))
	}
	if rec := do(t, srv, http.MethodDelete, "/api/v2/payment-methods/"+cards[0].CardToken, other, ""); rec.Code != http.StatusNotFound {
		t.Errorf("foreign delete = %d, want 404", rec.Code)
	}

	// Owner deletes the card.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/payment-methods/"+cards[0].CardToken, customer, ""); rec.Code != http.StatusOK {
		t.Fatalf("delete own card = %d (%s)", rec.Code, rec.Body)
	}
	if c := listCards(t, srv, customer); len(c) != 0 {
		t.Errorf("after delete cards = %d, want 0", len(c))
	}
	// Deleting again is a 404.
	if rec := do(t, srv, http.MethodDelete, "/api/v2/payment-methods/"+cards[0].CardToken, customer, ""); rec.Code != http.StatusNotFound {
		t.Errorf("re-delete = %d, want 404", rec.Code)
	}

	// The endpoints require authentication.
	if rec := do(t, srv, http.MethodGet, "/api/v2/payment-methods", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("anon list = %d, want 401", rec.Code)
	}
}

// listCards fetches the caller's saved cards.
func listCards(t *testing.T, srv http.Handler, token string) []domain.SavedCard {
	t.Helper()
	rec := do(t, srv, http.MethodGet, "/api/v2/payment-methods", token, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("list cards = %d (%s)", rec.Code, rec.Body)
	}
	var resp struct {
		Data []domain.SavedCard `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode cards: %v", err)
	}
	return resp.Data
}

func TestPayment_RefundOnCancel(t *testing.T) {
	srv := newTestServer()
	_, itemID := seedChefWithItem(t, srv, "chefa", "chefa@example.com")
	customer := registerCustomerToken(t, srv, "cust", "cust@example.com")
	orderID := placeCardOrder(t, srv, customer, itemID)

	// Pay, then cancel while still pending -> refunded via the gateway.
	token := payToken(t, srv, customer, orderID)
	callback(t, srv, token)

	rec := do(t, srv, http.MethodPost, "/api/v2/orders/"+itoa(orderID)+"/cancel", customer, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel paid order = %d, want 200 (%s)", rec.Code, rec.Body)
	}
	var order domain.Order
	_ = json.Unmarshal(rec.Body.Bytes(), &order)
	if order.Status != "cancelled" || order.PaymentStatus != "refunded" {
		t.Errorf("after cancel: status=%q payment=%q, want cancelled/refunded", order.Status, order.PaymentStatus)
	}
}
