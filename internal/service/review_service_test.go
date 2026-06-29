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
// dish (menuItemID) from chefID, and returns it.
func seedDeliveredOrder(t *testing.T, orders *fakeOrderRepo, userID, chefID, menuItemID int) *domain.Order {
	t.Helper()
	o := domain.NewOrder(userID, "123 St")
	o.Status = domain.OrderStatusDelivered
	o.Items = []*domain.OrderItem{{ChefID: chefID, MenuItemID: menuItemID, ItemName: "Soup", Quantity: 1}}
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
