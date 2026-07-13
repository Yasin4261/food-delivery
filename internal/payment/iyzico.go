// Package payment holds the driven adapters for the domain.PaymentGateway
// port: iyzico (hosted Checkout Form) for real card charges and a mock that
// simulates the same dance in development.
package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// API paths (iyzico "Checkout Form" flow).
const (
	pathCheckoutInit     = "/payment/iyzipos/checkoutform/initialize/auth/ecom"
	pathCheckoutRetrieve = "/payment/iyzipos/checkoutform/auth/ecom/detail"
	pathCancel           = "/payment/cancel"
	pathRefundV2         = "/v2/payment/refund"
)

// Iyzico implements domain.PaymentGateway against the iyzico REST API using
// the IYZWSv2 (HMAC-SHA256) authorization scheme.
type Iyzico struct {
	apiKey    string
	secretKey string
	baseURL   string
	currency  string
	client    *http.Client
	// randomKey is injectable so signature tests are deterministic.
	randomKey func() string
}

// NewIyzico builds an iyzico gateway. baseURL selects sandbox vs production
// (e.g. https://sandbox-api.iyzipay.com).
func NewIyzico(apiKey, secretKey, baseURL string) *Iyzico {
	return &Iyzico{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		currency:  "TRY",
		client:    &http.Client{Timeout: 15 * time.Second},
		randomKey: func() string { return strconv.FormatInt(time.Now().UnixMilli(), 10) + "123456789" },
	}
}

// authorization computes the IYZWSv2 header for a request:
//
//	signature = hex(HMAC-SHA256(randomKey + uriPath + body, secretKey))
//	Authorization: IYZWSv2 base64("apiKey:<k>&randomKey:<r>&signature:<s>")
func (g *Iyzico) authorization(uriPath string, body []byte) (header, random string) {
	random = g.randomKey()
	mac := hmac.New(sha256.New, []byte(g.secretKey))
	mac.Write([]byte(random + uriPath + string(body)))
	signature := hex.EncodeToString(mac.Sum(nil))
	params := "apiKey:" + g.apiKey + "&randomKey:" + random + "&signature:" + signature
	return "IYZWSv2 " + base64.StdEncoding.EncodeToString([]byte(params)), random
}

// post sends an authenticated JSON request and decodes the response into out.
func (g *Iyzico) post(ctx context.Context, uriPath string, in, out any) error {
	body, err := json.Marshal(in)
	if err != nil {
		return fmt.Errorf("iyzico: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+uriPath, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("iyzico: build request: %w", err)
	}
	auth, random := g.authorization(uriPath, body)
	req.Header.Set("Authorization", auth)
	req.Header.Set("x-iyzi-rnd", random)
	req.Header.Set("Content-Type", "application/json")

	res, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("iyzico: %s: %w", uriPath, err)
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("iyzico: decode response: %w", err)
	}
	return nil
}

// --- request/response shapes (subset of the iyzico API) ---

type iyzicoBuyer struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Surname             string `json:"surname"`
	Email               string `json:"email"`
	IdentityNumber      string `json:"identityNumber"`
	RegistrationAddress string `json:"registrationAddress"`
	IP                  string `json:"ip"`
	City                string `json:"city"`
	Country             string `json:"country"`
}

type iyzicoAddress struct {
	ContactName string `json:"contactName"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Address     string `json:"address"`
}

type iyzicoBasketItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Category1 string `json:"category1"`
	ItemType  string `json:"itemType"`
	Price     string `json:"price"`
}

type checkoutInitRequest struct {
	Locale              string             `json:"locale"`
	ConversationID      string             `json:"conversationId"`
	Price               string             `json:"price"`
	PaidPrice           string             `json:"paidPrice"`
	Currency            string             `json:"currency"`
	BasketID            string             `json:"basketId"`
	PaymentGroup        string             `json:"paymentGroup"`
	CallbackURL         string             `json:"callbackUrl"`
	EnabledInstallments []int              `json:"enabledInstallments"`
	Buyer               iyzicoBuyer        `json:"buyer"`
	ShippingAddress     iyzicoAddress      `json:"shippingAddress"`
	BillingAddress      iyzicoAddress      `json:"billingAddress"`
	BasketItems         []iyzicoBasketItem `json:"basketItems"`
}

type checkoutInitResponse struct {
	Status         string `json:"status"`
	ErrorMessage   string `json:"errorMessage"`
	Token          string `json:"token"`
	PaymentPageURL string `json:"paymentPageUrl"`
}

type checkoutRetrieveResponse struct {
	Status        string `json:"status"`
	ErrorMessage  string `json:"errorMessage"`
	Token         string `json:"token"`
	PaymentStatus string `json:"paymentStatus"`
	PaymentID     string `json:"paymentId"`
}

type cancelResponse struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage"`
}

func money(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }

// InitiateCheckout opens a hosted Checkout Form session for the order.
func (g *Iyzico) InitiateCheckout(ctx context.Context, order *domain.Order, buyer *domain.User, callbackURL string) (*domain.CheckoutSession, error) {
	city := "Istanbul"
	if order.DeliveryCity != nil && *order.DeliveryCity != "" {
		city = *order.DeliveryCity
	}
	items := make([]iyzicoBasketItem, 0, len(order.Items))
	for _, it := range order.Items {
		items = append(items, iyzicoBasketItem{
			ID:        strconv.Itoa(it.MenuItemID),
			Name:      it.ItemName,
			Category1: "Food",
			ItemType:  "PHYSICAL",
			Price:     money(it.Subtotal),
		})
	}
	address := iyzicoAddress{ContactName: buyer.Username, City: city, Country: "Turkey", Address: order.DeliveryAddress}

	req := checkoutInitRequest{
		Locale:              "en",
		ConversationID:      strconv.Itoa(order.ID),
		Price:               money(order.Subtotal),
		PaidPrice:           money(order.TotalPrice),
		Currency:            g.currency,
		BasketID:            strconv.Itoa(order.ID),
		PaymentGroup:        "PRODUCT",
		CallbackURL:         callbackURL,
		EnabledInstallments: []int{1},
		Buyer: iyzicoBuyer{
			ID:      strconv.Itoa(buyer.ID),
			Name:    buyer.Username,
			Surname: buyer.Username,
			Email:   buyer.Email,
			// iyzico requires a national identity number; we do not collect
			// one, so the documented test value is sent. Collect a real TCKN
			// before a production launch.
			IdentityNumber:      "11111111111",
			RegistrationAddress: order.DeliveryAddress,
			IP:                  "127.0.0.1",
			City:                city,
			Country:             "Turkey",
		},
		ShippingAddress: address,
		BillingAddress:  address,
		BasketItems:     items,
	}

	var res checkoutInitResponse
	if err := g.post(ctx, pathCheckoutInit, req, &res); err != nil {
		return nil, err
	}
	if res.Status != "success" {
		return nil, fmt.Errorf("iyzico: checkout initialize failed: %s", res.ErrorMessage)
	}
	return &domain.CheckoutSession{Token: res.Token, PaymentPageURL: res.PaymentPageURL}, nil
}

// VerifyCheckout retrieves the checkout result for a callback token.
func (g *Iyzico) VerifyCheckout(ctx context.Context, token string) (*domain.PaymentResult, error) {
	req := map[string]string{"locale": "en", "token": token}
	var res checkoutRetrieveResponse
	if err := g.post(ctx, pathCheckoutRetrieve, req, &res); err != nil {
		return nil, err
	}
	if res.Status != "success" {
		return nil, fmt.Errorf("iyzico: checkout retrieve failed: %s", res.ErrorMessage)
	}
	return &domain.PaymentResult{
		Token:     token,
		Paid:      res.PaymentStatus == "SUCCESS",
		PaymentID: res.PaymentID,
	}, nil
}

// Refund cancels a captured payment in full.
func (g *Iyzico) Refund(ctx context.Context, paymentID string) error {
	req := map[string]string{"locale": "en", "paymentId": paymentID, "ip": "127.0.0.1"}
	var res cancelResponse
	if err := g.post(ctx, pathCancel, req, &res); err != nil {
		return err
	}
	if res.Status != "success" {
		return fmt.Errorf("iyzico: cancel failed: %s", res.ErrorMessage)
	}
	return nil
}

// RefundPartial returns part of a captured payment via iyzico's amount-based
// refund (v2 refund-by-paymentId) — a declined sub-order's slice of a
// multi-chef order.
func (g *Iyzico) RefundPartial(ctx context.Context, paymentID string, amount float64) error {
	req := map[string]string{
		"locale":    "en",
		"paymentId": paymentID,
		"price":     strconv.FormatFloat(amount, 'f', 2, 64),
		"ip":        "127.0.0.1",
	}
	var res cancelResponse
	if err := g.post(ctx, pathRefundV2, req, &res); err != nil {
		return err
	}
	if res.Status != "success" {
		return fmt.Errorf("iyzico: partial refund failed: %s", res.ErrorMessage)
	}
	return nil
}
