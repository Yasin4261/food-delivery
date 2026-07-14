package domain

import "math"

// FeePolicy is the platform's money model (#65), loaded from configuration:
//
//   - Delivery fee, charged to the customer per chef slice: a base amount
//     plus a per-kilometre component when both the kitchen and the delivery
//     address have coordinates (Haversine). Without coordinates only the
//     base applies.
//   - Commission, deducted from the chef: a percentage of the slice's food
//     subtotal. The chef keeps the delivery fee in full — they deliver.
//
// Amounts computed here are snapshotted onto sub_orders at placement, so a
// later rate change never alters historical orders.
type FeePolicy struct {
	DeliveryBaseFee  float64 // per chef slice
	DeliveryFeePerKm float64 // added per km of kitchen -> customer distance
	CommissionRate   float64 // percent of the food subtotal, 0–100
}

// RoundMoney rounds a monetary amount to two decimals — float sums pick up
// representation dust that must never reach totals or gateways.
func RoundMoney(v float64) float64 { return math.Round(v*100) / 100 }

// DeliveryFee returns the customer-facing delivery fee for one chef slice.
// distanceKm < 0 means "distance unknown" (missing coordinates): only the
// base fee applies.
func (p FeePolicy) DeliveryFee(distanceKm float64) float64 {
	fee := p.DeliveryBaseFee
	if distanceKm > 0 {
		fee += p.DeliveryFeePerKm * distanceKm
	}
	return RoundMoney(fee)
}

// Commission returns the platform's cut of a slice's food subtotal.
func (p FeePolicy) Commission(subtotal float64) float64 {
	return RoundMoney(subtotal * p.CommissionRate / 100)
}
