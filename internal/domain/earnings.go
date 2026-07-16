package domain

// Earnings is a chef's derived earnings summary. It is not a stored table —
// it is aggregated from order_items and sub_orders, counting only slices the
// chef delivered on paid orders. Money model (#65): the chef keeps the food
// subtotal and the delivery fee; the platform's commission (snapshotted per
// slice) is deducted.
type Earnings struct {
	ChefID          int     `json:"chef_id"`
	TotalEarnings   float64 `json:"total_earnings"` // food subtotal (gross, pre-commission)
	DeliveryFees    float64 `json:"delivery_fees"`  // kept by the chef in full
	Tips            float64 `json:"tips"`           // customer gratuity, kept in full (#105)
	Commission      float64 `json:"commission"`     // platform's cut
	NetEarnings     float64 `json:"net_earnings"`   // subtotal + delivery fees + tips - commission
	DeliveredOrders int     `json:"delivered_orders"`
	ItemsSold       int     `json:"items_sold"`
}
