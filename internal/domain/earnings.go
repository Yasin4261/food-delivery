package domain

// Earnings is a chef's derived earnings summary. It is not a stored table —
// it is aggregated from order_items joined to their parent orders, counting
// only delivered and paid orders.
type Earnings struct {
	ChefID          int     `json:"chef_id"`
	TotalEarnings   float64 `json:"total_earnings"`
	DeliveredOrders int     `json:"delivered_orders"`
	ItemsSold       int     `json:"items_sold"`
}
