//go:build integration

package repository_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/repository"
)

// seedDeliveredOrder persists a delivered order for the customer containing the
// given dish, and returns its id (reviews reference a real order row).
func seedDeliveredOrder(t *testing.T, customerID, chefID int, item *domain.MenuItem, code string) int {
	t.Helper()
	repo := repository.NewOrderRepository(testDB)
	o := domain.NewOrder(customerID, "123 St")
	o.OrderCode = code
	method := domain.PaymentMethodCash
	o.PaymentMethod = &method
	o.Subtotal, o.TotalPrice = item.Price, item.Price
	o.Items = []*domain.OrderItem{domain.NewOrderItem(item.ID, chefID, item.Name, 1, item.Price)}
	if err := repo.Create(ctx(), o); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	return o.ID
}

// TestReviewRepository_RecomputesChefRating is the headline integration test:
// it proves the aggregate chefs.rating / total_reviews are recomputed in SQL.
func TestReviewRepository_RecomputesChefRating(t *testing.T) {
	resetDB(t)
	repo := repository.NewReviewRepository(testDB)
	chefRepo := repository.NewChefRepository(testDB)

	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 100)

	// Two customers leave chef reviews of 4 and 2 -> average 3.00, count 2.
	for i, rating := range []int{4, 2} {
		cust := seedUser(t, []string{"c1@e.com", "c2@e.com"}[i])
		orderID := seedDeliveredOrder(t, cust.ID, chef.ID, item, "ORD-CHEF-"+strconv.Itoa(i))
		rv := &domain.Review{UserID: cust.ID, OrderID: orderID, ChefID: &chef.ID, Rating: rating}
		if err := repo.Create(ctx(), rv); err != nil {
			t.Fatalf("create review: %v", err)
		}
	}

	got, err := chefRepo.FindByID(ctx(), chef.ID)
	if err != nil {
		t.Fatalf("reload chef: %v", err)
	}
	if got.Rating != 3.00 {
		t.Errorf("chef rating = %v, want 3.00", got.Rating)
	}
	if got.TotalReviews != 2 {
		t.Errorf("chef total_reviews = %d, want 2", got.TotalReviews)
	}
}

func TestReviewRepository_RecomputesItemRatingAndLists(t *testing.T) {
	resetDB(t)
	repo := repository.NewReviewRepository(testDB)
	itemRepo := repository.NewMenuItemRepository(testDB)

	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 100)
	cust := seedUser(t, "cust@example.com")
	orderID := seedDeliveredOrder(t, cust.ID, chef.ID, item, "ORD-ITEM-1")

	rv := &domain.Review{UserID: cust.ID, OrderID: orderID, MenuItemID: &item.ID, Rating: 5}
	if err := repo.Create(ctx(), rv); err != nil {
		t.Fatalf("create: %v", err)
	}
	if rv.ID == 0 {
		t.Fatal("create did not back-fill id")
	}

	got, _ := itemRepo.FindByID(ctx(), item.ID)
	if got.Rating != 5.00 || got.TotalReviews != 1 {
		t.Errorf("item rating/total = %v/%d, want 5.00/1", got.Rating, got.TotalReviews)
	}

	list, total, err := repo.ListByMenuItem(ctx(), item.ID, 20, 0)
	if total != 1 {
		t.Errorf("review total = %d, want 1", total)
	}
	if err != nil || len(list) != 1 {
		t.Errorf("list by item = %d, %v", len(list), err)
	}
}

func TestReviewRepository_DuplicateRejected(t *testing.T) {
	resetDB(t)
	repo := repository.NewReviewRepository(testDB)
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 100)
	cust := seedUser(t, "cust@example.com")
	orderID := seedDeliveredOrder(t, cust.ID, chef.ID, item, "ORD-DUP-1")

	first := &domain.Review{UserID: cust.ID, OrderID: orderID, ChefID: &chef.ID, Rating: 5}
	if err := repo.Create(ctx(), first); err != nil {
		t.Fatalf("first review: %v", err)
	}
	dup := &domain.Review{UserID: cust.ID, OrderID: orderID, ChefID: &chef.ID, Rating: 3}
	if err := repo.Create(ctx(), dup); !errors.Is(err, domain.ErrReviewExists) {
		t.Errorf("duplicate = %v, want ErrReviewExists", err)
	}
}

func TestReviewRepository_ListByUserOrder(t *testing.T) {
	resetDB(t)
	repo := repository.NewReviewRepository(testDB)
	customer := seedUser(t, "cust@example.com")
	stranger := seedUser(t, "stranger@example.com")
	chef := seedChef(t, seedUser(t, "chef@example.com").ID)
	menu := seedMenu(t, chef.ID)
	item := seedItem(t, menu.ID, chef.ID, 5, 10)

	order := buildOrder(customer.ID, "ORD-HIST",
		domain.NewOrderItem(item.ID, chef.ID, item.Name, 1, 5))
	if err := repository.NewOrderRepository(testDB).Create(ctx(), order); err != nil {
		t.Fatalf("create order: %v", err)
	}

	comment := "tasty"
	for _, rv := range []*domain.Review{
		{UserID: customer.ID, OrderID: order.ID, ChefID: &chef.ID, Rating: 5, Comment: &comment},
		{UserID: customer.ID, OrderID: order.ID, MenuItemID: &item.ID, Rating: 4},
	} {
		if err := repo.Create(ctx(), rv); err != nil {
			t.Fatalf("create review: %v", err)
		}
	}

	mine, err := repo.ListByUserOrder(ctx(), customer.ID, order.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(mine) != 2 || mine[0].ChefID == nil || mine[1].MenuItemID == nil {
		t.Errorf("history wrong: %+v", mine)
	}

	// Scoped by user: a stranger sees nothing for the same order.
	foreign, err := repo.ListByUserOrder(ctx(), stranger.ID, order.ID)
	if err != nil {
		t.Fatalf("foreign list: %v", err)
	}
	if len(foreign) != 0 {
		t.Errorf("foreign history = %d, want 0", len(foreign))
	}
}
