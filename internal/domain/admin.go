package domain

// PlatformStats is the admin dashboard summary, aggregated by SQL (no stored
// table). GMV counts only delivered & paid orders.
type PlatformStats struct {
	TotalUsers      int       `json:"total_users"`
	TotalChefs      int       `json:"total_chefs"`
	ActiveChefs     int       `json:"active_chefs"`
	TotalOrders     int       `json:"total_orders"`
	DeliveredOrders int       `json:"delivered_orders"`
	OrdersToday     int       `json:"orders_today"`
	GMV             float64   `json:"gmv"`
	TopChefs        []TopChef `json:"top_chefs"`
}

// TopChef is one row of the "best performing chefs" leaderboard.
type TopChef struct {
	ChefID       int     `json:"chef_id"`
	BusinessName string  `json:"business_name"`
	Orders       int     `json:"orders"`
	Revenue      float64 `json:"revenue"`
}
