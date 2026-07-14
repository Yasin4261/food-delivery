package domain

import "context"

// ChefHoursRepository is the port for working-hours persistence.
type ChefHoursRepository interface {
	// ReplaceForChef atomically replaces the chef's whole weekly schedule
	// (the editor submits the full week; an empty slice clears it, meaning
	// "always open").
	ReplaceForChef(ctx context.Context, chefID int, hours []*ChefHours) error
	// ListByChef returns a chef's windows ordered by weekday, opens_at.
	ListByChef(ctx context.Context, chefID int) ([]*ChefHours, error)
	// ListByChefs returns the windows of several chefs in one query, grouped
	// by chef id (chefs without windows are absent from the map).
	ListByChefs(ctx context.Context, chefIDs []int) (map[int][]*ChefHours, error)
}
