package domain

import "time"

// Order is a customer order (mirrors the orders table,
// migrations/000002_create_orders_table.up.sql). A single order may contain
// items from several chefs — each line in Items carries its own chef_id — so
// the order belongs to the customer (UserID) while fulfilment is split per chef
// via OrderItem.ChefID.
//
// Status is a state machine: always move through the transition methods
// (Confirm, StartPreparing, …) which reject illegal moves with
// ErrInvalidStatusTransition. Never assign Status directly.
type Order struct {
	ID        int    `json:"id"`
	OrderCode string `json:"order_code"`
	UserID    int    `json:"user_id"`

	// Pricing
	Subtotal    float64 `json:"subtotal"`
	DeliveryFee float64 `json:"delivery_fee"`
	ServiceFee  float64 `json:"service_fee"`
	Tax         float64 `json:"tax"`
	Discount    float64 `json:"discount"`
	// Tip is the customer's optional gratuity (#105), added to the total and
	// paid to the chef(s) uncommissioned; split per slice on the sub-orders.
	Tip        float64 `json:"tip"`
	TotalPrice float64 `json:"total_price"`

	// Status and payment
	Status        string  `json:"status"`
	PaymentMethod *string `json:"payment_method,omitempty"`
	PaymentStatus string  `json:"payment_status"`

	// PromoCode is the applied discount code, snapshotted at placement (#94).
	PromoCode *string `json:"promo_code,omitempty"`

	// Delivery
	DeliveryAddress       string     `json:"delivery_address"`
	DeliveryCity          *string    `json:"delivery_city,omitempty"`
	DeliveryLatitude      *float64   `json:"delivery_latitude,omitempty"`
	DeliveryLongitude     *float64   `json:"delivery_longitude,omitempty"`
	EstimatedDeliveryTime *time.Time `json:"estimated_delivery_time,omitempty"`
	ActualDeliveryTime    *time.Time `json:"actual_delivery_time,omitempty"`

	// Notes
	CustomerNotes *string `json:"customer_notes,omitempty"`
	ChefNotes     *string `json:"chef_notes,omitempty"`
	DeliveryNotes *string `json:"delivery_notes,omitempty"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	// Items is loaded alongside the order; it is not a column.
	Items []*OrderItem `json:"items,omitempty"`

	// SubOrders is the per-chef slices of the order (one per participating
	// chef), loaded alongside it. Status is derived from these — see
	// SyncStatusFromSubOrders.
	SubOrders []*SubOrder `json:"sub_orders,omitempty"`
}

// Order status values. See the lifecycle in CLAUDE.md §4.
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusPreparing  = "preparing"
	OrderStatusReady      = "ready"
	OrderStatusDelivering = "delivering"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

// Payment method values.
const (
	PaymentMethodCash = "cash"
	PaymentMethodCard = "card"
)

// Payment status values.
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// ValidPaymentMethod reports whether m is a recognised payment method.
func ValidPaymentMethod(m string) bool {
	return m == PaymentMethodCash || m == PaymentMethodCard
}

// NewOrder builds a pending, unpaid order for a customer. The caller fills
// pricing, items and OrderCode before persisting.
func NewOrder(userID int, deliveryAddress string) *Order {
	now := time.Now()
	return &Order{
		UserID:          userID,
		DeliveryAddress: deliveryAddress,
		Status:          OrderStatusPending,
		PaymentStatus:   PaymentStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Confirm moves a pending order to confirmed (a chef accepting the order).
func (o *Order) Confirm() error { return o.transition(OrderStatusPending, OrderStatusConfirmed) }

// StartPreparing moves a confirmed order to preparing.
func (o *Order) StartPreparing() error {
	return o.transition(OrderStatusConfirmed, OrderStatusPreparing)
}

// MarkReady moves a preparing order to ready.
func (o *Order) MarkReady() error { return o.transition(OrderStatusPreparing, OrderStatusReady) }

// StartDelivering moves a ready order to delivering.
func (o *Order) StartDelivering() error {
	return o.transition(OrderStatusReady, OrderStatusDelivering)
}

// MarkDelivered moves a delivering order to delivered and stamps the delivery
// time.
func (o *Order) MarkDelivered() error {
	if err := o.transition(OrderStatusDelivering, OrderStatusDelivered); err != nil {
		return err
	}
	now := time.Now()
	o.ActualDeliveryTime = &now
	return nil
}

// CanCancel reports whether the order may still be cancelled: only pending or
// confirmed orders (before preparation starts). When sub-orders are loaded,
// every one of them must still be cancellable — one chef having started
// preparing locks the whole order in.
func (o *Order) CanCancel() bool {
	if o.Status != OrderStatusPending && o.Status != OrderStatusConfirmed {
		return false
	}
	for _, s := range o.SubOrders {
		if s.Status != OrderStatusCancelled && !s.CanCancel() {
			return false
		}
	}
	return true
}

// IsCardPaid reports whether the order was paid by card — the case that needs
// a gateway refund when the order is cancelled.
func (o *Order) IsCardPaid() bool {
	return o.PaymentStatus == PaymentStatusPaid &&
		o.PaymentMethod != nil && *o.PaymentMethod == PaymentMethodCard
}

// Cancel cancels an order — and every still-active sub-order with it (a
// customer changing their mind before any preparation starts). Only pending or
// confirmed orders may be cancelled.
func (o *Order) Cancel() error {
	if !o.CanCancel() {
		return ErrInvalidStatusTransition
	}
	for _, s := range o.SubOrders {
		if s.Status != OrderStatusCancelled {
			if err := s.Cancel(); err != nil {
				return err
			}
		}
	}
	now := time.Now()
	o.Status = OrderStatusCancelled
	o.CancelledAt = &now
	o.UpdatedAt = now
	return nil
}

// SetEstimatedDelivery stamps an ETA of now+window (once): the first chef to
// accept sets it; later transitions leave it. A zero/negative window is a
// no-op. Kept idempotent so a multi-chef order's ETA doesn't drift each time
// another chef confirms.
func (o *Order) SetEstimatedDelivery(window time.Duration) {
	if window <= 0 || o.EstimatedDeliveryTime != nil {
		return
	}
	eta := time.Now().Add(window)
	o.EstimatedDeliveryTime = &eta
	o.UpdatedAt = time.Now()
}

// SubOrderFor returns the chef's sub-order, or nil when the chef has no slice
// of this order. It is the basis for chef-scoped status transitions.
func (o *Order) SubOrderFor(chefID int) *SubOrder {
	for _, s := range o.SubOrders {
		if s.ChefID == chefID {
			return s
		}
	}
	return nil
}

// SyncStatusFromSubOrders re-derives the order-level status after a sub-order
// transition, stamping delivery/cancellation times when the derived status
// lands there. Call it after mutating any sub-order.
func (o *Order) SyncStatusFromSubOrders() {
	derived := DeriveOrderStatus(o.SubOrders)
	if derived == o.Status {
		return
	}
	now := time.Now()
	o.Status = derived
	o.UpdatedAt = now
	switch derived {
	case OrderStatusDelivered:
		o.ActualDeliveryTime = &now
	case OrderStatusCancelled:
		o.CancelledAt = &now
	}
}

// SettleCashOnDelivery marks payment as paid for a delivered cash order —
// cash changes hands at the door, so delivery settles the payment. It is a
// no-op for card orders (a gateway drives those), undelivered orders, and
// orders whose payment is already paid/failed/refunded.
func (o *Order) SettleCashOnDelivery() {
	if o.Status != OrderStatusDelivered {
		return
	}
	if o.PaymentMethod == nil || *o.PaymentMethod != PaymentMethodCash {
		return
	}
	if o.PaymentStatus != PaymentStatusPending {
		return
	}
	_ = o.MarkPaid()
}

// MarkPaid records a successful payment (pending → paid).
func (o *Order) MarkPaid() error {
	if o.PaymentStatus != PaymentStatusPending {
		return ErrInvalidPaymentTransition
	}
	o.PaymentStatus = PaymentStatusPaid
	o.UpdatedAt = time.Now()
	return nil
}

// Refund records a refund (paid → refunded).
func (o *Order) Refund() error {
	if o.PaymentStatus != PaymentStatusPaid {
		return ErrInvalidPaymentTransition
	}
	o.PaymentStatus = PaymentStatusRefunded
	o.UpdatedAt = time.Now()
	return nil
}

// HasChef reports whether any line in the order belongs to chefID. It is the
// basis for chef-scoped authorization on an order.
func (o *Order) HasChef(chefID int) bool {
	for _, it := range o.Items {
		if it.ChefID == chefID {
			return true
		}
	}
	return false
}

// HasMenuItem reports whether the order contains the given dish. It backs the
// "you can only review what you ordered" rule for product reviews.
func (o *Order) HasMenuItem(menuItemID int) bool {
	for _, it := range o.Items {
		if it.MenuItemID == menuItemID {
			return true
		}
	}
	return false
}

// transition enforces a single legal status move.
func (o *Order) transition(from, to string) error {
	if o.Status != from {
		return ErrInvalidStatusTransition
	}
	o.Status = to
	o.UpdatedAt = time.Now()
	return nil
}
