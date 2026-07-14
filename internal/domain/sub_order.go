package domain

import "time"

// SubOrder is the per-chef slice of a (possibly multi-chef) order — one row per
// (order, chef) in sub_orders (migrations/000015). It runs the same status
// lifecycle as an order (§4), but scoped to a single chef: chef A advancing
// their sub-order never moves chef B's. The parent Order.Status is derived
// from its sub-orders via DeriveOrderStatus.
//
// A sub-order owns no items of its own: its lines are the parent's Items
// filtered by ChefID.
type SubOrder struct {
	ID      int `json:"id"`
	OrderID int `json:"order_id"`
	ChefID  int `json:"chef_id"`

	Status   string  `json:"status"`
	Subtotal float64 `json:"subtotal"`

	// DeliveryFee (customer pays, chef keeps) and Commission (platform's cut
	// of the subtotal) are snapshots taken at placement from the FeePolicy of
	// the day — rate changes never rewrite history.
	DeliveryFee float64 `json:"delivery_fee"`
	Commission  float64 `json:"commission"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// ChefName is the chef's business name, loaded for display; not a column.
	ChefName string `json:"chef_name,omitempty"`
}

// NewSubOrder builds a pending sub-order for a chef's slice of an order.
func NewSubOrder(chefID int, subtotal float64) *SubOrder {
	now := time.Now()
	return &SubOrder{
		ChefID:    chefID,
		Status:    OrderStatusPending,
		Subtotal:  subtotal,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Confirm moves a pending sub-order to confirmed (the chef accepting it).
func (s *SubOrder) Confirm() error { return s.transition(OrderStatusPending, OrderStatusConfirmed) }

// StartPreparing moves a confirmed sub-order to preparing.
func (s *SubOrder) StartPreparing() error {
	return s.transition(OrderStatusConfirmed, OrderStatusPreparing)
}

// MarkReady moves a preparing sub-order to ready.
func (s *SubOrder) MarkReady() error { return s.transition(OrderStatusPreparing, OrderStatusReady) }

// StartDelivering moves a ready sub-order to delivering.
func (s *SubOrder) StartDelivering() error {
	return s.transition(OrderStatusReady, OrderStatusDelivering)
}

// MarkDelivered moves a delivering sub-order to delivered.
func (s *SubOrder) MarkDelivered() error {
	return s.transition(OrderStatusDelivering, OrderStatusDelivered)
}

// CanCancel reports whether the sub-order may still be cancelled: only pending
// or confirmed (before preparation starts), mirroring Order.CanCancel.
func (s *SubOrder) CanCancel() bool {
	return s.Status == OrderStatusPending || s.Status == OrderStatusConfirmed
}

// Cancel cancels the sub-order (the chef declining, or the customer cancelling
// the whole order).
func (s *SubOrder) Cancel() error {
	if !s.CanCancel() {
		return ErrInvalidStatusTransition
	}
	s.Status = OrderStatusCancelled
	s.UpdatedAt = time.Now()
	return nil
}

func (s *SubOrder) transition(from, to string) error {
	if s.Status != from {
		return ErrInvalidStatusTransition
	}
	s.Status = to
	s.UpdatedAt = time.Now()
	return nil
}

// statusRank orders the lifecycle for deriving the parent status: the parent
// sits at the least-advanced active sub-order. Cancelled is excluded before
// ranking.
var statusRank = map[string]int{
	OrderStatusPending:    0,
	OrderStatusConfirmed:  1,
	OrderStatusPreparing:  2,
	OrderStatusReady:      3,
	OrderStatusDelivering: 4,
	OrderStatusDelivered:  5,
}

// DeriveOrderStatus computes the parent order's status from its sub-orders:
//   - every sub-order cancelled → cancelled
//   - every active (non-cancelled) sub-order delivered → delivered
//   - otherwise → the least-advanced active sub-order's status
//
// An empty slice returns pending (a just-created order before its sub-orders
// are attached).
func DeriveOrderStatus(subs []*SubOrder) string {
	if len(subs) == 0 {
		return OrderStatusPending
	}
	derived := ""
	for _, s := range subs {
		if s.Status == OrderStatusCancelled {
			continue
		}
		if derived == "" || statusRank[s.Status] < statusRank[derived] {
			derived = s.Status
		}
	}
	if derived == "" {
		return OrderStatusCancelled
	}
	return derived
}
