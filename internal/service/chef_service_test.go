package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeChefRepo is an in-memory domain.ChefRepository for tests.
type fakeChefRepo struct {
	chefs  map[int]*domain.Chef
	nextID int
}

func newFakeChefRepo() *fakeChefRepo {
	return &fakeChefRepo{chefs: map[int]*domain.Chef{}, nextID: 1}
}

func (f *fakeChefRepo) Create(_ context.Context, c *domain.Chef) error {
	c.ID = f.nextID
	f.nextID++
	cp := *c
	f.chefs[c.ID] = &cp
	return nil
}

func (f *fakeChefRepo) FindByID(_ context.Context, id int) (*domain.Chef, error) {
	if c, ok := f.chefs[id]; ok && c.IsActive {
		cp := *c
		return &cp, nil
	}
	return nil, domain.ErrChefNotFound
}

func (f *fakeChefRepo) FindByUserID(_ context.Context, userID int) (*domain.Chef, error) {
	for _, c := range f.chefs {
		if c.UserID == userID {
			cp := *c
			return &cp, nil
		}
	}
	return nil, domain.ErrChefNotFound
}

func (f *fakeChefRepo) List(_ context.Context, limit, offset int, onlineOnly bool) ([]*domain.Chef, int, error) {
	out := make([]*domain.Chef, 0)
	for _, c := range f.chefs {
		if c.IsActive && (!onlineOnly || c.IsOnline) {
			cp := *c
			out = append(out, &cp)
		}
	}
	total := len(out)
	if offset >= len(out) {
		return []*domain.Chef{}, total, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

func (f *fakeChefRepo) SetOnline(_ context.Context, chefID int, online bool) error {
	if c, ok := f.chefs[chefID]; ok {
		c.IsOnline = online
		return nil
	}
	return domain.ErrChefNotFound
}

func (f *fakeChefRepo) FindNearby(_ context.Context, lat, lng float64, limit int, onlineOnly bool) ([]*domain.Chef, error) {
	out := make([]*domain.Chef, 0)
	for _, c := range f.chefs {
		if c.IsActive && (!onlineOnly || c.IsOnline) && c.CanDeliverTo(lat, lng) {
			cp := *c
			out = append(out, &cp)
			if len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

func validProfile() service.CreateProfileInput {
	return service.CreateProfileInput{BusinessName: "Yasin's Kitchen", KitchenAddress: "123 Main St"}
}

func TestCreateProfile_Success(t *testing.T) {
	svc := service.NewChefService(newFakeChefRepo())

	chef, err := svc.CreateProfile(context.Background(), 42, validProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chef.ID == 0 || chef.UserID != 42 {
		t.Errorf("unexpected chef: %+v", chef)
	}
	if chef.DeliveryRadius != 5 {
		t.Errorf("default delivery radius = %d, want 5", chef.DeliveryRadius)
	}
}

func TestCreateProfile_OnePerUser(t *testing.T) {
	svc := service.NewChefService(newFakeChefRepo())
	ctx := context.Background()
	if _, err := svc.CreateProfile(ctx, 42, validProfile()); err != nil {
		t.Fatalf("first profile failed: %v", err)
	}

	_, err := svc.CreateProfile(ctx, 42, validProfile())
	if !errors.Is(err, domain.ErrChefProfileExists) {
		t.Errorf("err = %v, want ErrChefProfileExists", err)
	}
}

func TestCreateProfile_Validation(t *testing.T) {
	svc := service.NewChefService(newFakeChefRepo())
	lat := 41.0
	cases := map[string]func(*service.CreateProfileInput){
		"missing business name": func(in *service.CreateProfileInput) { in.BusinessName = "  " },
		"missing address":       func(in *service.CreateProfileInput) { in.KitchenAddress = "" },
		"negative radius":       func(in *service.CreateProfileInput) { in.DeliveryRadius = -1 },
		"half coordinates":      func(in *service.CreateProfileInput) { in.Latitude = &lat },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			in := validProfile()
			mutate(&in)
			_, err := svc.CreateProfile(context.Background(), 1, in)
			var ve service.ValidationError
			if !errors.As(err, &ve) {
				t.Errorf("err = %v, want ValidationError", err)
			}
		})
	}
}

func TestGetChef_NotFound(t *testing.T) {
	svc := service.NewChefService(newFakeChefRepo())
	if _, err := svc.Get(context.Background(), 999); !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("err = %v, want ErrChefNotFound", err)
	}
}

func TestNearby_FiltersByRadius(t *testing.T) {
	repo := newFakeChefRepo()
	svc := service.NewChefService(repo)
	ctx := context.Background()

	lat, lng := 41.0082, 28.9784
	near := validProfile()
	near.Latitude, near.Longitude, near.DeliveryRadius = &lat, &lng, 10
	if _, err := svc.CreateProfile(ctx, 1, near); err != nil {
		t.Fatal(err)
	}

	farLat, farLng := 39.9334, 32.8597 // Ankara, ~350km
	far := validProfile()
	far.Latitude, far.Longitude, far.DeliveryRadius = &farLat, &farLng, 10
	if _, err := svc.CreateProfile(ctx, 2, far); err != nil {
		t.Fatal(err)
	}

	got, err := svc.Nearby(ctx, lat, lng, 20, false)
	if err != nil {
		t.Fatalf("nearby failed: %v", err)
	}
	if len(got) != 1 || got[0].UserID != 1 {
		t.Errorf("nearby returned %d chefs, want only the Istanbul chef", len(got))
	}
}
