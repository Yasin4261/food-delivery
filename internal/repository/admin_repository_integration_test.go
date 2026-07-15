//go:build integration

package repository_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

func TestAdminRepository_UsersAndActiveToggles(t *testing.T) {
	resetDB(t)
	repo := repository.NewAdminRepository(testDB)
	u1 := seedUser(t, "a@example.com")
	u2 := seedUser(t, "b@example.com")
	chef := seedChef(t, u2.ID)

	users, total, err := repo.ListUsers(ctx(), 20, 0)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Fatalf("users = %d/%d, want 2", len(users), total)
	}
	for _, u := range users {
		if u.PasswordHash == "" {
			t.Error("adapter should return the hash; the service clears it")
		}
	}

	// Deactivate a user; it still appears in the admin listing (unlike browse),
	// but now flagged inactive.
	if err := repo.SetUserActive(ctx(), u1.ID, false); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}
	after, total, _ := repo.ListUsers(ctx(), 20, 0)
	if total != 2 {
		t.Errorf("admin should still list inactive users: total = %d, want 2", total)
	}
	for _, u := range after {
		if u.ID == u1.ID && u.IsActive {
			t.Error("user should be inactive after SetUserActive(false)")
		}
	}

	// Deactivate the chef -> browse (List, active-only) drops it.
	if err := repo.SetChefActive(ctx(), chef.ID, false); err != nil {
		t.Fatalf("deactivate chef: %v", err)
	}
	chefRepo := repository.NewChefRepository(testDB)
	if _, n, _ := chefRepo.List(ctx(), domain.ChefListFilters{}, 20, 0); n != 0 {
		t.Errorf("deactivated chef still browseable: %d", n)
	}

	// Admin listing still shows the deactivated chef.
	if chefs, ctotal, err := repo.ListChefs(ctx(), 20, 0); err != nil || ctotal != 1 || chefs[0].IsActive {
		t.Errorf("admin chef listing wrong: total=%d err=%v", ctotal, err)
	}

	// Unknown ids.
	if err := repo.SetUserActive(ctx(), 9999, true); err != domain.ErrUserNotFound {
		t.Errorf("unknown user = %v, want ErrUserNotFound", err)
	}
	if err := repo.SetChefActive(ctx(), 9999, true); err != domain.ErrChefNotFound {
		t.Errorf("unknown chef = %v, want ErrChefNotFound", err)
	}
}

func TestAdminRepository_StatsAndOrders(t *testing.T) {
	resetDB(t)
	repo := repository.NewAdminRepository(testDB)
	orderRepo := repository.NewOrderRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 20, 100)

	// One delivered+paid order (counts toward GMV + top chefs), one pending.
	paid := buildOrder(customer.ID, "ORD-PAID",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 2, item.Price)) // 40
	if err := orderRepo.Create(ctx(), paid); err != nil {
		t.Fatalf("create paid: %v", err)
	}
	setOrderState(t, paid.ID, domain.OrderStatusDelivered, domain.PaymentStatusPaid)

	pending := buildOrder(customer.ID, "ORD-PEND",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, item.Price))
	_ = orderRepo.Create(ctx(), pending)

	stats, err := repo.Stats(ctx())
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.TotalUsers != 2 || stats.TotalChefs != 1 || stats.ActiveChefs != 1 {
		t.Errorf("counts = %+v", stats)
	}
	if stats.TotalOrders != 2 || stats.DeliveredOrders != 1 || stats.OrdersToday != 2 {
		t.Errorf("order counts = %+v", stats)
	}
	if stats.GMV != 40 {
		t.Errorf("GMV = %v, want 40 (delivered+paid only)", stats.GMV)
	}
	if len(stats.TopChefs) != 1 || stats.TopChefs[0].ChefID != chef.ID || stats.TopChefs[0].Revenue != 40 || stats.TopChefs[0].Orders != 1 {
		t.Errorf("top chefs = %+v", stats.TopChefs)
	}

	// Order overview returns all orders, newest first, with items loaded.
	orders, total, err := repo.ListOrders(ctx(), 20, 0)
	if err != nil {
		t.Fatalf("list orders: %v", err)
	}
	if total != 2 || len(orders) != 2 || orders[0].ID < orders[1].ID {
		t.Errorf("order overview wrong: total=%d newest-first?", total)
	}
	if len(orders[0].Items) == 0 {
		t.Error("order overview should load items")
	}
}
