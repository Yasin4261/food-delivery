package domain

import "testing"

func sub(status string) *SubOrder {
	return &SubOrder{Status: status}
}

func TestSubOrderLifecycle(t *testing.T) {
	s := NewSubOrder(7, 25.50)
	if s.Status != OrderStatusPending {
		t.Fatalf("new sub-order status = %q, want pending", s.Status)
	}

	steps := []struct {
		name string
		fn   func() error
		want string
	}{
		{"confirm", s.Confirm, OrderStatusConfirmed},
		{"preparing", s.StartPreparing, OrderStatusPreparing},
		{"ready", s.MarkReady, OrderStatusReady},
		{"delivering", s.StartDelivering, OrderStatusDelivering},
		{"delivered", s.MarkDelivered, OrderStatusDelivered},
	}
	for _, step := range steps {
		if err := step.fn(); err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if s.Status != step.want {
			t.Fatalf("%s: status = %q, want %q", step.name, s.Status, step.want)
		}
	}
}

func TestSubOrderIllegalTransitions(t *testing.T) {
	cases := []struct {
		name string
		s    *SubOrder
		fn   func(*SubOrder) error
	}{
		{"confirm from preparing", sub(OrderStatusPreparing), (*SubOrder).Confirm},
		{"prepare from pending", sub(OrderStatusPending), (*SubOrder).StartPreparing},
		{"deliver from confirmed", sub(OrderStatusConfirmed), (*SubOrder).StartDelivering},
		{"delivered from ready", sub(OrderStatusReady), (*SubOrder).MarkDelivered},
		{"cancel from preparing", sub(OrderStatusPreparing), (*SubOrder).Cancel},
		{"cancel from delivered", sub(OrderStatusDelivered), (*SubOrder).Cancel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := tc.s.Status
			if err := tc.fn(tc.s); err != ErrInvalidStatusTransition {
				t.Fatalf("err = %v, want ErrInvalidStatusTransition", err)
			}
			if tc.s.Status != before {
				t.Fatalf("status mutated to %q on illegal transition", tc.s.Status)
			}
		})
	}
}

func TestDeriveOrderStatus(t *testing.T) {
	cases := []struct {
		name string
		subs []*SubOrder
		want string
	}{
		{"no sub-orders", nil, OrderStatusPending},
		{"single pending", []*SubOrder{sub(OrderStatusPending)}, OrderStatusPending},
		{"single delivered", []*SubOrder{sub(OrderStatusDelivered)}, OrderStatusDelivered},
		{"least advanced wins", []*SubOrder{sub(OrderStatusDelivering), sub(OrderStatusConfirmed)}, OrderStatusConfirmed},
		{"delivered + preparing", []*SubOrder{sub(OrderStatusDelivered), sub(OrderStatusPreparing)}, OrderStatusPreparing},
		{"all delivered", []*SubOrder{sub(OrderStatusDelivered), sub(OrderStatusDelivered)}, OrderStatusDelivered},
		{"cancelled ignored", []*SubOrder{sub(OrderStatusCancelled), sub(OrderStatusReady)}, OrderStatusReady},
		{"cancelled + delivered", []*SubOrder{sub(OrderStatusCancelled), sub(OrderStatusDelivered)}, OrderStatusDelivered},
		{"all cancelled", []*SubOrder{sub(OrderStatusCancelled), sub(OrderStatusCancelled)}, OrderStatusCancelled},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeriveOrderStatus(tc.subs); got != tc.want {
				t.Fatalf("DeriveOrderStatus = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestSyncStatusFromSubOrders(t *testing.T) {
	o := NewOrder(1, "addr")
	o.SubOrders = []*SubOrder{sub(OrderStatusDelivered), sub(OrderStatusDelivered)}
	o.SyncStatusFromSubOrders()
	if o.Status != OrderStatusDelivered {
		t.Fatalf("status = %q, want delivered", o.Status)
	}
	if o.ActualDeliveryTime == nil {
		t.Fatal("ActualDeliveryTime not stamped on derived delivery")
	}

	o = NewOrder(1, "addr")
	o.SubOrders = []*SubOrder{sub(OrderStatusCancelled), sub(OrderStatusCancelled)}
	o.SyncStatusFromSubOrders()
	if o.Status != OrderStatusCancelled {
		t.Fatalf("status = %q, want cancelled", o.Status)
	}
	if o.CancelledAt == nil {
		t.Fatal("CancelledAt not stamped on derived cancellation")
	}
}

func TestOrderCancelCancelsSubOrders(t *testing.T) {
	o := NewOrder(1, "addr")
	o.SubOrders = []*SubOrder{sub(OrderStatusPending), sub(OrderStatusConfirmed)}
	if err := o.Cancel(); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	for i, s := range o.SubOrders {
		if s.Status != OrderStatusCancelled {
			t.Fatalf("sub-order %d status = %q, want cancelled", i, s.Status)
		}
	}
}

func TestOrderCanCancelBlockedByAdvancedSubOrder(t *testing.T) {
	o := NewOrder(1, "addr")
	o.Status = OrderStatusConfirmed
	o.SubOrders = []*SubOrder{sub(OrderStatusPending), sub(OrderStatusPreparing)}
	if o.CanCancel() {
		t.Fatal("CanCancel = true with a preparing sub-order")
	}
	if err := o.Cancel(); err != ErrInvalidStatusTransition {
		t.Fatalf("cancel err = %v, want ErrInvalidStatusTransition", err)
	}
}

func TestSubOrderFor(t *testing.T) {
	o := NewOrder(1, "addr")
	a, b := NewSubOrder(1, 10), NewSubOrder(2, 20)
	o.SubOrders = []*SubOrder{a, b}
	if got := o.SubOrderFor(2); got != b {
		t.Fatalf("SubOrderFor(2) = %+v, want chef 2's sub-order", got)
	}
	if got := o.SubOrderFor(99); got != nil {
		t.Fatalf("SubOrderFor(99) = %+v, want nil", got)
	}
}
