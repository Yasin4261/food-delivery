package domain_test

import (
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func intp(n int) *int { return &n }

func TestReview_Validate(t *testing.T) {
	chef := 1
	cases := []struct {
		name   string
		review domain.Review
		want   error
	}{
		{"valid chef review", domain.Review{ChefID: intp(chef), Rating: 5}, nil},
		{"valid product review", domain.Review{MenuItemID: intp(2), Rating: 1}, nil},
		{"rating too low", domain.Review{ChefID: intp(chef), Rating: 0}, domain.ErrInvalidRating},
		{"rating too high", domain.Review{ChefID: intp(chef), Rating: 6}, domain.ErrInvalidRating},
		{"no target", domain.Review{Rating: 3}, domain.ErrInvalidReviewTarget},
		{"both targets", domain.Review{ChefID: intp(chef), MenuItemID: intp(2), Rating: 3}, domain.ErrInvalidReviewTarget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.review.Validate(); !errors.Is(err, tc.want) {
				t.Errorf("Validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestReview_Targets(t *testing.T) {
	chef := domain.Review{ChefID: intp(1), Rating: 5}
	if !chef.TargetsChef() || chef.TargetsMenuItem() {
		t.Error("chef review target flags wrong")
	}
	dish := domain.Review{MenuItemID: intp(2), Rating: 5}
	if dish.TargetsChef() || !dish.TargetsMenuItem() {
		t.Error("product review target flags wrong")
	}
}

func TestOrder_HasMenuItem(t *testing.T) {
	o := domain.NewOrder(1, "addr")
	o.Items = []*domain.OrderItem{{MenuItemID: 10}, {MenuItemID: 20}}
	if !o.HasMenuItem(20) {
		t.Error("HasMenuItem(20) should be true")
	}
	if o.HasMenuItem(99) {
		t.Error("HasMenuItem(99) should be false")
	}
}
