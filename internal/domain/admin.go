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

// AdminDetailLimit caps the nested lists on a detail view, so one support
// lookup can never pull thousands of rows.
const AdminDetailLimit = 20

// AdminUserDetail is the support console's view of one account: who they are,
// their kitchen if they run one, and what they have done recently.
type AdminUserDetail struct {
	User *User `json:"user"`
	// Chef is the kitchen this user owns, when they are a chef.
	Chef    *Chef     `json:"chef,omitempty"`
	Orders  []*Order  `json:"orders"`
	Reviews []*Review `json:"reviews"`
}

// AdminOrderDetail is the support console's view of one order — the artefact
// most "where is my food" tickets are about. Payments carry only status and
// timestamps: the gateway token and payment id are `json:"-"` on
// PaymentSession and must stay hidden.
type AdminOrderDetail struct {
	Order    *Order            `json:"order"`
	Customer *User             `json:"customer,omitempty"`
	Payments []*PaymentSession `json:"payments"`
}

// AdminChefDetail is the support console's view of one kitchen.
type AdminChefDetail struct {
	Chef *Chef `json:"chef"`
	// Owner is the user account behind the kitchen.
	Owner  *User       `json:"owner,omitempty"`
	Items  []*MenuItem `json:"items"`
	Orders []*Order    `json:"orders"`
}

// AdminUserFilters narrows the admin user listing. The zero value matches
// every user (the previous unfiltered behaviour).
type AdminUserFilters struct {
	// Query matches email or username (case-insensitive, substring).
	Query string
	// Role restricts to one role ("" = any). Validated by the service.
	Role string
	// Active restricts by activation state; nil matches both.
	Active *bool
}

// AdminChefFilters narrows the admin chef listing. The zero value matches
// every chef.
type AdminChefFilters struct {
	// Query matches the kitchen's business name (case-insensitive, substring).
	Query string
	// Active restricts by activation state; nil matches both.
	Active *bool
}

// AdminOrderFilters narrows the admin order listing — the support view of
// "what happened to this order / this customer". The zero value matches every
// order.
type AdminOrderFilters struct {
	// Status / PaymentStatus restrict to one lifecycle value ("" = any).
	// Validated by the service against the domain's known values.
	Status        string
	PaymentStatus string
	// UserID scopes to one customer's orders (0 = any).
	UserID int
}
