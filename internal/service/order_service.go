package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// OrderService implements the ordering use cases: a customer places and tracks
// orders (which may span several chefs), and a chef advances the status of
// orders containing their items. It depends only on domain ports.
type OrderService struct {
	orders    domain.OrderRepository
	items     domain.MenuItemRepository
	chefs     domain.ChefRepository
	addresses domain.AddressRepository   // resolves saved address_id at placement; may be nil
	hours     domain.ChefHoursRepository // rejects orders outside working hours; may be nil
	refunder  domain.PaymentRefunder     // refunds card payments on cancel; may be nil
	notifier  *OrderNotifier             // fire-and-forget order emails; may be nil
	loc       *time.Location             // platform TZ for the working-hours check
	policy    domain.FeePolicy           // delivery fees + commission (#65); zero value = free
	etaWindow time.Duration              // ETA stamped when a chef accepts (#92); 0 disables
}

// SetETAWindow configures the prep+delivery window used to stamp an order's
// estimated delivery time when a chef accepts it. Zero disables ETAs.
func (s *OrderService) SetETAWindow(d time.Duration) { s.etaWindow = d }

// NewOrderService builds an OrderService. addresses resolves a saved
// address_id at placement (nil rejects address_id orders); hours enforces
// chef working hours at placement (nil disables the check; loc nil = UTC);
// refunder handles gateway refunds when a paid card order is cancelled (nil
// disables refunds); notifier sends order emails (nil disables them).
func NewOrderService(orders domain.OrderRepository, items domain.MenuItemRepository, chefs domain.ChefRepository, addresses domain.AddressRepository, hours domain.ChefHoursRepository, loc *time.Location, policy domain.FeePolicy, refunder domain.PaymentRefunder, notifier *OrderNotifier) *OrderService {
	if loc == nil {
		loc = time.UTC
	}
	return &OrderService{orders: orders, items: items, chefs: chefs, addresses: addresses, hours: hours, loc: loc, policy: policy, refunder: refunder, notifier: notifier}
}

// OrderLineInput is one requested dish in a new order.
type OrderLineInput struct {
	MenuItemID          int
	Quantity            int
	SpecialInstructions string
}

// PlaceOrderInput is the data needed to create an order. AddressID selects a
// saved address-book entry instead of a one-off DeliveryAddress; the entry's
// text is snapshotted onto the order, so later edits to the book never touch
// order history.
type PlaceOrderInput struct {
	AddressID         *int
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

	// A saved address (owner-checked) fills the delivery fields; a one-off
	// delivery_address in the same request would be ambiguous, so it is
	// rejected rather than silently overridden.
	if in.AddressID != nil {
		if s.addresses == nil {
			return nil, ValidationError{Msg: "address_id is not supported"}
		}
		if in.DeliveryAddress != "" {
			return nil, ValidationError{Msg: "provide either address_id or delivery_address, not both"}
		}
		saved, err := s.addresses.FindByID(ctx, *in.AddressID)
		if err != nil {
			return nil, err
		}
		if saved.UserID != userID {
			return nil, domain.ErrForbidden
		}
		in.DeliveryAddress = saved.Address
		if saved.City != nil {
			in.DeliveryCity = *saved.City
		}
		in.DeliveryLatitude = saved.Latitude
		in.DeliveryLongitude = saved.Longitude
	}

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

	// One sub-order per participating chef (in first-seen item order): the
	// chef-scoped slice that carries its own status lifecycle.
	for _, oi := range order.Items {
		if s := order.SubOrderFor(oi.ChefID); s != nil {
			s.Subtotal += oi.Subtotal
			continue
		}
		order.SubOrders = append(order.SubOrders, domain.NewSubOrder(oi.ChefID, oi.Subtotal))
	}

	// Money model (#65), snapshotted per slice: the customer pays a
	// distance-based delivery fee per chef (base only when either side lacks
	// coordinates); the platform's commission comes out of the chef's food
	// subtotal, never the customer's total. Working hours are checked in the
	// same pass so each chef is fetched once.
	now := time.Now().In(s.loc)
	for _, sub := range order.SubOrders {
		if s.hours != nil {
			schedule, err := s.hours.ListByChef(ctx, sub.ChefID)
			if err != nil {
				return nil, err
			}
			if !domain.IsOpenAt(schedule, now) {
				return nil, domain.ErrChefClosed
			}
		}

		// FindByID filters to active chefs, so a chef deactivated by an admin
		// (#69) — whose dishes are already hidden from browse/search — makes a
		// stale-cart order fail here with ErrChefNotFound.
		chef, err := s.chefs.FindByID(ctx, sub.ChefID)
		if err != nil {
			return nil, err
		}
		distance := -1.0 // unknown: base fee only
		if chef.HasLocation() && in.DeliveryLatitude != nil && in.DeliveryLongitude != nil {
			distance = domain.CalculateDistance(*chef.KitchenLatitude, *chef.KitchenLongitude, *in.DeliveryLatitude, *in.DeliveryLongitude)
		}
		sub.DeliveryFee = s.policy.DeliveryFee(distance)
		sub.Commission = s.policy.Commission(sub.Subtotal)
		order.DeliveryFee += sub.DeliveryFee
	}
	order.DeliveryFee = domain.RoundMoney(order.DeliveryFee)

	order.TotalPrice = domain.RoundMoney(subtotal + order.DeliveryFee + order.ServiceFee + order.Tax - order.Discount)

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

	if s.notifier != nil {
		s.notifier.OrderPlaced(ctx, order)
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

// NotificationSummary is the lightweight payload the SPA polls for its navbar
// badges.
type NotificationSummary struct {
	// ActiveOrders is the caller's in-flight order count (customer side).
	ActiveOrders int `json:"active_orders"`
	// PendingChefOrders is how many orders await the caller's accept/decline
	// (chef side; 0 for non-chefs or chefs without a profile).
	PendingChefOrders int `json:"pending_chef_orders"`
}

// Summary returns the notification counts for the caller. isChef comes from
// the token's role; a chef without a profile simply reports 0 pending.
func (s *OrderService) Summary(ctx context.Context, userID int, isChef bool) (*NotificationSummary, error) {
	out := &NotificationSummary{}
	active, err := s.orders.CountActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out.ActiveOrders = active

	if isChef {
		chef, err := s.chefs.FindByUserID(ctx, userID)
		if err == nil {
			pending, err := s.orders.CountPendingByChef(ctx, chef.ID)
			if err != nil {
				return nil, err
			}
			out.PendingChefOrders = pending
		} else if err != domain.ErrChefNotFound {
			return nil, err
		}
	}
	return out, nil
}

// ListForCustomer returns a page of the customer's order history and the total.
func (s *OrderService) ListForCustomer(ctx context.Context, userID, limit, offset int) ([]*domain.Order, int, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.orders.ListByUser(ctx, userID, limit, offset)
}

// CancelForCustomer cancels the customer's own order (pending/confirmed only).
// A paid card order is refunded through the gateway first — if the refund
// fails, the order stays uncancelled so money and state never diverge.
func (s *OrderService) CancelForCustomer(ctx context.Context, userID, orderID int) (*domain.Order, error) {
	order, err := s.GetForCustomer(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if !order.CanCancel() {
		return nil, domain.ErrInvalidStatusTransition
	}

	refund := order.IsCardPaid() && s.refunder != nil
	if refund {
		if err := s.refunder.RefundOrderPayment(ctx, order); err != nil {
			return nil, err
		}
	}
	if err := order.Cancel(); err != nil {
		return nil, err
	}
	if refund {
		_ = order.Refund() // paid -> refunded; guarded by IsCardPaid above
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

// AdvanceForChef applies a status transition to the caller's own sub-order —
// chef A advancing never touches chef B's slice. "confirm" accepts the
// sub-order; "decline" cancels it (refunding the chef's slice of a card-paid
// order first); the rest move it along the lifecycle. The order-level status
// is re-derived from the sub-orders and both are persisted atomically.
// Illegal transitions return ErrInvalidStatusTransition.
func (s *OrderService) AdvanceForChef(ctx context.Context, userID, orderID int, action string) (*domain.Order, error) {
	chef, err := s.chefs.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	order, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	sub := order.SubOrderFor(chef.ID)
	if sub == nil {
		return nil, domain.ErrForbidden
	}

	if err := applyChefAction(sub, action); err != nil {
		return nil, err
	}
	// Declining a slice of a card-paid order returns that chef's money —
	// food subtotal plus its delivery fee, since nothing will be delivered —
	// before anything persists. If the partial refund fails, the decline
	// aborts.
	if action == OrderActionDecline && order.IsCardPaid() && s.refunder != nil {
		if err := s.refunder.RefundSubOrderPayment(ctx, order, sub.Subtotal+sub.DeliveryFee); err != nil {
			return nil, err
		}
	}

	// A chef accepting stamps the ETA (once) — the customer gets a delivery
	// estimate as soon as the first kitchen commits.
	if action == OrderActionConfirm {
		order.SetEstimatedDelivery(s.etaWindow)
	}

	order.SyncStatusFromSubOrders()
	// Every chef declined a card-paid order: each slice was refunded above, so
	// the order's payment is fully returned.
	if order.Status == domain.OrderStatusCancelled && order.IsCardPaid() {
		_ = order.Refund()
	}
	// Cash settles at the door: once every sub-order is delivered the derived
	// order is delivered, and a cash order then counts as paid, so it flows
	// into chef earnings (delivered & paid).
	order.SettleCashOnDelivery()
	if err := s.orders.UpdateSubOrder(ctx, order, sub); err != nil {
		return nil, err
	}
	if s.notifier != nil {
		s.notifier.SubOrderAdvanced(ctx, order, sub)
	}
	return order, nil
}

func applyChefAction(sub *domain.SubOrder, action string) error {
	switch action {
	case OrderActionConfirm:
		return sub.Confirm()
	case OrderActionPreparing:
		return sub.StartPreparing()
	case OrderActionReady:
		return sub.MarkReady()
	case OrderActionDelivering:
		return sub.StartDelivering()
	case OrderActionDelivered:
		return sub.MarkDelivered()
	case OrderActionDecline:
		return sub.Cancel()
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
