//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// setOrderState forces an order's status/payment_status directly (bypassing the
// transition methods) to set up earnings scenarios.
func setOrderState(t *testing.T, orderID int, status, payment string) {
	t.Helper()
	if _, err := testDB.Exec(`UPDATE orders SET status = $2, payment_status = $3 WHERE id = $1`, orderID, status, payment); err != nil {
		t.Fatalf("set order state: %v", err)
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
