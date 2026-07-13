//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// buildOrder assembles a pending order for userID containing the given items,
// with one sub-order per participating chef (as the order service does).
func buildOrder(userID int, code string, items ...*domain.OrderItem) *domain.Order {
	o := domain.NewOrder(userID, "123 St")
	o.OrderCode = code
	method := domain.PaymentMethodCash
	o.PaymentMethod = &method
	var subtotal float64
	for _, it := range items {
		o.Items = append(o.Items, it)
		subtotal += it.Subtotal
		if s := o.SubOrderFor(it.ChefID); s != nil {
			s.Subtotal += it.Subtotal
		} else {
			o.SubOrders = append(o.SubOrders, domain.NewSubOrder(it.ChefID, it.Subtotal))
		}
	}
	o.Subtotal = subtotal
	o.TotalPrice = subtotal
	return o
}

func TestOrderRepository_CreateTransactionAndFind(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 10)

	order := buildOrder(customer.ID, "ORD-TEST-1",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 2, item.Price))
	if err := repo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.ID == 0 || order.Items[0].ID == 0 || order.Items[0].OrderID != order.ID {
		t.Fatalf("create did not back-fill ids: %+v", order)
	}
	if len(order.SubOrders) != 1 || order.SubOrders[0].ID == 0 || order.SubOrders[0].OrderID != order.ID {
		t.Fatalf("create did not persist sub-orders: %+v", order.SubOrders)
	}

	got, err := repo.FindByID(ctx(), order.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Status != domain.OrderStatusPending || got.TotalPrice != 10 || len(got.Items) != 1 {
		t.Errorf("unexpected order: %+v", got)
	}
	if got.Items[0].ChefID != chef.ID || got.Items[0].Quantity != 2 {
		t.Errorf("unexpected item: %+v", got.Items[0])
	}
	if len(got.SubOrders) != 1 || got.SubOrders[0].Status != domain.OrderStatusPending ||
		got.SubOrders[0].Subtotal != 10 || got.SubOrders[0].ChefName == "" {
		t.Errorf("unexpected sub-orders: %+v", got.SubOrders)
	}
}

// TestOrderRepository_MultiChefScoping verifies a single order spanning two
// chefs, and that ListByChef returns each order with only that chef's items.
func TestOrderRepository_MultiChefScoping(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")

	chefA := seedChef(t, seedUser(t, "chefa@example.com").ID)
	menuA := seedMenu(t, chefA.ID)
	itemA := seedItem(t, menuA.ID, chefA.ID, 5, 10)

	chefB := seedChef(t, seedUser(t, "chefb@example.com").ID)
	menuB := seedMenu(t, chefB.ID)
	itemB := seedItem(t, menuB.ID, chefB.ID, 3, 10)

	order := buildOrder(customer.ID, "ORD-MULTI",
		domain.NewOrderItem(itemA.ID, chefA.ID, itemA.Name, 1, 5),
		domain.NewOrderItem(itemB.ID, chefB.ID, itemB.Name, 2, 3))
	if err := repo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Customer view: all items.
	all, _, _ := repo.ListByUser(ctx(), customer.ID, 20, 0)
	if len(all) != 1 || len(all[0].Items) != 2 {
		t.Fatalf("customer view items = %d, want 2", len(all[0].Items))
	}

	// Chef A view: only chef A's item.
	aOrders, _, err := repo.ListByChef(ctx(), chefA.ID, 20, 0)
	if err != nil {
		t.Fatalf("list by chef A: %v", err)
	}
	if len(aOrders) != 1 || len(aOrders[0].Items) != 1 || aOrders[0].Items[0].ChefID != chefA.ID {
		t.Errorf("chef A scoping wrong: %+v", aOrders)
	}
	bOrders, _, _ := repo.ListByChef(ctx(), chefB.ID, 20, 0)
	if len(bOrders) != 1 || len(bOrders[0].Items) != 1 || bOrders[0].Items[0].ChefID != chefB.ID {
		t.Errorf("chef B scoping wrong: %+v", bOrders)
	}

	// Both chef views keep the whole order's sub-orders visible (progress),
	// each with its own subtotal.
	if len(aOrders[0].SubOrders) != 2 {
		t.Fatalf("chef view sub-orders = %d, want 2", len(aOrders[0].SubOrders))
	}
	subA, subB := aOrders[0].SubOrderFor(chefA.ID), aOrders[0].SubOrderFor(chefB.ID)
	if subA == nil || subA.Subtotal != 5 || subB == nil || subB.Subtotal != 6 {
		t.Errorf("sub-order split wrong: %+v / %+v", subA, subB)
	}
}

// TestOrderRepository_UpdateSubOrder: chef A's transition persists atomically
// with the derived order status, and chef B's sub-order is untouched.
func TestOrderRepository_UpdateSubOrder(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chefA := seedChef(t, seedUser(t, "chefa@example.com").ID)
	menuA := seedMenu(t, chefA.ID)
	itemA := seedItem(t, menuA.ID, chefA.ID, 5, 10)
	chefB := seedChef(t, seedUser(t, "chefb@example.com").ID)
	menuB := seedMenu(t, chefB.ID)
	itemB := seedItem(t, menuB.ID, chefB.ID, 3, 10)

	order := buildOrder(customer.ID, "ORD-SUB",
		domain.NewOrderItem(itemA.ID, chefA.ID, itemA.Name, 1, 5),
		domain.NewOrderItem(itemB.ID, chefB.ID, itemB.Name, 1, 3))
	if err := repo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}

	sub := order.SubOrderFor(chefA.ID)
	if err := sub.Confirm(); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	order.SyncStatusFromSubOrders()
	if err := repo.UpdateSubOrder(ctx(), order, sub); err != nil {
		t.Fatalf("update sub-order: %v", err)
	}

	got, err := repo.FindByID(ctx(), order.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if s := got.SubOrderFor(chefA.ID).Status; s != domain.OrderStatusConfirmed {
		t.Errorf("chef A sub-order = %q, want confirmed", s)
	}
	if s := got.SubOrderFor(chefB.ID).Status; s != domain.OrderStatusPending {
		t.Errorf("chef B sub-order = %q, want pending (untouched)", s)
	}
	// Derived order status: least-advanced active sub-order is still pending.
	if got.Status != domain.OrderStatusPending {
		t.Errorf("order status = %q, want pending (derived)", got.Status)
	}
}

func TestOrderRepository_NotificationCounts(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 100)

	mk := func(code, status string) {
		t.Helper()
		o := buildOrder(customer.ID, code, domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, item.Price))
		if err := repo.Create(ctx(), o); err != nil {
			t.Fatalf("create %s: %v", code, err)
		}
		if status != domain.OrderStatusPending {
			if _, err := testDB.Exec(`UPDATE orders SET status = $2 WHERE id = $1`, o.ID, status); err != nil {
				t.Fatalf("set status: %v", err)
			}
			if _, err := testDB.Exec(`UPDATE sub_orders SET status = $2 WHERE order_id = $1`, o.ID, status); err != nil {
				t.Fatalf("set sub-order status: %v", err)
			}
		}
	}
	mk("N-PENDING", domain.OrderStatusPending)     // active + chef-pending
	mk("N-PREPARING", domain.OrderStatusPreparing) // active only
	mk("N-DELIVERED", domain.OrderStatusDelivered) // neither
	mk("N-CANCELLED", domain.OrderStatusCancelled) // neither

	active, err := repo.CountActiveByUser(ctx(), customer.ID)
	if err != nil || active != 2 {
		t.Errorf("active = %d (%v), want 2 (pending + preparing)", active, err)
	}
	pending, err := repo.CountPendingByChef(ctx(), chef.ID)
	if err != nil || pending != 1 {
		t.Errorf("chef pending = %d (%v), want 1", pending, err)
	}
	if n, _ := repo.CountPendingByChef(ctx(), 9999); n != 0 {
		t.Errorf("unknown chef pending = %d, want 0", n)
	}
}

func TestOrderRepository_UpdateStatus(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 10)

	order := buildOrder(customer.ID, "ORD-STATUS",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, 5))
	if err := repo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Advance through to delivered and persist.
	_ = order.Confirm()
	if err := repo.UpdateStatus(ctx(), order); err != nil {
		t.Fatalf("update status: %v", err)
	}
	got, _ := repo.FindByID(ctx(), order.ID)
	if got.Status != domain.OrderStatusConfirmed {
		t.Errorf("status = %q, want confirmed", got.Status)
	}

	// Cancel stamps cancelled_at.
	cancelled := buildOrder(customer.ID, "ORD-CANCEL",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, 5))
	_ = repo.Create(ctx(), cancelled)
	_ = cancelled.Cancel()
	if err := repo.UpdateStatus(ctx(), cancelled); err != nil {
		t.Fatalf("update cancel: %v", err)
	}
	got, _ = repo.FindByID(ctx(), cancelled.ID)
	if got.Status != domain.OrderStatusCancelled || got.CancelledAt == nil {
		t.Errorf("cancel not persisted: %+v", got)
	}
	// A customer cancel takes the sub-orders with it, in the same transaction.
	if s := got.SubOrders[0].Status; s != domain.OrderStatusCancelled {
		t.Errorf("sub-order after cancel = %q, want cancelled", s)
	}

	if err := repo.UpdateStatus(ctx(), buildOrder(customer.ID, "missing")); !errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("update missing = %v, want ErrOrderNotFound", err)
	}
}

// TestOrderRepository_UpdateSubOrder_StaleSnapshot: two chefs advance from
// stale reads of the same order — the derived order status must be recomputed
// from the current rows inside the lock, not trusted from either snapshot.
func TestOrderRepository_UpdateSubOrder_StaleSnapshot(t *testing.T) {
	resetDB(t)
	repo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chefA := seedChef(t, seedUser(t, "chefa@example.com").ID)
	menuA := seedMenu(t, chefA.ID)
	itemA := seedItem(t, menuA.ID, chefA.ID, 5, 10)
	chefB := seedChef(t, seedUser(t, "chefb@example.com").ID)
	menuB := seedMenu(t, chefB.ID)
	itemB := seedItem(t, menuB.ID, chefB.ID, 3, 10)

	order := buildOrder(customer.ID, "ORD-STALE",
		domain.NewOrderItem(itemA.ID, chefA.ID, itemA.Name, 1, 5),
		domain.NewOrderItem(itemB.ID, chefB.ID, itemB.Name, 1, 3))
	if err := repo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Both chefs read the order before either writes (stale snapshots).
	viewA, _ := repo.FindByID(ctx(), order.ID)
	viewB, _ := repo.FindByID(ctx(), order.ID)

	subA := viewA.SubOrderFor(chefA.ID)
	_ = subA.Confirm()
	viewA.SyncStatusFromSubOrders() // derives pending: B is pending in A's view
	if err := repo.UpdateSubOrder(ctx(), viewA, subA); err != nil {
		t.Fatalf("update A: %v", err)
	}

	subB := viewB.SubOrderFor(chefB.ID)
	_ = subB.Confirm()
	viewB.SyncStatusFromSubOrders() // derives pending: A is pending in B's *stale* view
	if err := repo.UpdateSubOrder(ctx(), viewB, subB); err != nil {
		t.Fatalf("update B: %v", err)
	}

	got, err := repo.FindByID(ctx(), order.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	// Both slices are confirmed, so the derived status must be confirmed even
	// though B's snapshot said pending.
	if got.Status != domain.OrderStatusConfirmed {
		t.Errorf("order status = %q, want confirmed (re-derived under lock)", got.Status)
	}
}
