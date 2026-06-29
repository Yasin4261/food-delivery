package service_test

import (
	"context"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

type fakeSearchRepo struct{}

func (fakeSearchRepo) SearchChefs(_ context.Context, q string, limit, offset int) ([]*domain.Chef, error) {
	return []*domain.Chef{{ID: 1}}, nil
}
func (fakeSearchRepo) SearchMenuItems(_ context.Context, q string, limit, offset int) ([]*domain.MenuItem, error) {
	return []*domain.MenuItem{{ID: 1}}, nil
}
func (fakeSearchRepo) SearchUsers(_ context.Context, q string, limit, offset int) ([]*domain.User, error) {
	return []*domain.User{{ID: 1, PasswordHash: "secret"}}, nil
}

func TestSearchService_EmptyQueryRejected(t *testing.T) {
	svc := service.NewSearchService(fakeSearchRepo{})
	if _, err := svc.Chefs(context.Background(), "   ", 20, 0); !isValidation(err) {
		t.Errorf("blank chef query = %v, want ValidationError", err)
	}
	if _, err := svc.Foods(context.Background(), "", 20, 0); !isValidation(err) {
		t.Errorf("blank food query = %v, want ValidationError", err)
	}
}

func TestSearchService_UsersClearsPasswordHash(t *testing.T) {
	svc := service.NewSearchService(fakeSearchRepo{})
	users, err := svc.Users(context.Background(), "yasin", 20, 0)
	if err != nil {
		t.Fatalf("users: %v", err)
	}
	if len(users) != 1 || users[0].PasswordHash != "" {
		t.Errorf("password hash must be cleared, got %+v", users[0])
	}
}

func TestSearchService_DelegatesChefAndFood(t *testing.T) {
	svc := service.NewSearchService(fakeSearchRepo{})
	if chefs, _ := svc.Chefs(context.Background(), "x", 20, 0); len(chefs) != 1 {
		t.Errorf("chef search len = %d, want 1", len(chefs))
	}
	if foods, _ := svc.Foods(context.Background(), "x", 20, 0); len(foods) != 1 {
		t.Errorf("food search len = %d, want 1", len(foods))
	}
}
