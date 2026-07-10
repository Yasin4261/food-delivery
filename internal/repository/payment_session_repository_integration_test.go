//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestPaymentSessionRepository_RoundTrip(t *testing.T) {
	resetDB(t)
	repo := repository.NewPaymentSessionRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 10)
	orderID := seedDeliveredOrder(t, customer.ID, chef.ID, item, "ORD-PAY-1")

	s := &domain.PaymentSession{OrderID: orderID, Token: "tok-abc"}
	if err := repo.Create(ctx(), s); err != nil {
		t.Fatalf("create: %v", err)
	}
	if s.ID == 0 || s.Status != domain.PaymentSessionInitiated {
		t.Fatalf("create defaults wrong: %+v", s)
	}

	got, err := repo.FindByToken(ctx(), "tok-abc")
	if err != nil || got.OrderID != orderID {
		t.Fatalf("find by token = %+v, %v", got, err)
	}

	// Mark paid with a gateway payment id; find it via FindPaidByOrder.
	payID := "pay-9"
	if err := repo.UpdateStatus(ctx(), s.ID, domain.PaymentSessionPaid, &payID); err != nil {
		t.Fatalf("update: %v", err)
	}
	paid, err := repo.FindPaidByOrder(ctx(), orderID)
	if err != nil || paid.PaymentID == nil || *paid.PaymentID != "pay-9" {
		t.Fatalf("paid session = %+v, %v", paid, err)
	}

	// Refunded keeps the payment id (COALESCE with nil).
	if err := repo.UpdateStatus(ctx(), s.ID, domain.PaymentSessionRefunded, nil); err != nil {
		t.Fatalf("refund update: %v", err)
	}
	got, _ = repo.FindByToken(ctx(), "tok-abc")
	if got.Status != domain.PaymentSessionRefunded || got.PaymentID == nil || *got.PaymentID != "pay-9" {
		t.Errorf("after refund = %+v, want refunded with payment id retained", got)
	}

	if _, err := repo.FindByToken(ctx(), "ghost"); !errors.Is(err, domain.ErrPaymentSessionNotFound) {
		t.Errorf("missing token = %v, want ErrPaymentSessionNotFound", err)
	}
	if err := repo.UpdateStatus(ctx(), 9999, domain.PaymentSessionPaid, nil); !errors.Is(err, domain.ErrPaymentSessionNotFound) {
		t.Errorf("update missing = %v, want ErrPaymentSessionNotFound", err)
	}
}
