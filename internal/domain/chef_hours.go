package domain

import "time"

// MinutesPerDay bounds the minutes-since-midnight fields of a working-hours
// window.
const MinutesPerDay = 24 * 60

// ChefHours is one weekly working-hours window (mirrors chef_hours,
// migrations/000020). Times are minutes since midnight in the platform time
// zone; OpensAt > ClosesAt means the window wraps past midnight (a dinner
// kitchen open 18:00–02:00). A chef with no windows at all is always open.
type ChefHours struct {
	ID     int `json:"id,omitempty"`
	ChefID int `json:"chef_id,omitempty"`

	Weekday  int `json:"weekday"` // 0 = Sunday, matching Go's time.Weekday
	OpensAt  int `json:"opens_at"`
	ClosesAt int `json:"closes_at"`
}

// Validate checks a single window's invariants.
func (h *ChefHours) Validate() error {
	if h.Weekday < 0 || h.Weekday > 6 {
		return ErrInvalidWeekday
	}
	if h.OpensAt < 0 || h.OpensAt >= MinutesPerDay || h.ClosesAt < 0 || h.ClosesAt >= MinutesPerDay {
		return ErrInvalidHoursWindow
	}
	if h.OpensAt == h.ClosesAt {
		return ErrInvalidHoursWindow
	}
	return nil
}

// IsOpenAt reports whether a chef with the given windows is open at t (which
// the caller has already moved into the platform time zone). No windows means
// always open. Opens is inclusive, closes exclusive — a 09:00–17:00 kitchen
// takes its last order at 16:59.
func IsOpenAt(hours []*ChefHours, t time.Time) bool {
	if len(hours) == 0 {
		return true
	}
	weekday := int(t.Weekday())
	minutes := t.Hour()*60 + t.Minute()
	yesterday := (weekday + 6) % 7

	for _, h := range hours {
		if h.OpensAt < h.ClosesAt {
			// Same-day window.
			if h.Weekday == weekday && minutes >= h.OpensAt && minutes < h.ClosesAt {
				return true
			}
			continue
		}
		// Overnight window: the evening part today, or the after-midnight
		// tail of yesterday's window.
		if h.Weekday == weekday && minutes >= h.OpensAt {
			return true
		}
		if h.Weekday == yesterday && minutes < h.ClosesAt {
			return true
		}
	}
	return false
}
