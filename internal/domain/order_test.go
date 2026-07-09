package domain_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func TestNewOrder_Defaults(t *testing.T) {
	o := domain.NewOrder(7, "123 St")
	if o.UserID != 7 || o.DeliveryAddress != "123 St" {
		t.Errorf("unexpected order: %+v", o)
	}
	if o.Status != domain.OrderStatusPending || o.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("status/payment = %q/%q, want pending/pending", o.Status, o.PaymentStatus)
	}
}

func TestOrder_HappyPathLifecycle(t *testing.T) {
	o := domain.NewOrder(1, "addr")
	steps := []struct {
		do   func() error
		want string
	}{
		{o.Confirm, domain.OrderStatusConfirmed},
		{o.StartPreparing, domain.OrderStatusPreparing},
		{o.MarkReady, domain.OrderStatusReady},
		{o.StartDelivering, domain.OrderStatusDelivering},
		{o.MarkDelivered, domain.OrderStatusDelivered},
	}
	for _, s := range steps {
		if err := s.do(); err != nil {
			t.Fatalf("transition to %s failed: %v", s.want, err)
		}
		if o.Status != s.want {
			t.Fatalf("status = %q, want %q", o.Status, s.want)
		}
	}
	if o.ActualDeliveryTime == nil {
		t.Error("MarkDelivered should stamp ActualDeliveryTime")
	}
}

func TestOrder_IllegalTransitionsRejected(t *testing.T) {
	o := domain.NewOrder(1, "addr")
	// Cannot skip straight to preparing from pending.
	if err := o.StartPreparing(); err != domain.ErrInvalidStatusTransition {
		t.Errorf("StartPreparing on pending = %v, want ErrInvalidStatusTransition", err)
	}
	// Confirm twice is illegal.
	if err := o.Confirm(); err != nil {
		t.Fatalf("first confirm: %v", err)
	}
	if err := o.Confirm(); err != domain.ErrInvalidStatusTransition {
		t.Errorf("second confirm = %v, want ErrInvalidStatusTransition", err)
	}
}

func TestOrder_Cancel(t *testing.T) {
	pending := domain.NewOrder(1, "addr")
	if err := pending.Cancel(); err != nil {
		t.Fatalf("cancel pending: %v", err)
	}
	if pending.Status != domain.OrderStatusCancelled || pending.CancelledAt == nil {
		t.Errorf("cancel did not stamp state: %+v", pending)
	}

	// Cannot cancel once preparing.
	preparing := domain.NewOrder(1, "addr")
	_ = preparing.Confirm()
	_ = preparing.StartPreparing()
	if err := preparing.Cancel(); err != domain.ErrInvalidStatusTransition {
		t.Errorf("cancel preparing = %v, want ErrInvalidStatusTransition", err)
	}
}

func TestOrder_SettleCashOnDelivery(t *testing.T) {
	cash, card := domain.PaymentMethodCash, domain.PaymentMethodCard
	deliver := func(o *domain.Order) {
		_ = o.Confirm()
		_ = o.StartPreparing()
		_ = o.MarkReady()
		_ = o.StartDelivering()
		_ = o.MarkDelivered()
	}

	// Delivered cash order settles to paid.
	o := domain.NewOrder(1, "addr")
	o.PaymentMethod = &cash
	deliver(o)
	o.SettleCashOnDelivery()
	if o.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("delivered cash = %q, want paid", o.PaymentStatus)
	}

	// Card orders are untouched (the gateway drives them).
	o = domain.NewOrder(1, "addr")
	o.PaymentMethod = &card
	deliver(o)
	o.SettleCashOnDelivery()
	if o.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("delivered card = %q, want pending", o.PaymentStatus)
	}

	// Not delivered yet -> no-op.
	o = domain.NewOrder(1, "addr")
	o.PaymentMethod = &cash
	_ = o.Confirm()
	o.SettleCashOnDelivery()
	if o.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("confirmed cash = %q, want pending", o.PaymentStatus)
	}

	// Already refunded -> never overwritten.
	o = domain.NewOrder(1, "addr")
	o.PaymentMethod = &cash
	_ = o.MarkPaid()
	_ = o.Refund()
	deliver(o)
	o.SettleCashOnDelivery()
	if o.PaymentStatus != domain.PaymentStatusRefunded {
		t.Errorf("refunded cash = %q, want refunded", o.PaymentStatus)
	}

	// No payment method recorded -> no-op.
	o = domain.NewOrder(1, "addr")
	deliver(o)
	o.SettleCashOnDelivery()
	if o.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("no method = %q, want pending", o.PaymentStatus)
	}
}

func TestOrder_PaymentTransitions(t *testing.T) {
	o := domain.NewOrder(1, "addr")
	if err := o.MarkPaid(); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if o.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("payment status = %q, want paid", o.PaymentStatus)
	}
	if err := o.MarkPaid(); err != domain.ErrInvalidPaymentTransition {
		t.Errorf("double pay = %v, want ErrInvalidPaymentTransition", err)
	}
	if err := o.Refund(); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if o.PaymentStatus != domain.PaymentStatusRefunded {
		t.Errorf("payment status = %q, want refunded", o.PaymentStatus)
	}
}

func TestOrder_HasChef(t *testing.T) {
	o := domain.NewOrder(1, "addr")
	o.Items = []*domain.OrderItem{{ChefID: 3}, {ChefID: 5}}
	if !o.HasChef(5) {
		t.Error("HasChef(5) should be true")
	}
	if o.HasChef(9) {
		t.Error("HasChef(9) should be false")
	}
}

func TestNewOrderItem_Subtotal(t *testing.T) {
	it := domain.NewOrderItem(2, 3, "Soup", 4, 2.5)
	if it.Subtotal != 10 {
		t.Errorf("subtotal = %v, want 10", it.Subtotal)
	}
	if it.MenuItemID != 2 || it.ChefID != 3 || it.Quantity != 4 {
		t.Errorf("unexpected item: %+v", it)
	}
}
