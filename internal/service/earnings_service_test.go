package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

type fakeEarningsRepo struct {
	gotChefID int
	gotSince  *time.Time
	result    *domain.Earnings
}

func (f *fakeEarningsRepo) ChefEarnings(_ context.Context, chefID int, since *time.Time) (*domain.Earnings, error) {
	f.gotChefID = chefID
	f.gotSince = since
	if f.result != nil {
		return f.result, nil
	}
	return &domain.Earnings{ChefID: chefID}, nil
}

func TestEarningsService_ForChef_ResolvesChefAndWindow(t *testing.T) {
	chefRepo := newFakeChefRepo()
	if err := chefRepo.Create(context.Background(), &domain.Chef{UserID: 50, IsActive: true}); err != nil {
		t.Fatalf("seed chef: %v", err)
	}
	earn := &fakeEarningsRepo{result: &domain.Earnings{TotalEarnings: 99}}
	svc := service.NewEarningsService(earn, chefRepo)

	// All-time: no window.
	got, err := svc.ForChef(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("for chef: %v", err)
	}
	if got.TotalEarnings != 99 {
		t.Errorf("total = %v, want 99", got.TotalEarnings)
	}
	if earn.gotChefID != 1 { // first seeded chef gets id 1
		t.Errorf("chef id passed = %d, want 1", earn.gotChefID)
	}
	if earn.gotSince != nil {
		t.Errorf("all-time should pass nil since, got %v", earn.gotSince)
	}

	// Windowed: since ~ now-7d.
	if _, err := svc.ForChef(context.Background(), 50, 7); err != nil {
		t.Fatalf("windowed: %v", err)
	}
	if earn.gotSince == nil || time.Since(*earn.gotSince) < 6*24*time.Hour {
		t.Errorf("since not set to ~7 days ago: %v", earn.gotSince)
	}
}

func TestEarningsService_ForChef_NoProfile(t *testing.T) {
	svc := service.NewEarningsService(&fakeEarningsRepo{}, newFakeChefRepo())
	if _, err := svc.ForChef(context.Background(), 999, 0); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("err = %v, want ErrChefNotFound", err)
	}
}
