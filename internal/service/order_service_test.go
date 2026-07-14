package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeOrderRepo is an in-memory domain.OrderRepository for service tests.
type fakeOrderRepo struct {
	orders map[int]*domain.Order
	nextID int
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{orders: map[int]*domain.Order{}, nextID: 1}
}

// copyOrder deep-copies the sub-orders so mutations on a fetched order never
// leak into "storage" without an explicit update (mirrors a real database).
func copyOrder(o *domain.Order) *domain.Order {
	cp := *o
	cp.SubOrders = make([]*domain.SubOrder, 0, len(o.SubOrders))
	for _, s := range o.SubOrders {
		sc := *s
		cp.SubOrders = append(cp.SubOrders, &sc)
	}
	return &cp
}

func (f *fakeOrderRepo) Create(_ context.Context, o *domain.Order) error {
	o.ID = f.nextID
	f.nextID++
	f.orders[o.ID] = copyOrder(o)
	return nil
}
func (f *fakeOrderRepo) FindByID(_ context.Context, id int) (*domain.Order, error) {
	if o, ok := f.orders[id]; ok {
		return copyOrder(o), nil
	}
	return nil, domain.ErrOrderNotFound
}
func (f *fakeOrderRepo) ListByUser(_ context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	out := make([]*domain.Order, 0)
	for _, o := range f.orders {
		if o.UserID == userID {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}
func (f *fakeOrderRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Order, int, error) {
	out := make([]*domain.Order, 0)
	for _, o := range f.orders {
		if o.HasChef(chefID) {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, len(out), nil
}
func (f *fakeOrderRepo) UpdateStatus(_ context.Context, o *domain.Order) error {
	if _, ok := f.orders[o.ID]; !ok {
		return domain.ErrOrderNotFound
	}
	f.orders[o.ID] = copyOrder(o)
	return nil
}
func (f *fakeOrderRepo) UpdateSubOrder(ctx context.Context, o *domain.Order, _ *domain.SubOrder) error {
	return f.UpdateStatus(ctx, o)
}
func (f *fakeOrderRepo) CountActiveByUser(_ context.Context, userID int) (int, error) {
	n := 0
	for _, o := range f.orders {
		if o.UserID == userID && o.Status != domain.OrderStatusDelivered && o.Status != domain.OrderStatusCancelled {
			n++
		}
	}
	return n, nil
}
func (f *fakeOrderRepo) CountPendingByChef(_ context.Context, chefID int) (int, error) {
	n := 0
	for _, o := range f.orders {
		if s := o.SubOrderFor(chefID); s != nil && s.Status == domain.OrderStatusPending {
			n++
		}
	}
	return n, nil
}

// orderFixture wires an OrderService over fakes, seeds chef profiles for the
// given user ids, and returns the service plus the item repo for seeding dishes.
func orderFixture(t *testing.T, userIDs ...int) (*service.OrderService, *fakeMenuItemRepo, *fakeChefRepo) {
	t.Helper()
	chefRepo := newFakeChefRepo()
	for _, uid := range userIDs {
		if err := chefRepo.Create(context.Background(), &domain.Chef{UserID: uid, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}
	itemRepo := newFakeMenuItemRepo()
	svc := service.NewOrderService(newFakeOrderRepo(), itemRepo, chefRepo, nil, nil, nil, domain.FeePolicy{}, nil, nil)
	return svc, itemRepo, chefRepo
}

// seedItem inserts a limited-stock dish owned by chefID and returns it.
func seedItem(t *testing.T, repo *fakeMenuItemRepo, chefID int, price float64, qty int) *domain.MenuItem {
	t.Helper()
	item := domain.NewMenuItem(1, chefID, "Dish", price)
	item.AvailableQuantity = &qty
	if err := repo.Create(context.Background(), item); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	return item
}

func TestOrderService_PlaceOrder_Success(t *testing.T) {
	svc, items, _ := orderFixture(t, 1) // user1 -> chef1
	item := seedItem(t, items, 1, 5, 10)

	order, err := svc.PlaceOrder(context.Background(), 100, service.PlaceOrderInput{
		DeliveryAddress: "123 St",
		PaymentMethod:   domain.PaymentMethodCash,
		Lines:           []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	if order.TotalPrice != 10 || order.Subtotal != 10 {
		t.Errorf("totals = %v/%v, want 10/10", order.Subtotal, order.TotalPrice)
	}
	if len(order.Items) != 1 || order.Items[0].ChefID != 1 || order.OrderCode == "" {
		t.Errorf("unexpected order: %+v", order)
	}
	// Stock decremented 10 -> 8.
	if got, _ := items.FindByID(context.Background(), item.ID); *got.AvailableQuantity != 8 {
		t.Errorf("stock = %d, want 8", *got.AvailableQuantity)
	}
}

func TestOrderService_PlaceOrder_MultiChef(t *testing.T) {
	svc, items, _ := orderFixture(t, 1, 2)
	a := seedItem(t, items, 1, 5, 10)
	b := seedItem(t, items, 2, 3, 10)

	order, err := svc.PlaceOrder(context.Background(), 100, service.PlaceOrderInput{
		DeliveryAddress: "123 St",
		PaymentMethod:   domain.PaymentMethodCard,
		Lines: []service.OrderLineInput{
			{MenuItemID: a.ID, Quantity: 1},
			{MenuItemID: b.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	if order.TotalPrice != 11 {
		t.Errorf("total = %v, want 11", order.TotalPrice)
	}
	if !order.HasChef(1) || !order.HasChef(2) {
		t.Errorf("order should span chefs 1 and 2: %+v", order.Items)
	}
}

func TestOrderService_PlaceOrder_Errors(t *testing.T) {
	svc, items, _ := orderFixture(t, 1)
	item := seedItem(t, items, 1, 5, 3)
	ctx := context.Background()
	base := func() service.PlaceOrderInput {
		return service.PlaceOrderInput{
			DeliveryAddress: "x",
			PaymentMethod:   domain.PaymentMethodCash,
			Lines:           []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}},
		}
	}

	t.Run("empty cart", func(t *testing.T) {
		in := base()
		in.Lines = nil
		if _, err := svc.PlaceOrder(ctx, 1, in); !errors.Is(err, domain.ErrEmptyOrder) {
			t.Errorf("err = %v, want ErrEmptyOrder", err)
		}
	})
	t.Run("bad payment", func(t *testing.T) {
		in := base()
		in.PaymentMethod = "bitcoin"
		if _, err := svc.PlaceOrder(ctx, 1, in); !isValidation(err) {
			t.Errorf("err = %v, want ValidationError", err)
		}
	})
	t.Run("out of stock", func(t *testing.T) {
		in := base()
		in.Lines = []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 99}}
		if _, err := svc.PlaceOrder(ctx, 1, in); !errors.Is(err, domain.ErrItemOutOfStock) {
			t.Errorf("err = %v, want ErrItemOutOfStock", err)
		}
	})
	t.Run("missing item", func(t *testing.T) {
		in := base()
		in.Lines = []service.OrderLineInput{{MenuItemID: 999, Quantity: 1}}
		if _, err := svc.PlaceOrder(ctx, 1, in); !errors.Is(err, domain.ErrMenuItemNotFound) {
			t.Errorf("err = %v, want ErrMenuItemNotFound", err)
		}
	})
}

func TestOrderService_CustomerOwnership(t *testing.T) {
	svc, items, _ := orderFixture(t, 1)
	item := seedItem(t, items, 1, 5, 10)
	ctx := context.Background()
	order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		DeliveryAddress: "x", PaymentMethod: domain.PaymentMethodCash,
		Lines: []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}

	if _, err := svc.GetForCustomer(ctx, 100, order.ID); err != nil {
		t.Errorf("owner get failed: %v", err)
	}
	if _, err := svc.GetForCustomer(ctx, 200, order.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("non-owner get = %v, want ErrForbidden", err)
	}
}

// TestOrderService_CashSettlesOnDelivery: delivering a cash order marks it
// paid (so it counts toward earnings); card orders stay pending for the
// gateway (#42 phase 2).
func TestOrderService_CashSettlesOnDelivery(t *testing.T) {
	svc, items, _ := orderFixture(t, 1)
	item := seedItem(t, items, 1, 5, 10)
	ctx := context.Background()

	place := func(method string) *domain.Order {
		t.Helper()
		order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
			DeliveryAddress: "x", PaymentMethod: method,
			Lines: []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}},
		})
		if err != nil {
			t.Fatalf("place %s order: %v", method, err)
		}
		return order
	}
	deliver := func(orderID int) *domain.Order {
		t.Helper()
		var out *domain.Order
		for _, action := range []string{"confirm", "preparing", "ready", "delivering", "delivered"} {
			o, err := svc.AdvanceForChef(ctx, 1, orderID, action)
			if err != nil {
				t.Fatalf("advance %s: %v", action, err)
			}
			out = o
		}
		return out
	}

	cash := deliver(place(domain.PaymentMethodCash).ID)
	if cash.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("delivered cash order payment = %q, want paid", cash.PaymentStatus)
	}

	card := deliver(place(domain.PaymentMethodCard).ID)
	if card.PaymentStatus != domain.PaymentStatusPending {
		t.Errorf("delivered card order payment = %q, want pending (gateway drives it)", card.PaymentStatus)
	}
}

func TestOrderService_AdvanceForChef(t *testing.T) {
	svc, items, _ := orderFixture(t, 1, 2) // user1->chef1, user2->chef2
	item := seedItem(t, items, 1, 5, 10)
	ctx := context.Background()
	order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		DeliveryAddress: "x", PaymentMethod: domain.PaymentMethodCash,
		Lines: []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}

	// The participating chef (user1 -> chef1) can confirm.
	updated, err := svc.AdvanceForChef(ctx, 1, order.ID, service.OrderActionConfirm)
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if updated.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want confirmed", updated.Status)
	}

	// A chef with no items in the order (user2 -> chef2) cannot.
	if _, err := svc.AdvanceForChef(ctx, 2, order.ID, service.OrderActionPreparing); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("unrelated chef = %v, want ErrForbidden", err)
	}

	// Unknown action is a validation error.
	if _, err := svc.AdvanceForChef(ctx, 1, order.ID, "teleport"); !isValidation(err) {
		t.Errorf("unknown action = %v, want ValidationError", err)
	}
}

// placeMultiChef seeds two chefs (users 1 and 2) with one dish each and places
// a single order spanning both, returning the service, repo and order.
func placeMultiChef(t *testing.T, method string, refunder domain.PaymentRefunder) (*service.OrderService, *fakeOrderRepo, *domain.Order) {
	t.Helper()
	ctx := context.Background()
	chefRepo := newFakeChefRepo()
	for _, uid := range []int{1, 2} {
		if err := chefRepo.Create(ctx, &domain.Chef{UserID: uid, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}
	items := newFakeMenuItemRepo()
	a := seedItem(t, items, 1, 5, 10)
	b := seedItem(t, items, 2, 3, 10)
	orders := newFakeOrderRepo()
	svc := service.NewOrderService(orders, items, chefRepo, nil, nil, nil, domain.FeePolicy{}, refunder, nil)

	order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		DeliveryAddress: "x", PaymentMethod: method,
		Lines: []service.OrderLineInput{
			{MenuItemID: a.ID, Quantity: 1}, // chef1: 5
			{MenuItemID: b.ID, Quantity: 2}, // chef2: 6
		},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	return svc, orders, order
}

// The issue's acceptance case: chef A advancing never changes chef B's slice,
// and the order-level status derives from the least-advanced active sub-order.
func TestOrderService_SubOrderIsolation(t *testing.T) {
	svc, _, order := placeMultiChef(t, domain.PaymentMethodCash, nil)
	ctx := context.Background()

	if len(order.SubOrders) != 2 {
		t.Fatalf("sub-orders = %d, want 2", len(order.SubOrders))
	}
	if order.SubOrders[0].Subtotal != 5 || order.SubOrders[1].Subtotal != 6 {
		t.Fatalf("sub-order subtotals = %v/%v, want 5/6", order.SubOrders[0].Subtotal, order.SubOrders[1].Subtotal)
	}

	// Chef 1 races ahead to delivering; chef 2 hasn't touched theirs.
	for _, action := range []string{"confirm", "preparing", "ready", "delivering"} {
		if _, err := svc.AdvanceForChef(ctx, 1, order.ID, action); err != nil {
			t.Fatalf("chef1 %s: %v", action, err)
		}
	}
	got, err := svc.AdvanceForChef(ctx, 2, order.ID, service.OrderActionConfirm)
	if err != nil {
		t.Fatalf("chef2 confirm: %v", err)
	}
	if s := got.SubOrderFor(1).Status; s != domain.OrderStatusDelivering {
		t.Errorf("chef1 sub-order = %q, want delivering (untouched by chef2)", s)
	}
	if s := got.SubOrderFor(2).Status; s != domain.OrderStatusConfirmed {
		t.Errorf("chef2 sub-order = %q, want confirmed", s)
	}
	// Order-level status = least-advanced active sub-order.
	if got.Status != domain.OrderStatusConfirmed {
		t.Errorf("order status = %q, want confirmed (derived)", got.Status)
	}

	// Chef 2 cannot replay chef 1's transitions on their own slice.
	if _, err := svc.AdvanceForChef(ctx, 2, order.ID, service.OrderActionReady); !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Errorf("chef2 illegal move = %v, want ErrInvalidStatusTransition", err)
	}
}

// A cash multi-chef order settles to paid only when every sub-order is
// delivered — the derived order status drives SettleCashOnDelivery.
func TestOrderService_MultiChefCashSettlesWhenAllDelivered(t *testing.T) {
	svc, orders, order := placeMultiChef(t, domain.PaymentMethodCash, nil)
	ctx := context.Background()
	all := []string{"confirm", "preparing", "ready", "delivering", "delivered"}

	for _, action := range all {
		if _, err := svc.AdvanceForChef(ctx, 1, order.ID, action); err != nil {
			t.Fatalf("chef1 %s: %v", action, err)
		}
	}
	mid, _ := orders.FindByID(ctx, order.ID)
	if mid.Status != domain.OrderStatusPending || mid.PaymentStatus != domain.PaymentStatusPending {
		t.Fatalf("after chef1 only: status=%q payment=%q, want pending/pending", mid.Status, mid.PaymentStatus)
	}

	var final *domain.Order
	for _, action := range all {
		o, err := svc.AdvanceForChef(ctx, 2, order.ID, action)
		if err != nil {
			t.Fatalf("chef2 %s: %v", action, err)
		}
		final = o
	}
	if final.Status != domain.OrderStatusDelivered {
		t.Errorf("order status = %q, want delivered (all sub-orders delivered)", final.Status)
	}
	if final.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("payment = %q, want paid (cash settled on full delivery)", final.PaymentStatus)
	}
}

// Declining one slice of a card-paid order partial-refunds that chef's
// subtotal; the other chef's slice and the order stay alive. Declining the
// last active slice marks the whole payment refunded.
func TestOrderService_DeclinePartialRefund(t *testing.T) {
	refunder := &recordingRefunder{}
	svc, orders, order := placeMultiChef(t, domain.PaymentMethodCard, refunder)
	ctx := context.Background()

	stored, _ := orders.FindByID(ctx, order.ID)
	_ = stored.MarkPaid()
	_ = orders.UpdateStatus(ctx, stored)

	got, err := svc.AdvanceForChef(ctx, 1, order.ID, service.OrderActionDecline)
	if err != nil {
		t.Fatalf("chef1 decline: %v", err)
	}
	if refunder.partialCalls != 1 || refunder.partialAmounts[0] != 5 {
		t.Fatalf("partial refunds = %d %v, want one of 5", refunder.partialCalls, refunder.partialAmounts)
	}
	if s := got.SubOrderFor(2).Status; s != domain.OrderStatusPending {
		t.Errorf("chef2 sub-order = %q, want pending (unaffected by decline)", s)
	}
	if got.Status != domain.OrderStatusPending || got.PaymentStatus != domain.PaymentStatusPaid {
		t.Errorf("order = %q/%q, want pending/paid (still alive)", got.Status, got.PaymentStatus)
	}

	// The last chef declining cancels the order; every slice has been refunded.
	final, err := svc.AdvanceForChef(ctx, 2, order.ID, service.OrderActionDecline)
	if err != nil {
		t.Fatalf("chef2 decline: %v", err)
	}
	if refunder.partialCalls != 2 || refunder.partialAmounts[1] != 6 {
		t.Fatalf("partial refunds = %d %v, want second of 6", refunder.partialCalls, refunder.partialAmounts)
	}
	if final.Status != domain.OrderStatusCancelled || final.PaymentStatus != domain.PaymentStatusRefunded {
		t.Errorf("order = %q/%q, want cancelled/refunded", final.Status, final.PaymentStatus)
	}
}

// A failing partial refund aborts the decline — the sub-order stays alive so
// money and state never diverge.
func TestOrderService_DeclineAbortsWhenPartialRefundFails(t *testing.T) {
	refunder := &recordingRefunder{err: errors.New("gateway down")}
	svc, orders, order := placeMultiChef(t, domain.PaymentMethodCard, refunder)
	ctx := context.Background()

	stored, _ := orders.FindByID(ctx, order.ID)
	_ = stored.MarkPaid()
	_ = orders.UpdateStatus(ctx, stored)

	if _, err := svc.AdvanceForChef(ctx, 1, order.ID, service.OrderActionDecline); err == nil {
		t.Fatal("decline should fail when the partial refund fails")
	}
	after, _ := orders.FindByID(ctx, order.ID)
	if s := after.SubOrderFor(1).Status; s == domain.OrderStatusCancelled {
		t.Error("sub-order must stay uncancelled when the refund fails")
	}
}

// Customer cancel is blocked once any chef has started preparing, and a
// successful cancel takes every sub-order with it.
func TestOrderService_CustomerCancelWithSubOrders(t *testing.T) {
	svc, _, order := placeMultiChef(t, domain.PaymentMethodCash, nil)
	ctx := context.Background()

	// Chef 1 starts preparing -> the whole order is locked in.
	for _, action := range []string{"confirm", "preparing"} {
		if _, err := svc.AdvanceForChef(ctx, 1, order.ID, action); err != nil {
			t.Fatalf("chef1 %s: %v", action, err)
		}
	}
	if _, err := svc.CancelForCustomer(ctx, 100, order.ID); !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Errorf("cancel after preparing = %v, want ErrInvalidStatusTransition", err)
	}

	// A fresh order cancels cleanly, cancelling both slices.
	svc2, _, order2 := placeMultiChef(t, domain.PaymentMethodCash, nil)
	cancelled, err := svc2.CancelForCustomer(ctx, 100, order2.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	for _, s := range cancelled.SubOrders {
		if s.Status != domain.OrderStatusCancelled {
			t.Errorf("sub-order chef %d = %q, want cancelled", s.ChefID, s.Status)
		}
	}
}

// The money model (#65): distance-based delivery fee per slice (base only
// without coordinates), commission snapshotted from the policy, order total =
// subtotal + delivery fees, and a declined card-paid slice refunds food +
// its delivery fee.
func TestOrderService_FeesAndCommission(t *testing.T) {
	ctx := context.Background()
	chefRepo := newFakeChefRepo()
	// Chef 1 has kitchen coordinates ~10 km from the delivery point; chef 2
	// has none (base fee only).
	lat1, lng1 := 41.0, 29.0
	if err := chefRepo.Create(ctx, &domain.Chef{UserID: 1, IsActive: true, KitchenLatitude: &lat1, KitchenLongitude: &lng1}); err != nil {
		t.Fatalf("seed chef1: %v", err)
	}
	if err := chefRepo.Create(ctx, &domain.Chef{UserID: 2, IsActive: true}); err != nil {
		t.Fatalf("seed chef2: %v", err)
	}
	items := newFakeMenuItemRepo()
	a := seedItem(t, items, 1, 100, 10)
	b := seedItem(t, items, 2, 50, 10)
	orders := newFakeOrderRepo()
	refunder := &recordingRefunder{}
	policy := domain.FeePolicy{DeliveryBaseFee: 10, DeliveryFeePerKm: 2, CommissionRate: 10}
	svc := service.NewOrderService(orders, items, chefRepo, nil, nil, nil, policy, refunder, nil)

	// Delivery point ~0.09 degrees north of chef 1 (~10.0 km Haversine).
	dlat, dlng := 41.09, 29.0
	order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		DeliveryAddress: "x", PaymentMethod: domain.PaymentMethodCard,
		DeliveryLatitude: &dlat, DeliveryLongitude: &dlng,
		Lines: []service.OrderLineInput{
			{MenuItemID: a.ID, Quantity: 1}, // chef1: 100
			{MenuItemID: b.ID, Quantity: 1}, // chef2: 50
		},
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}

	sub1, sub2 := order.SubOrderFor(1), order.SubOrderFor(2)
	// Chef 1: base 10 + 2/km over ~10 km => ~30; allow Haversine wiggle.
	if sub1.DeliveryFee < 29 || sub1.DeliveryFee > 31 {
		t.Errorf("chef1 delivery fee = %v, want ~30 (distance-based)", sub1.DeliveryFee)
	}
	// Chef 2 has no coordinates: base only.
	if sub2.DeliveryFee != 10 {
		t.Errorf("chef2 delivery fee = %v, want 10 (base only)", sub2.DeliveryFee)
	}
	// Commission: 10% of each food subtotal, never charged to the customer.
	if sub1.Commission != 10 || sub2.Commission != 5 {
		t.Errorf("commissions = %v/%v, want 10/5", sub1.Commission, sub2.Commission)
	}
	wantTotal := domain.RoundMoney(150 + sub1.DeliveryFee + sub2.DeliveryFee)
	if order.DeliveryFee != domain.RoundMoney(sub1.DeliveryFee+sub2.DeliveryFee) || order.TotalPrice != wantTotal {
		t.Errorf("order fees/total = %v/%v, want %v/%v", order.DeliveryFee, order.TotalPrice, sub1.DeliveryFee+sub2.DeliveryFee, wantTotal)
	}

	// Declining a paid slice refunds food + its delivery fee.
	stored, _ := orders.FindByID(ctx, order.ID)
	_ = stored.MarkPaid()
	_ = orders.UpdateStatus(ctx, stored)
	if _, err := svc.AdvanceForChef(ctx, 2, order.ID, service.OrderActionDecline); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if len(refunder.partialAmounts) != 1 || refunder.partialAmounts[0] != 50+10 {
		t.Errorf("refund = %v, want [60] (food 50 + delivery 10)", refunder.partialAmounts)
	}
}

// A zero-value policy keeps everything free — the pre-#65 behaviour.
func TestOrderService_ZeroPolicyIsFree(t *testing.T) {
	_, _, order := placeMultiChef(t, domain.PaymentMethodCash, nil)
	if order.DeliveryFee != 0 || order.TotalPrice != order.Subtotal {
		t.Errorf("zero policy charged fees: %+v", order)
	}
	for _, sub := range order.SubOrders {
		if sub.DeliveryFee != 0 || sub.Commission != 0 {
			t.Errorf("zero policy sub fees: %+v", sub)
		}
	}
}
