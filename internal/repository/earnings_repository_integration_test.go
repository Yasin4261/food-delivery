//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// setOrderState forces an order's status/payment_status directly (bypassing the
// transition methods) to set up earnings scenarios. Sub-orders follow the
// order-level status, as the backfill migration does.
func setOrderState(t *testing.T, orderID int, status, payment string) {
	t.Helper()
	if _, err := testDB.Exec(`UPDATE orders SET status = $2, payment_status = $3 WHERE id = $1`, orderID, status, payment); err != nil {
		t.Fatalf("set order state: %v", err)
	}
	if _, err := testDB.Exec(`UPDATE sub_orders SET status = $2 WHERE order_id = $1`, orderID, status); err != nil {
		t.Fatalf("set sub-order state: %v", err)
	}
}

// TestEarningsRepository_OnlyDeliveredAndPaid is the headline test: earnings
// must count only orders that are both delivered and paid.
func TestEarningsRepository_OnlyDeliveredAndPaid(t *testing.T) {
	resetDB(t)
	orderRepo := repository.NewOrderRepository(testDB)
	earnRepo := repository.NewEarningsRepository(testDB)

	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 100)

	// Counts: delivered + paid, 2 units -> 10.00.
	counted := buildOrder(customer.ID, "ORD-PAID",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 2, item.Price))
	if err := orderRepo.Create(ctx(), counted); err != nil {
		t.Fatalf("create counted: %v", err)
	}
	setOrderState(t, counted.ID, domain.OrderStatusDelivered, domain.PaymentStatusPaid)

	// Excluded: delivered but unpaid.
	unpaid := buildOrder(customer.ID, "ORD-UNPAID",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 3, item.Price))
	_ = orderRepo.Create(ctx(), unpaid)
	setOrderState(t, unpaid.ID, domain.OrderStatusDelivered, domain.PaymentStatusPending)

	// Excluded: paid but not delivered.
	pending := buildOrder(customer.ID, "ORD-PENDING",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 4, item.Price))
	_ = orderRepo.Create(ctx(), pending)
	setOrderState(t, pending.ID, domain.OrderStatusPending, domain.PaymentStatusPaid)

	got, err := earnRepo.ChefEarnings(ctx(), chef.ID, nil)
	if err != nil {
		t.Fatalf("earnings: %v", err)
	}
	if got.TotalEarnings != 10.00 {
		t.Errorf("total earnings = %v, want 10.00 (only delivered+paid)", got.TotalEarnings)
	}
	if got.DeliveredOrders != 1 {
		t.Errorf("delivered orders = %d, want 1", got.DeliveredOrders)
	}
	if got.ItemsSold != 2 {
		t.Errorf("items sold = %d, want 2", got.ItemsSold)
	}
}

// TestEarningsRepository_PerChefSubOrder: in a multi-chef paid order, a chef
// earns as soon as *their* sub-order is delivered — the other chef, still
// preparing, earns nothing yet.
func TestEarningsRepository_PerChefSubOrder(t *testing.T) {
	resetDB(t)
	orderRepo := repository.NewOrderRepository(testDB)
	earnRepo := repository.NewEarningsRepository(testDB)

	customer := seedUser(t, "cust@example.com")
	chefA := seedChef(t, seedUser(t, "chefa@example.com").ID)
	menuA := seedMenu(t, chefA.ID)
	itemA := seedItem(t, menuA.ID, chefA.ID, 5, 100)
	chefB := seedChef(t, seedUser(t, "chefb@example.com").ID)
	menuB := seedMenu(t, chefB.ID)
	itemB := seedItem(t, menuB.ID, chefB.ID, 3, 100)

	order := buildOrder(customer.ID, "ORD-SPLIT",
		domain.NewOrderItem(itemA.ID, chefA.ID, itemA.Name, 2, itemA.Price), // 10
		domain.NewOrderItem(itemB.ID, chefB.ID, itemB.Name, 1, itemB.Price)) // 3
	if err := orderRepo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Order paid (card), chef A delivered their slice, chef B still preparing.
	if _, err := testDB.Exec(`UPDATE orders SET payment_status = 'paid' WHERE id = $1`, order.ID); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	if _, err := testDB.Exec(`UPDATE sub_orders SET status = 'delivered' WHERE order_id = $1 AND chef_id = $2`, order.ID, chefA.ID); err != nil {
		t.Fatalf("deliver chef A: %v", err)
	}
	if _, err := testDB.Exec(`UPDATE sub_orders SET status = 'preparing' WHERE order_id = $1 AND chef_id = $2`, order.ID, chefB.ID); err != nil {
		t.Fatalf("prepare chef B: %v", err)
	}

	a, err := earnRepo.ChefEarnings(ctx(), chefA.ID, nil)
	if err != nil {
		t.Fatalf("chef A earnings: %v", err)
	}
	if a.TotalEarnings != 10 || a.DeliveredOrders != 1 || a.ItemsSold != 2 {
		t.Errorf("chef A = %+v, want 10/1/2 (their slice delivered)", a)
	}

	b, err := earnRepo.ChefEarnings(ctx(), chefB.ID, nil)
	if err != nil {
		t.Fatalf("chef B earnings: %v", err)
	}
	if b.TotalEarnings != 0 {
		t.Errorf("chef B = %+v, want 0 (their slice not delivered)", b)
	}
}

func TestEarningsRepository_EmptyIsZero(t *testing.T) {
	resetDB(t)
	earnRepo := repository.NewEarningsRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)

	got, err := earnRepo.ChefEarnings(ctx(), chef.ID, nil)
	if err != nil {
		t.Fatalf("earnings: %v", err)
	}
	if got.TotalEarnings != 0 || got.DeliveredOrders != 0 || got.ItemsSold != 0 {
		t.Errorf("empty earnings should be zero, got %+v", got)
	}
}

// Net earnings (#65): delivery fees are kept in full, the snapshotted
// commission is deducted, and only delivered slices on paid orders count.
func TestEarningsRepository_FeesAndCommission(t *testing.T) {
	resetDB(t)
	orderRepo := repository.NewOrderRepository(testDB)
	earnRepo := repository.NewEarningsRepository(testDB)

	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 50, 100)

	order := buildOrder(customer.ID, "ORD-FEES",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 2, item.Price)) // food 100
	order.SubOrders[0].DeliveryFee = 25
	order.SubOrders[0].Commission = 10
	order.SubOrders[0].Tip = 8
	if err := orderRepo.Create(ctx(), order); err != nil {
		t.Fatalf("create: %v", err)
	}
	setOrderState(t, order.ID, domain.OrderStatusDelivered, domain.PaymentStatusPaid)

	// A second, undelivered order must not count.
	pending := buildOrder(customer.ID, "ORD-PEND",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, item.Price))
	pending.SubOrders[0].DeliveryFee = 99
	pending.SubOrders[0].Commission = 99
	pending.SubOrders[0].Tip = 99
	_ = orderRepo.Create(ctx(), pending)

	got, err := earnRepo.ChefEarnings(ctx(), chef.ID, nil)
	if err != nil {
		t.Fatalf("earnings: %v", err)
	}
	if got.TotalEarnings != 100 || got.DeliveryFees != 25 || got.Commission != 10 || got.Tips != 8 {
		t.Errorf("gross/fees/commission/tips = %v/%v/%v/%v, want 100/25/10/8",
			got.TotalEarnings, got.DeliveryFees, got.Commission, got.Tips)
	}
	if got.NetEarnings != 123 { // 100 + 25 + 8 - 10
		t.Errorf("net = %v, want 123", got.NetEarnings)
	}
}
