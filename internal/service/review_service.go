package service

import (
	"context"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ReviewService implements the rating use cases: a customer reviews a chef or a
// dish from one of their delivered orders. It depends only on domain ports.
type ReviewService struct {
	reviews domain.ReviewRepository
	orders  domain.OrderRepository
}

// NewReviewService builds a ReviewService.
func NewReviewService(reviews domain.ReviewRepository, orders domain.OrderRepository) *ReviewService {
	return &ReviewService{reviews: reviews, orders: orders}
}

// CreateReviewInput is the data needed to leave a review. Exactly one of ChefID
// / MenuItemID must be set.
type CreateReviewInput struct {
	OrderID    int
	ChefID     *int
	MenuItemID *int
	Rating     int
	Comment    string
}

// Create validates the review, enforces that the caller owns the (delivered)
// order and actually ordered the reviewed chef/dish, then persists it (which
// also recomputes the aggregate rating).
func (s *ReviewService) Create(ctx context.Context, userID int, in CreateReviewInput) (*domain.Review, error) {
	review := &domain.Review{
		UserID:     userID,
		OrderID:    in.OrderID,
		ChefID:     in.ChefID,
		MenuItemID: in.MenuItemID,
		Rating:     in.Rating,
		Comment:    optional(in.Comment),
	}
	if err := review.Validate(); err != nil {
		return nil, err
	}

	order, err := s.orders.FindByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	if order.Status != domain.OrderStatusDelivered {
		return nil, domain.ErrOrderNotReviewable
	}

	// The reviewed target must actually appear in the order.
	if review.TargetsChef() && !order.HasChef(*review.ChefID) {
		return nil, domain.ErrReviewTargetNotInOrder
	}
	if review.TargetsMenuItem() && !order.HasMenuItem(*review.MenuItemID) {
		return nil, domain.ErrReviewTargetNotInOrder
	}

	if err := s.reviews.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

// ListForChef returns a page of a chef's reviews and the total.
func (s *ReviewService) ListForChef(ctx context.Context, chefID, limit, offset int) ([]*domain.Review, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.reviews.ListByChef(ctx, chefID, limit, offset)
}

// ListForMenuItem returns a page of a dish's reviews and the total.
func (s *ReviewService) ListForMenuItem(ctx context.Context, menuItemID, limit, offset int) ([]*domain.Review, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.reviews.ListByMenuItem(ctx, menuItemID, limit, offset)
}
