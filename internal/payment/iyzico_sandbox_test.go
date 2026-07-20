//go:build iyzico_sandbox

// Package payment's iyzico sandbox smoke test (#51). Unlike iyzico_test.go
// (which drives a local httptest server), this exercises the adapter against
// the REAL iyzico sandbox, so it needs merchant credentials and network access.
//
// It is gated behind the `iyzico_sandbox` build tag AND skips when the
// credentials are absent, so it never runs in the default suite or CI.
//
// Run it once you have a sandbox merchant account
// (https://sandbox-merchant.iyzipay.com):
//
//	IYZICO_API_KEY=sandbox-xxxx \
//	IYZICO_SECRET_KEY=sandbox-yyyy \
//	go test -tags=iyzico_sandbox -v ./internal/payment/
//
// or: make test-iyzico-sandbox
//
// What it can verify automatically:
//   - InitiateCheckout is accepted by the live API — this proves the IYZWSv2
//     HMAC signature, request body shape, currency and buyer fields are all
//     valid (the single most useful smoke for the adapter).
//   - VerifyCheckout resolves a fresh (unpaid) token without erroring.
//   - Refund / RefundPartial, IF you pass a captured sandbox paymentId (see
//     below) — a real capture needs a browser to complete the hosted 3-D
//     Secure form, which a Go test can't drive.
//
// Completing an actual payment is a MANUAL step: open the PaymentPageURL this
// test prints, pay with an iyzico test card (e.g. 5528790000000008, exp 12/30,
// CVC 123 — see iyzico's sandbox card list), then feed the resulting paymentId
// back via IYZICO_SANDBOX_PAYMENT_ID to exercise the refund paths.
package payment

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// sandboxGateway builds an Iyzico adapter from the environment, skipping the
// test when credentials are not configured.
func sandboxGateway(t *testing.T) *Iyzico {
	t.Helper()
	apiKey := os.Getenv("IYZICO_API_KEY")
	secret := os.Getenv("IYZICO_SECRET_KEY")
	if apiKey == "" || secret == "" {
		t.Skip("iyzico sandbox creds not set (IYZICO_API_KEY / IYZICO_SECRET_KEY); skipping")
	}
	baseURL := os.Getenv("IYZICO_BASE_URL")
	if baseURL == "" {
		baseURL = "https://sandbox-api.iyzipay.com"
	}
	return NewIyzico(apiKey, secret, baseURL)
}

// TestSandbox_InitiateCheckout verifies the live sandbox accepts an initialize
// request: a success status with a token + hosted payment page URL means our
// auth header, body shape and field formats are all valid.
func TestSandbox_InitiateCheckout(t *testing.T) {
	g := sandboxGateway(t)
	order, buyer := cardOrder()

	cs, err := g.InitiateCheckout(context.Background(), order, buyer,
		"https://example.test/api/v2/payments/callback", domain.CheckoutOptions{})
	if err != nil {
		t.Fatalf("initiate against sandbox: %v", err)
	}
	if cs.Token == "" || cs.PaymentPageURL == "" {
		t.Fatalf("sandbox returned empty session: %+v", cs)
	}
	t.Logf("hosted checkout ready — complete it in a browser with an iyzico test card:\n  %s", cs.PaymentPageURL)

	// VerifyCheckout on the fresh (unpaid) token must resolve without error;
	// it is simply not paid yet.
	res, err := g.VerifyCheckout(context.Background(), cs.Token)
	if err != nil {
		t.Fatalf("verify fresh token: %v", err)
	}
	if res.Paid {
		t.Errorf("a brand-new checkout token should not be paid yet: %+v", res)
	}
}

// TestSandbox_CardStorage checks that a checkout advertising a card wallet is
// accepted (the saved-card path, #67). Whether a card is actually returned
// depends on the buyer ticking "save" in the hosted form — a manual step — so
// this only asserts the request is accepted, then prints the page to complete.
func TestSandbox_CardStorage(t *testing.T) {
	g := sandboxGateway(t)
	order, buyer := cardOrder()

	cs, err := g.InitiateCheckout(context.Background(), order, buyer,
		"https://example.test/api/v2/payments/callback",
		domain.CheckoutOptions{RegisterCard: true})
	if err != nil {
		t.Fatalf("initiate (save card) against sandbox: %v", err)
	}
	if cs.PaymentPageURL == "" {
		t.Fatal("empty payment page url")
	}
	t.Logf("save-card checkout ready — tick 'save card', pay, then retrieve to confirm cardToken:\n  %s", cs.PaymentPageURL)
}

// TestSandbox_Refund exercises the refund paths against a real captured
// payment. It needs a paymentId from a payment you completed manually today
// (refunds are same-settlement-day in the sandbox).
//
//	IYZICO_SANDBOX_PAYMENT_ID=<id> [IYZICO_SANDBOX_PARTIAL_AMOUNT=1.00] \
//	go test -tags=iyzico_sandbox -run TestSandbox_Refund -v ./internal/payment/
func TestSandbox_Refund(t *testing.T) {
	g := sandboxGateway(t)
	paymentID := os.Getenv("IYZICO_SANDBOX_PAYMENT_ID")
	if paymentID == "" {
		t.Skip("IYZICO_SANDBOX_PAYMENT_ID not set; skipping refund (complete a payment first, then pass its id)")
	}

	if amount := os.Getenv("IYZICO_SANDBOX_PARTIAL_AMOUNT"); amount != "" {
		// Partial refund (a declined sub-order's slice). Parse leniently — the
		// operator supplies a plain decimal like "1.00".
		v, err := strconv.ParseFloat(amount, 64)
		if err != nil {
			t.Fatalf("bad IYZICO_SANDBOX_PARTIAL_AMOUNT %q: %v", amount, err)
		}
		if err := g.RefundPartial(context.Background(), paymentID, v); err != nil {
			t.Fatalf("partial refund: %v", err)
		}
		t.Logf("partial refund of %.2f on payment %s ok", v, paymentID)
		return
	}

	if err := g.Refund(context.Background(), paymentID); err != nil {
		t.Fatalf("full refund: %v", err)
	}
	t.Logf("full refund on payment %s ok", paymentID)
}
