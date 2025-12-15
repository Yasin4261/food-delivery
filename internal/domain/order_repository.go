package domain

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	// Create creates a new order with order items
	Create(order *Order, items []OrderItem) error
	
	// FindByID finds an order by ID
	FindByID(id int) (*Order, error)
	
	// FindByOrderCode finds an order by order code
	FindByOrderCode(orderCode string) (*Order, error)
	
	// Update updates order information
	Update(order *Order) error
	
	// UpdateStatus updates order status
	UpdateStatus(id int, status string) error
	
	// FindByUserID finds orders by user ID
	FindByUserID(userID int, offset, limit int) ([]*Order, error)
	
	// FindByChefID finds orders containing items from a chef
	FindByChefID(chefID int, offset, limit int) ([]*Order, error)
	
	// FindByStatus finds orders by status
	FindByStatus(status string, offset, limit int) ([]*Order, error)
	
	// FindActiveOrders finds all non-completed/non-cancelled orders
	FindActiveOrders(offset, limit int) ([]*Order, error)
	
	// FindOrderItems finds items for an order
	FindOrderItems(orderID int) ([]OrderItem, error)
	
	// CountByStatus counts orders by status
	CountByStatus(status string) (int, error)
	
	// GetTotalRevenue calculates total revenue for a chef
	GetTotalRevenue(chefID int) (float64, error)
}
