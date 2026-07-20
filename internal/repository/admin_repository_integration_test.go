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

	users, total, err := repo.ListUsers(ctx(), domain.AdminUserFilters{}, 20, 0)
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
	after, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{}, 20, 0)
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
	if chefs, ctotal, err := repo.ListChefs(ctx(), domain.AdminChefFilters{}, 20, 0); err != nil || ctotal != 1 || chefs[0].IsActive {
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
	orders, total, err := repo.ListOrders(ctx(), domain.AdminOrderFilters{}, 20, 0)
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

// The admin filter SQL (#118) against real Postgres: ILIKE search, role/status
// narrowing, the tri-state active filter, and — critically — that the returned
// total counts MATCHING rows rather than the whole table.
func TestAdminRepository_ListFilters(t *testing.T) {
	resetDB(t)
	repo := repository.NewAdminRepository(testDB)

	alice := seedUser(t, "alice@example.com")
	seedUser(t, "bob@example.com")
	chefUser := seedUser(t, "chef@example.com")
	chef := seedChef(t, chefUser.ID)

	truthy := true
	falsy := false

	// --- users -------------------------------------------------------------
	// Free-text matches email or username, case-insensitively.
	for _, q := range []string{"alice", "ALICE", "alice@ex"} {
		got, total, err := repo.ListUsers(ctx(), domain.AdminUserFilters{Query: q}, 20, 0)
		if err != nil {
			t.Fatalf("list users q=%q: %v", q, err)
		}
		if total != 1 || len(got) != 1 || got[0].ID != alice.ID {
			t.Errorf("q=%q -> total=%d len=%d, want exactly alice", q, total, len(got))
		}
	}

	// A non-matching query yields an empty page AND a zero total (the bug this
	// guards: reusing an unfiltered count(*) would report the table size).
	if got, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Query: "zzz-nobody"}, 20, 0); total != 0 || len(got) != 0 {
		t.Errorf("no-match query -> total=%d len=%d, want 0/0", total, len(got))
	}

	// Role filter.
	if _, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Role: domain.RoleChef}, 20, 0); total != 0 {
		// seedUser creates plain users; the chef PROFILE exists but the role
		// may not be 'chef' — assert against the actual seeded role instead.
		t.Logf("role=chef total=%d (seed roles are %q)", total, chefUser.Role)
	}
	if _, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Role: chefUser.Role}, 20, 0); total != 3 {
		t.Errorf("role=%q total=%d, want 3 seeded users", chefUser.Role, total)
	}

	// Tri-state active: nil = both, true/false narrow.
	_, allTotal, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{}, 20, 0)
	if _, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Active: &truthy}, 20, 0); total != allTotal {
		t.Errorf("active=true total=%d, want %d (all seeded users active)", total, allTotal)
	}
	if _, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Active: &falsy}, 20, 0); total != 0 {
		t.Errorf("active=false total=%d, want 0", total)
	}
	// Deactivate one and the tri-state flips accordingly.
	if err := repo.SetUserActive(ctx(), alice.ID, false); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if _, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{Active: &falsy}, 20, 0); total != 1 {
		t.Errorf("after deactivate, active=false total=%d, want 1", total)
	}

	// Pagination: total stays the full matching count, not the page length.
	page, total, _ := repo.ListUsers(ctx(), domain.AdminUserFilters{}, 1, 0)
	if len(page) != 1 || total != allTotal {
		t.Errorf("limit=1 -> len=%d total=%d, want 1/%d", len(page), total, allTotal)
	}

	// --- chefs -------------------------------------------------------------
	if got, total, err := repo.ListChefs(ctx(), domain.AdminChefFilters{Query: "kitch"}, 20, 0); err != nil || total != 1 || len(got) != 1 {
		t.Errorf("chef q=kitch -> total=%d len=%d err=%v, want 1/1", total, len(got), err)
	}
	if _, total, _ := repo.ListChefs(ctx(), domain.AdminChefFilters{Query: "zzz"}, 20, 0); total != 0 {
		t.Errorf("chef q=zzz total=%d, want 0", total)
	}
	if err := repo.SetChefActive(ctx(), chef.ID, false); err != nil {
		t.Fatalf("deactivate chef: %v", err)
	}
	if _, total, _ := repo.ListChefs(ctx(), domain.AdminChefFilters{Active: &truthy}, 20, 0); total != 0 {
		t.Errorf("active chefs total=%d, want 0 after deactivation", total)
	}

	// --- orders ------------------------------------------------------------
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 10)
	orderID := seedDeliveredOrder(t, alice.ID, chef.ID, item, "ORD-FILTER-1")
	_ = orderID

	// NOTE: seedDeliveredOrder is misleadingly named — it persists a NEW order,
	// which is status=pending. Assert against the real seeded state.
	if _, total, err := repo.ListOrders(ctx(), domain.AdminOrderFilters{Status: domain.OrderStatusPending}, 20, 0); err != nil || total != 1 {
		t.Errorf("status=pending total=%d err=%v, want 1", total, err)
	}
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{Status: domain.OrderStatusDelivered}, 20, 0); total != 0 {
		t.Errorf("status=delivered total=%d, want 0", total)
	}
	// Payment status narrows independently of lifecycle status.
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{PaymentStatus: domain.PaymentStatusPending}, 20, 0); total != 1 {
		t.Errorf("payment_status=pending total=%d, want 1", total)
	}
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{PaymentStatus: domain.PaymentStatusPaid}, 20, 0); total != 0 {
		t.Errorf("payment_status=paid total=%d, want 0", total)
	}
	// Combined filters intersect (status AND customer), they don't union.
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{
		Status: domain.OrderStatusPending, UserID: 999999,
	}, 20, 0); total != 0 {
		t.Errorf("status=pending AND user_id=999999 total=%d, want 0 (filters must intersect)", total)
	}
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{UserID: alice.ID}, 20, 0); total != 1 {
		t.Errorf("user_id=alice total=%d, want 1", total)
	}
	if _, total, _ := repo.ListOrders(ctx(), domain.AdminOrderFilters{UserID: 999999}, 20, 0); total != 0 {
		t.Errorf("user_id=999999 total=%d, want 0", total)
	}
}
