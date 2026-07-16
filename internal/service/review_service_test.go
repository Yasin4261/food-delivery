package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeReviewRepo is an in-memory domain.ReviewRepository for service/handler
// tests. It enforces the per-order uniqueness but does not recompute aggregate
// ratings (that is the SQL adapter's job, covered by integration tests).
type fakeReviewRepo struct {
	reviews []*domain.Review
	nextID  int
}

func newFakeReviewRepo() *fakeReviewRepo { return &fakeReviewRepo{nextID: 1} }

func sameTarget(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (f *fakeReviewRepo) Create(_ context.Context, rv *domain.Review) error {
	for _, ex := range f.reviews {
		if ex.UserID == rv.UserID && ex.OrderID == rv.OrderID &&
			sameTarget(ex.ChefID, rv.ChefID) && sameTarget(ex.MenuItemID, rv.MenuItemID) {
			return domain.ErrReviewExists
		}
	}
	rv.ID = f.nextID
	f.nextID++
	cp := *rv
	f.reviews = append(f.reviews, &cp)
	return nil
}
func (f *fakeReviewRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Review, int, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.ChefID != nil && *rv.ChefID == chefID {
			out = append(out, rv)
		}
	}
	return out, len(out), nil
}
func (f *fakeReviewRepo) ListByMenuItem(_ context.Context, menuItemID, limit, offset int) ([]*domain.Review, int, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.MenuItemID != nil && *rv.MenuItemID == menuItemID {
			out = append(out, rv)
		}
	}
	return out, len(out), nil
}

// seedDeliveredOrder stores a delivered order owned by userID containing one
// dish (menuItemID) from chefID, with the chef's sub-order delivered too (the
// reviewability gate), and returns it.
func seedDeliveredOrder(t *testing.T, orders *fakeOrderRepo, userID, chefID, menuItemID int) *domain.Order {
	t.Helper()
	o := domain.NewOrder(userID, "123 St")
	o.Status = domain.OrderStatusDelivered
	o.Items = []*domain.OrderItem{{ChefID: chefID, MenuItemID: menuItemID, ItemName: "Soup", Quantity: 1}}
	o.SubOrders = []*domain.SubOrder{{ChefID: chefID, Status: domain.OrderStatusDelivered}}
	if err := orders.Create(context.Background(), o); err != nil {
		t.Fatalf("seed order: %v", err)
	}
	return o
}

func reviewFixture(t *testing.T) (*service.ReviewService, *fakeOrderRepo) {
	t.Helper()
	orders := newFakeOrderRepo()
	svc := service.NewReviewService(newFakeReviewRepo(), orders)
	return svc, orders
}

func TestReviewService_CreateChefAndProduct(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()
	order := seedDeliveredOrder(t, orders, 100, 1, 10)

	chef := 1
	rv, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &chef, Rating: 5, Comment: "great"})
	if err != nil {
		t.Fatalf("chef review: %v", err)
	}
	if rv.ID == 0 || rv.Comment == nil || *rv.Comment != "great" {
		t.Errorf("unexpected review: %+v", rv)
	}

	item := 10
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, MenuItemID: &item, Rating: 4}); err != nil {
		t.Fatalf("product review: %v", err)
	}
}

func TestReviewService_OnlyReviewWhatYouOrdered(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()
	order := seedDeliveredOrder(t, orders, 100, 1, 10)
	chef := 1

	// Not the order owner.
	if _, err := svc.Create(ctx, 200, service.CreateReviewInput{OrderID: order.ID, ChefID: &chef, Rating: 5}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("non-owner = %v, want ErrForbidden", err)
	}

	// A chef that was not part of the order.
	other := 99
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &other, Rating: 5}); !errors.Is(err, domain.ErrReviewTargetNotInOrder) {
		t.Errorf("foreign chef = %v, want ErrReviewTargetNotInOrder", err)
	}
	// A dish that was not part of the order.
	otherItem := 88
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, MenuItemID: &otherItem, Rating: 5}); !errors.Is(err, domain.ErrReviewTargetNotInOrder) {
		t.Errorf("foreign dish = %v, want ErrReviewTargetNotInOrder", err)
	}
}

func TestReviewService_OrderMustBeDelivered(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()
	// Pending order (not delivered).
	o := domain.NewOrder(100, "addr")
	o.Items = []*domain.OrderItem{{ChefID: 1, MenuItemID: 10}}
	_ = orders.Create(ctx, o)

	chef := 1
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, ChefID: &chef, Rating: 5}); !errors.Is(err, domain.ErrOrderNotReviewable) {
		t.Errorf("pending order = %v, want ErrOrderNotReviewable", err)
	}
}

func TestReviewService_ValidationAndDuplicate(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()
	order := seedDeliveredOrder(t, orders, 100, 1, 10)
	chef := 1

	// Invalid rating (caught before any lookup).
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &chef, Rating: 9}); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("bad rating = %v, want ErrInvalidRating", err)
	}
	// No target.
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, Rating: 3}); !errors.Is(err, domain.ErrInvalidReviewTarget) {
		t.Errorf("no target = %v, want ErrInvalidReviewTarget", err)
	}

	// First review ok, second (same chef, same order) is a duplicate.
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &chef, Rating: 5}); err != nil {
		t.Fatalf("first review: %v", err)
	}
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &chef, Rating: 4}); !errors.Is(err, domain.ErrReviewExists) {
		t.Errorf("duplicate = %v, want ErrReviewExists", err)
	}

	// Listing reflects the one chef review.
	list, _, _ := svc.ListForChef(ctx, chef, 20, 0)
	if len(list) != 1 {
		t.Errorf("chef reviews = %d, want 1", len(list))
	}
}

func (f *fakeReviewRepo) ListByUserOrder(_ context.Context, userID, orderID int) ([]*domain.Review, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.UserID == userID && rv.OrderID == orderID {
			cp := *rv
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeReviewRepo) ListByUser(_ context.Context, userID int) ([]*domain.Review, error) {
	out := make([]*domain.Review, 0)
	for _, rv := range f.reviews {
		if rv.UserID == userID {
			cp := *rv
			out = append(out, &cp)
		}
	}
	return out, nil
}

// The sub-order is the reviewability gate: in a multi-chef order a chef who
// delivered is reviewable immediately (order-level status still pending), a
// chef who declined never is — and neither are the declined chef's dishes.
func TestReviewService_SubOrderGating(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()

	o := domain.NewOrder(100, "123 St")
	o.Status = domain.OrderStatusPending // derived: chef 2 still pending
	o.Items = []*domain.OrderItem{
		{ChefID: 1, MenuItemID: 10, ItemName: "Soup", Quantity: 1},
		{ChefID: 2, MenuItemID: 20, ItemName: "Kebap", Quantity: 1},
		{ChefID: 3, MenuItemID: 30, ItemName: "Baklava", Quantity: 1},
	}
	o.SubOrders = []*domain.SubOrder{
		{ChefID: 1, Status: domain.OrderStatusDelivered},
		{ChefID: 2, Status: domain.OrderStatusPending},
		{ChefID: 3, Status: domain.OrderStatusCancelled}, // declined
	}
	if err := orders.Create(ctx, o); err != nil {
		t.Fatalf("seed: %v", err)
	}

	chef1, chef2, chef3 := 1, 2, 3
	dish1, dish3 := 10, 30

	// Delivered slice: chef and dish reviewable even though the order isn't
	// delivered yet.
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, ChefID: &chef1, Rating: 5}); err != nil {
		t.Errorf("delivered chef review = %v, want nil", err)
	}
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, MenuItemID: &dish1, Rating: 4}); err != nil {
		t.Errorf("delivered chef's dish review = %v, want nil", err)
	}

	// Still-cooking slice: not reviewable.
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, ChefID: &chef2, Rating: 5}); !errors.Is(err, domain.ErrOrderNotReviewable) {
		t.Errorf("pending chef review = %v, want ErrOrderNotReviewable", err)
	}

	// Declined slice: never reviewable — the food never arrived.
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, ChefID: &chef3, Rating: 1}); !errors.Is(err, domain.ErrOrderNotReviewable) {
		t.Errorf("declined chef review = %v, want ErrOrderNotReviewable", err)
	}
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: o.ID, MenuItemID: &dish3, Rating: 1}); !errors.Is(err, domain.ErrOrderNotReviewable) {
		t.Errorf("declined chef's dish review = %v, want ErrOrderNotReviewable", err)
	}
}

// ListForOrder returns only the caller's reviews for the order.
func TestReviewService_ListForOrder(t *testing.T) {
	svc, orders := reviewFixture(t)
	ctx := context.Background()
	order := seedDeliveredOrder(t, orders, 100, 1, 10)

	chefID, dishID := 1, 10
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, ChefID: &chefID, Rating: 5, Comment: "great"}); err != nil {
		t.Fatalf("chef review: %v", err)
	}
	if _, err := svc.Create(ctx, 100, service.CreateReviewInput{OrderID: order.ID, MenuItemID: &dishID, Rating: 4}); err != nil {
		t.Fatalf("dish review: %v", err)
	}

	mine, err := svc.ListForOrder(ctx, 100, order.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(mine) != 2 {
		t.Fatalf("history = %d reviews, want 2", len(mine))
	}

	// Another user asking about the same order sees nothing.
	other, err := svc.ListForOrder(ctx, 200, order.ID)
	if err != nil {
		t.Fatalf("other list: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("foreign history = %d reviews, want 0", len(other))
	}
}
