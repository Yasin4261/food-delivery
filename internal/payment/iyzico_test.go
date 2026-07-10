package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// newTestGateway points an Iyzico gateway at a local httptest server with a
// fixed random key so signatures are deterministic.
func newTestGateway(t *testing.T, handler http.HandlerFunc) (*Iyzico, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	g := NewIyzico("api-key", "secret-key", srv.URL)
	g.randomKey = func() string { return "fixed-random" }
	return g, srv
}

func cardOrder() (*domain.Order, *domain.User) {
	method := domain.PaymentMethodCard
	order := &domain.Order{
		ID:              42,
		UserID:          7,
		Subtotal:        14.5,
		TotalPrice:      14.5,
		PaymentMethod:   &method,
		PaymentStatus:   domain.PaymentStatusPending,
		Status:          domain.OrderStatusPending,
		DeliveryAddress: "1 Test St",
		Items: []*domain.OrderItem{
			{MenuItemID: 2, ItemName: "Plot Pilaf", Quantity: 2, UnitPrice: 7.25, Subtotal: 14.5, ChefID: 3},
		},
	}
	buyer := &domain.User{ID: 7, Username: "cust", Email: "cust@example.com"}
	return order, buyer
}

func TestIyzico_AuthorizationHeader(t *testing.T) {
	var gotAuth, gotRnd string
	var gotBody []byte
	g, _ := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotRnd = r.Header.Get("x-iyzi-rnd")
		gotBody, _ = io.ReadAll(r.Body)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "token": "t", "paymentPageUrl": "u"})
	})

	order, buyer := cardOrder()
	if _, err := g.InitiateCheckout(context.Background(), order, buyer, "http://cb"); err != nil {
		t.Fatalf("initiate: %v", err)
	}

	if gotRnd != "fixed-random" {
		t.Errorf("x-iyzi-rnd = %q", gotRnd)
	}
	// Recompute the IYZWSv2 signature over what was actually sent.
	mac := hmac.New(sha256.New, []byte("secret-key"))
	mac.Write([]byte("fixed-random" + pathCheckoutInit + string(gotBody)))
	wantSig := hex.EncodeToString(mac.Sum(nil))
	wantAuth := "IYZWSv2 " + base64.StdEncoding.EncodeToString(
		[]byte("apiKey:api-key&randomKey:fixed-random&signature:"+wantSig))
	if gotAuth != wantAuth {
		t.Errorf("authorization header mismatch\n got %s\nwant %s", gotAuth, wantAuth)
	}
}

func TestIyzico_InitiateCheckout(t *testing.T) {
	g, _ := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pathCheckoutInit {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req checkoutInitRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.ConversationID != "42" || req.BasketID != "42" {
			t.Errorf("conversation/basket = %s/%s, want 42", req.ConversationID, req.BasketID)
		}
		if req.Price != "14.50" || req.PaidPrice != "14.50" || req.Currency != "TRY" {
			t.Errorf("price fields wrong: %+v", req)
		}
		if len(req.BasketItems) != 1 || req.BasketItems[0].Price != "14.50" || req.BasketItems[0].Name != "Plot Pilaf" {
			t.Errorf("basket items wrong: %+v", req.BasketItems)
		}
		if req.CallbackURL != "http://cb" {
			t.Errorf("callback = %s", req.CallbackURL)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "success", "token": "tok-1", "paymentPageUrl": "https://pay.example/x",
		})
	})

	order, buyer := cardOrder()
	cs, err := g.InitiateCheckout(context.Background(), order, buyer, "http://cb")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	if cs.Token != "tok-1" || cs.PaymentPageURL != "https://pay.example/x" {
		t.Errorf("session = %+v", cs)
	}
}

func TestIyzico_InitiateCheckout_GatewayError(t *testing.T) {
	g, _ := newTestGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "failure", "errorMessage": "invalid signature"})
	})
	order, buyer := cardOrder()
	if _, err := g.InitiateCheckout(context.Background(), order, buyer, "cb"); err == nil || !strings.Contains(err.Error(), "invalid signature") {
		t.Errorf("err = %v, want gateway error surfaced", err)
	}
}

func TestIyzico_VerifyCheckout(t *testing.T) {
	g, _ := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pathCheckoutRetrieve {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["token"] != "tok-1" {
			t.Errorf("token = %s", req["token"])
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "success", "paymentStatus": "SUCCESS", "paymentId": "pay-9",
		})
	})

	res, err := g.VerifyCheckout(context.Background(), "tok-1")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !res.Paid || res.PaymentID != "pay-9" || res.Token != "tok-1" {
		t.Errorf("result = %+v", res)
	}
}

func TestIyzico_VerifyCheckout_Failure(t *testing.T) {
	g, _ := newTestGateway(t, func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success", "paymentStatus": "FAILURE"})
	})
	res, err := g.VerifyCheckout(context.Background(), "tok-2")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.Paid {
		t.Error("FAILURE paymentStatus must not be Paid")
	}
}

func TestIyzico_Refund(t *testing.T) {
	g, _ := newTestGateway(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != pathCancel {
			t.Errorf("path = %s", r.URL.Path)
		}
		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["paymentId"] != "pay-9" {
			t.Errorf("paymentId = %s", req["paymentId"])
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
	if err := g.Refund(context.Background(), "pay-9"); err != nil {
		t.Errorf("refund: %v", err)
	}
}

func TestMock_RoundTrip(t *testing.T) {
	m := NewMock("http://app.test/")
	order, buyer := cardOrder()

	cs, err := m.InitiateCheckout(context.Background(), order, buyer, "cb")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	if !strings.HasPrefix(cs.PaymentPageURL, "http://app.test/mock-pay?token=mock-42-") {
		t.Errorf("payment page = %s", cs.PaymentPageURL)
	}

	ok, _ := m.VerifyCheckout(context.Background(), cs.Token)
	if !ok.Paid || ok.Token != cs.Token || ok.PaymentID == "" {
		t.Errorf("success verify = %+v", ok)
	}

	fail, _ := m.VerifyCheckout(context.Background(), cs.Token+":fail")
	if fail.Paid || fail.Token != cs.Token {
		t.Errorf("fail verify = %+v (token must be stripped)", fail)
	}

	if err := m.Refund(context.Background(), "x"); err != nil {
		t.Errorf("mock refund: %v", err)
	}
}
