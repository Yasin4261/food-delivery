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

func (f *fakeOrderRepo) Create(_ context.Context, o *domain.Order) error {
	o.ID = f.nextID
	f.nextID++
	cp := *o
	f.orders[o.ID] = &cp
	return nil
}
func (f *fakeOrderRepo) FindByID(_ context.Context, id int) (*domain.Order, error) {
	if o, ok := f.orders[id]; ok {
		cp := *o
		return &cp, nil
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
	cp := *o
	f.orders[o.ID] = &cp
	return nil
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
		if o.Status == domain.OrderStatusPending && o.HasChef(chefID) {
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
	svc := service.NewOrderService(newFakeOrderRepo(), itemRepo, chefRepo, nil)
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
