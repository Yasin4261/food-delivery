package domain

import "time"

// Order represents a customer order
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
	TotalPrice  float64 `json:"total_price"`
	
	// Status and payment
	Status        string  `json:"status"`
	PaymentMethod *string `json:"payment_method,omitempty"`
	PaymentStatus string  `json:"payment_status"`
	
	// Delivery information
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
	
	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	
	// Relations (not in DB, loaded separately)
	Items []OrderItem `json:"items,omitempty"`
}

// OrderStatus constants
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusPreparing  = "preparing"
	OrderStatusReady      = "ready"
	OrderStatusDelivering = "delivering"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

// PaymentStatus constants
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// NewOrder creates a new order
func NewOrder(userID int, deliveryAddress string, subtotal float64) *Order {
	return &Order{
		UserID:          userID,
		DeliveryAddress: deliveryAddress,
		Subtotal:        subtotal,
		Status:          OrderStatusPending,
		PaymentStatus:   PaymentStatusPending,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// CalculateTotal calculates the total price
func (o *Order) CalculateTotal() {
	o.TotalPrice = o.Subtotal + o.DeliveryFee + o.ServiceFee + o.Tax - o.Discount
	o.UpdatedAt = time.Now()
}

// Confirm confirms the order
func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return ErrInvalidStatusTransition
	}
	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
	return nil
}

// StartPreparing starts preparing the order
func (o *Order) StartPreparing() error {
	if o.Status != OrderStatusConfirmed {
		return ErrInvalidStatusTransition
	}
	o.Status = OrderStatusPreparing
	o.UpdatedAt = time.Now()
	return nil
}

// MarkReady marks order as ready for delivery
func (o *Order) MarkReady() error {
	if o.Status != OrderStatusPreparing {
		return ErrInvalidStatusTransition
	}
	o.Status = OrderStatusReady
	o.UpdatedAt = time.Now()
	return nil
}

// StartDelivering starts delivering the order
func (o *Order) StartDelivering() error {
	if o.Status != OrderStatusReady {
		return ErrInvalidStatusTransition
	}
	o.Status = OrderStatusDelivering
	o.UpdatedAt = time.Now()
	return nil
}

// MarkDelivered marks order as delivered
func (o *Order) MarkDelivered() error {
	if o.Status != OrderStatusDelivering {
		return ErrInvalidStatusTransition
	}
	o.Status = OrderStatusDelivered
	now := time.Now()
	o.ActualDeliveryTime = &now
	o.UpdatedAt = now
	return nil
}

// Cancel cancels the order
func (o *Order) Cancel() error {
	if o.Status == OrderStatusDelivered || o.Status == OrderStatusCancelled {
		return ErrCannotCancelOrder
	}
	o.Status = OrderStatusCancelled
	now := time.Now()
	o.CancelledAt = &now
	o.UpdatedAt = now
	return nil
}

// IsCancellable checks if order can be cancelled
func (o *Order) IsCancellable() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

// OrderItem represents an item in an order (snapshot)
type OrderItem struct {
	ID                  int       `json:"id"`
	OrderID             int       `json:"order_id"`
	MenuItemID          int       `json:"menu_item_id"`
	ChefID              int       `json:"chef_id"`
	ItemName            string    `json:"item_name"`
	Quantity            int       `json:"quantity"`
	UnitPrice           float64   `json:"unit_price"`
	Subtotal            float64   `json:"subtotal"`
	SpecialInstructions *string   `json:"special_instructions,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// NewOrderItem creates a new order item
func NewOrderItem(orderID, menuItemID, chefID int, itemName string, quantity int, unitPrice float64) *OrderItem {
	return &OrderItem{
		OrderID:    orderID,
		MenuItemID: menuItemID,
		ChefID:     chefID,
		ItemName:   itemName,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		Subtotal:   float64(quantity) * unitPrice,
		CreatedAt:  time.Now(),
	}
}
