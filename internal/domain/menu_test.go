package domain_test

import (
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

func TestNewMenu_Defaults(t *testing.T) {
	m := domain.NewMenu(7, "Dinner")
	if m.ChefID != 7 || m.Name != "Dinner" {
		t.Errorf("unexpected menu: %+v", m)
	}
	if m.MenuType != domain.MenuTypeRegular {
		t.Errorf("menu type = %q, want regular", m.MenuType)
	}
	if !m.IsActive {
		t.Error("new menu should be active")
	}
}

func TestValidMenuType(t *testing.T) {
	for _, ok := range []string{"regular", "daily_special", "seasonal", "weekend"} {
		if !domain.ValidMenuType(ok) {
			t.Errorf("%q should be valid", ok)
		}
	}
	if domain.ValidMenuType("brunch") {
		t.Error("brunch should be invalid")
	}
}

func TestMenuItem_InStock(t *testing.T) {
	qty := 3
	cases := []struct {
		name string
		item domain.MenuItem
		want bool
	}{
		{"unlimited", domain.MenuItem{IsUnlimited: true}, true},
		{"enough", domain.MenuItem{AvailableQuantity: &qty}, true},
		{"untracked limited", domain.MenuItem{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.item.InStock(2); got != tc.want {
				t.Errorf("InStock(2) = %v, want %v", got, tc.want)
			}
		})
	}
	// Non-positive quantities are never in stock.
	unlimited := domain.MenuItem{IsUnlimited: true}
	if unlimited.InStock(0) {
		t.Error("InStock(0) should be false")
	}
}

func TestMenuItem_IsOrderable(t *testing.T) {
	m := domain.NewMenuItem(1, 1, "Soup", 5)
	m.IsUnlimited = true
	if !m.IsOrderable() {
		t.Error("active, available, unlimited item should be orderable")
	}
	m.IsAvailable = false
	if m.IsOrderable() {
		t.Error("unavailable item should not be orderable")
	}
}
