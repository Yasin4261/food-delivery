//go:build integration

package repository_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// buildOrder assembles a pending order for userID containing the given items.
func buildOrder(userID int, code string, items ...*domain.OrderItem) *domain.Order {
	o := domain.NewOrder(userID, "123 St")
	o.OrderCode = code
	method := domain.PaymentMethodCash
	o.PaymentMethod = &method
	var subtotal float64
	for _, it := range items {
		o.Items = append(o.Items, it)
		subtotal += it.Subtotal
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

	if err := repo.UpdateStatus(ctx(), buildOrder(customer.ID, "missing")); !errors.Is(err, domain.ErrOrderNotFound) {
		t.Errorf("update missing = %v, want ErrOrderNotFound", err)
	}
}
