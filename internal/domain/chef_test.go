package domain

import (
	"math"
	"testing"
)

func TestNewChefDefaults(t *testing.T) {
	c := NewChef(7, "Yasin's Kitchen", "123 Main St")

	if c.UserID != 7 || c.BusinessName != "Yasin's Kitchen" {
		t.Errorf("unexpected fields: %+v", c)
	}
	if c.DeliveryRadius != defaultDeliveryRadiusKm {
		t.Errorf("delivery radius = %d, want %d", c.DeliveryRadius, defaultDeliveryRadiusKm)
	}
	if !c.IsActive || !c.IsAcceptingOrders {
		t.Error("new chef should be active and accepting orders")
	}
}

func TestChefCanDeliverTo(t *testing.T) {
	lat, lng := 41.0082, 28.9784 // Istanbul
	c := NewChef(1, "K", "addr")
	c.KitchenLatitude = &lat
	c.KitchenLongitude = &lng
	c.DeliveryRadius = 5

	// ~1 km away -> deliverable.
	if !c.CanDeliverTo(41.0150, 28.9784) {
		t.Error("a point ~1km away should be deliverable within a 5km radius")
	}
	// ~50 km away -> not deliverable.
	if c.CanDeliverTo(41.4500, 28.9784) {
		t.Error("a far point should not be deliverable")
	}
}

func TestChefCanDeliverTo_NoLocation(t *testing.T) {
	c := NewChef(1, "K", "addr") // no coordinates
	if c.CanDeliverTo(41.0, 29.0) {
		t.Error("a chef without coordinates can deliver nowhere")
	}
}

func TestCalculateDistance(t *testing.T) {
	// Istanbul -> Ankara is ~350 km.
	d := CalculateDistance(41.0082, 28.9784, 39.9334, 32.8597)
	if math.Abs(d-350) > 40 {
		t.Errorf("distance = %.1f km, want ~350 km", d)
	}
	// Same point -> 0.
	if d0 := CalculateDistance(41, 29, 41, 29); d0 != 0 {
		t.Errorf("distance to self = %.4f, want 0", d0)
	}
}
