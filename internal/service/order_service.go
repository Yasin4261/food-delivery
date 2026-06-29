package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// OrderService implements the ordering use cases: a customer places and tracks
// orders (which may span several chefs), and a chef advances the status of
// orders containing their items. It depends only on domain ports.
type OrderService struct {
	orders domain.OrderRepository
	items  domain.MenuItemRepository
	chefs  domain.ChefRepository
}

// NewOrderService builds an OrderService.
func NewOrderService(orders domain.OrderRepository, items domain.MenuItemRepository, chefs domain.ChefRepository) *OrderService {
	return &OrderService{orders: orders, items: items, chefs: chefs}
}

// OrderLineInput is one requested dish in a new order.
type OrderLineInput struct {
	MenuItemID          int
	Quantity            int
	SpecialInstructions string
}

// PlaceOrderInput is the data needed to create an order.
type PlaceOrderInput struct {
	DeliveryAddress   string
	DeliveryCity      string
	DeliveryLatitude  *float64
	DeliveryLongitude *float64
	PaymentMethod     string
	CustomerNotes     string
	Lines             []OrderLineInput
}

// PlaceOrder validates the cart, snapshots each dish's price, computes totals,
// persists the order with its items atomically, then decrements stock for
// limited items. The order can contain items from multiple chefs.
func (s *OrderService) PlaceOrder(ctx context.Context, userID int, in PlaceOrderInput) (*domain.Order, error) {
	in.DeliveryAddress = strings.TrimSpace(in.DeliveryAddress)
	in.PaymentMethod = strings.TrimSpace(in.PaymentMethod)

	if in.DeliveryAddress == "" {
		return nil, ValidationError{Msg: "delivery_address is required"}
	}
	if !domain.ValidPaymentMethod(in.PaymentMethod) {
		return nil, ValidationError{Msg: "payment_method must be cash or card"}
	}
	if len(in.Lines) == 0 {
		return nil, domain.ErrEmptyOrder
	}

	order := domain.NewOrder(userID, in.DeliveryAddress)
	order.OrderCode = newOrderCode()
	order.PaymentMethod = &in.PaymentMethod
	order.DeliveryCity = optional(in.DeliveryCity)
	order.DeliveryLatitude = in.DeliveryLatitude
	order.DeliveryLongitude = in.DeliveryLongitude
	order.CustomerNotes = optional(in.CustomerNotes)

	var subtotal float64
	for _, line := range in.Lines {
		if line.Quantity < 1 {
			return nil, ValidationError{Msg: "each item quantity must be at least 1"}
		}
		item, err := s.items.FindByID(ctx, line.MenuItemID)
		if err != nil {
			return nil, err
		}
		if !item.IsActive || !item.IsAvailable {
			return nil, domain.ErrItemNotOrderable
		}
		if !item.InStock(line.Quantity) {
			return nil, domain.ErrItemOutOfStock
		}

		oi := domain.NewOrderItem(item.ID, item.ChefID, item.Name, line.Quantity, item.Price)
		oi.SpecialInstructions = optional(line.SpecialInstructions)
		order.Items = append(order.Items, oi)
		subtotal += oi.Subtotal
	}

	order.Subtotal = subtotal
	order.TotalPrice = subtotal + order.DeliveryFee + order.ServiceFee + order.Tax - order.Discount

	if err := s.orders.Create(ctx, order); err != nil {
		return nil, err
	}

	// Decrement stock for limited items. Unlimited items track no quantity.
	for _, oi := range order.Items {
		item, err := s.items.FindByID(ctx, oi.MenuItemID)
		if err != nil {
			return nil, err
		}
		if item.IsUnlimited {
			continue
		}
		if err := s.items.DecrementStock(ctx, oi.MenuItemID, oi.Quantity); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// GetForCustomer returns an order owned by the customer, or ErrForbidden.
func (s *OrderService) GetForCustomer(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return order, nil
}

// ListForCustomer returns a page of the customer's order history and the total.
func (s *OrderService) ListForCustomer(ctx context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.orders.ListByUser(ctx, userID, limit, offset)
}

// CancelForCustomer cancels the customer's own order (pending/confirmed only).
func (s *OrderService) CancelForCustomer(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	order, err := s.GetForCustomer(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if err := order.Cancel(); err != nil {
		return nil, err
	}
	if err := s.orders.UpdateStatus(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// ListForChef returns a page of orders containing the chef's items (scoped to
// that chef) and the total.
func (s *OrderService) ListForChef(ctx context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	limit, offset = normalisePaging(limit, offset)
	return s.orders.ListByChef(ctx, chef.ID, limit, offset)
}

// Chef status actions accepted by AdvanceForChef.
const (
	OrderActionConfirm    = "confirm"
	OrderActionPreparing  = "preparing"
	OrderActionReady      = "ready"
	OrderActionDelivering = "delivering"
	OrderActionDelivered  = "delivered"
	OrderActionDecline    = "decline"
)

// AdvanceForChef applies a status transition to an order the chef participates
// in. "confirm" accepts the order; "decline" cancels it; the rest move it along
// the lifecycle. Illegal transitions return ErrInvalidStatusTransition.
func (s *OrderService) AdvanceForChef(ctx context.Context, userID, orderID int, action string) (*domain.Order, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if !order.HasChef(chef.ID) {
		return nil, domain.ErrForbidden
	}

	if err := applyChefAction(order, action); err != nil {
		return nil, err
	}
	if err := s.orders.UpdateStatus(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func applyChefAction(order *domain.Order, action string) error {
	switch action {
	case OrderActionConfirm:
		return order.Confirm()
	case OrderActionPreparing:
		return order.StartPreparing()
	case OrderActionReady:
		return order.MarkReady()
	case OrderActionDelivering:
		return order.StartDelivering()
	case OrderActionDelivered:
		return order.MarkDelivered()
	case OrderActionDecline:
		return order.Cancel()
	default:
		return ValidationError{Msg: "unknown action: must be confirm, preparing, ready, delivering, delivered or decline"}
	}
}

// newOrderCode returns a unique-enough human-facing order code.
func newOrderCode() string {
	var b [5]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("ORD-%s", strings.ToUpper(hex.EncodeToString(b[:])))
}
