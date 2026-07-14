package domain

import (
	"testing"
	"time"
)

// istanbul returns a time on the given weekday at hh:mm in Europe/Istanbul.
// 2026-07-12 is a Sunday, so weekday d lands on July 12+d.
func istanbul(t *testing.T, weekday, hh, mm int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		t.Fatalf("load tz: %v", err)
	}
	tm := time.Date(2026, 7, 12+weekday, hh, mm, 0, 0, loc)
	if int(tm.Weekday()) != weekday {
		t.Fatalf("fixture broken: %v is weekday %d, want %d", tm, tm.Weekday(), weekday)
	}
	return tm
}

func win(weekday, opens, closes int) *ChefHours {
	return &ChefHours{Weekday: weekday, OpensAt: opens, ClosesAt: closes}
}

func TestChefHours_Validate(t *testing.T) {
	cases := map[string]struct {
		h    *ChefHours
		want error
	}{
		"valid":         {win(1, 9*60, 17*60), nil},
		"overnight ok":  {win(5, 18*60, 2*60), nil},
		"weekday low":   {win(-1, 0, 60), ErrInvalidWeekday},
		"weekday high":  {win(7, 0, 60), ErrInvalidWeekday},
		"opens too big": {win(1, MinutesPerDay, 60), ErrInvalidHoursWindow},
		"negative":      {win(1, -1, 60), ErrInvalidHoursWindow},
		"empty window":  {win(1, 600, 600), ErrInvalidHoursWindow},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := tc.h.Validate(); err != tc.want {
				t.Errorf("Validate() = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestIsOpenAt(t *testing.T) {
	// Mon 09:00–17:00, plus Fri 18:00–02:00 (overnight into Saturday).
	hours := []*ChefHours{
		win(1, 9*60, 17*60),
		win(5, 18*60, 2*60),
	}

	cases := map[string]struct {
		at   time.Time
		want bool
	}{
		"monday during":            {istanbul(t, 1, 12, 0), true},
		"monday opens inclusive":   {istanbul(t, 1, 9, 0), true},
		"monday closes exclusive":  {istanbul(t, 1, 17, 0), false},
		"monday before":            {istanbul(t, 1, 8, 59), false},
		"tuesday closed all day":   {istanbul(t, 2, 12, 0), false},
		"friday evening open":      {istanbul(t, 5, 23, 30), true},
		"friday afternoon closed":  {istanbul(t, 5, 15, 0), false},
		"saturday after midnight":  {istanbul(t, 6, 1, 30), true},
		"saturday past the close":  {istanbul(t, 6, 2, 0), false},
		"sunday not fri spillover": {istanbul(t, 0, 1, 0), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsOpenAt(hours, tc.at); got != tc.want {
				t.Errorf("IsOpenAt = %v, want %v", got, tc.want)
			}
		})
	}

	// No windows at all = always open (backwards compatible).
	if !IsOpenAt(nil, istanbul(t, 2, 3, 0)) {
		t.Error("a chef without hours must be always open")
	}
	// Two windows on one day (lunch + dinner).
	split := []*ChefHours{win(3, 11*60, 14*60), win(3, 18*60, 22*60)}
	if !IsOpenAt(split, istanbul(t, 3, 12, 0)) || IsOpenAt(split, istanbul(t, 3, 16, 0)) || !IsOpenAt(split, istanbul(t, 3, 19, 0)) {
		t.Error("split shift windows evaluated wrongly")
	}
}
