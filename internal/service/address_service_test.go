package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// fakeAddressRepo is an in-memory domain.AddressRepository mirroring the
// Postgres adapter's default semantics (marking one default clears the rest).
type fakeAddressRepo struct {
	addresses map[int]*domain.Address
	nextID    int
}

func newFakeAddressRepo() *fakeAddressRepo {
	return &fakeAddressRepo{addresses: map[int]*domain.Address{}, nextID: 1}
}

func (f *fakeAddressRepo) clearDefault(userID int) {
	for _, a := range f.addresses {
		if a.UserID == userID {
			a.IsDefault = false
		}
	}
}

func (f *fakeAddressRepo) Create(_ context.Context, a *domain.Address) error {
	if a.IsDefault {
		f.clearDefault(a.UserID)
	}
	a.ID = f.nextID
	f.nextID++
	cp := *a
	f.addresses[a.ID] = &cp
	return nil
}

func (f *fakeAddressRepo) FindByID(_ context.Context, id int) (*domain.Address, error) {
	if a, ok := f.addresses[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, domain.ErrAddressNotFound
}

func (f *fakeAddressRepo) ListByUser(_ context.Context, userID int) ([]*domain.Address, error) {
	out := make([]*domain.Address, 0)
	// Default first (the SQL adapter orders by is_default DESC).
	for _, a := range f.addresses {
		if a.UserID == userID && a.IsDefault {
			cp := *a
			out = append(out, &cp)
		}
	}
	for _, a := range f.addresses {
		if a.UserID == userID && !a.IsDefault {
			cp := *a
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeAddressRepo) Update(_ context.Context, a *domain.Address) error {
	if _, ok := f.addresses[a.ID]; !ok {
		return domain.ErrAddressNotFound
	}
	if a.IsDefault {
		f.clearDefault(a.UserID)
	}
	cp := *a
	f.addresses[a.ID] = &cp
	return nil
}

func (f *fakeAddressRepo) Delete(_ context.Context, id int) error {
	if _, ok := f.addresses[id]; !ok {
		return domain.ErrAddressNotFound
	}
	delete(f.addresses, id)
	return nil
}

func validAddress() service.AddressInput {
	return service.AddressInput{Label: "Home", Address: "1 Main St", City: "Istanbul"}
}

func TestAddressService_CreateAndDefaults(t *testing.T) {
	repo := newFakeAddressRepo()
	svc := service.NewAddressService(repo)
	ctx := context.Background()

	// The first address becomes the default even when not asked.
	first, err := svc.Create(ctx, 1, validAddress())
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !first.IsDefault {
		t.Error("first address must become the default")
	}

	// A later default takes over; the old one is cleared.
	in := validAddress()
	in.Label = "Work"
	in.IsDefault = true
	second, err := svc.Create(ctx, 1, in)
	if err != nil {
		t.Fatalf("create second: %v", err)
	}
	if !second.IsDefault {
		t.Error("second address should be the default now")
	}
	list, _ := svc.List(ctx, 1)
	defaults := 0
	for _, a := range list {
		if a.IsDefault {
			defaults++
		}
	}
	if len(list) != 2 || defaults != 1 {
		t.Errorf("list = %d addresses with %d defaults, want 2/1", len(list), defaults)
	}
	if list[0].Label != "Work" {
		t.Errorf("default must sort first, got %q", list[0].Label)
	}
}

func TestAddressService_Validation(t *testing.T) {
	svc := service.NewAddressService(newFakeAddressRepo())
	ctx := context.Background()
	lat := 41.0

	cases := map[string]struct {
		mutate func(*service.AddressInput)
		want   error
	}{
		"missing label":   {func(in *service.AddressInput) { in.Label = "  " }, domain.ErrAddressLabelRequired},
		"missing address": {func(in *service.AddressInput) { in.Address = "" }, domain.ErrAddressRequired},
		"lat without lng": {func(in *service.AddressInput) { in.Latitude = &lat }, domain.ErrCoordinatesIncomplete},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			in := validAddress()
			tc.mutate(&in)
			if _, err := svc.Create(ctx, 1, in); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestAddressService_Ownership(t *testing.T) {
	repo := newFakeAddressRepo()
	svc := service.NewAddressService(repo)
	ctx := context.Background()

	mine, err := svc.Create(ctx, 1, validAddress())
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Another user can neither edit nor delete it.
	if _, err := svc.Update(ctx, 2, mine.ID, validAddress()); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("foreign update = %v, want ErrForbidden", err)
	}
	if err := svc.Delete(ctx, 2, mine.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("foreign delete = %v, want ErrForbidden", err)
	}
	// The owner can.
	if err := svc.Delete(ctx, 1, mine.ID); err != nil {
		t.Errorf("owner delete: %v", err)
	}
	if err := svc.Delete(ctx, 1, mine.ID); !errors.Is(err, domain.ErrAddressNotFound) {
		t.Errorf("double delete = %v, want ErrAddressNotFound", err)
	}
}

// Placing an order from a saved address snapshots its text; other users'
// addresses are rejected; combining address_id with a one-off address is
// ambiguous and rejected.
func TestOrderService_PlaceWithSavedAddress(t *testing.T) {
	ctx := context.Background()
	chefRepo := newFakeChefRepo()
	if err := chefRepo.Create(ctx, &domain.Chef{UserID: 1, IsActive: true}); err != nil {
		t.Fatalf("seed chef: %v", err)
	}
	items := newFakeMenuItemRepo()
	item := seedItem(t, items, 1, 5, 10)
	addresses := newFakeAddressRepo()
	svc := service.NewOrderService(newFakeOrderRepo(), items, chefRepo, addresses, nil, nil, nil, nil)

	lat, lng := 41.0, 29.0
	saved, err := service.NewAddressService(addresses).Create(ctx, 100, service.AddressInput{
		Label: "Home", Address: "5 Saved St", City: "Istanbul", Latitude: &lat, Longitude: &lng,
	})
	if err != nil {
		t.Fatalf("seed address: %v", err)
	}

	lines := []service.OrderLineInput{{MenuItemID: item.ID, Quantity: 1}}

	order, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		AddressID: &saved.ID, PaymentMethod: domain.PaymentMethodCash, Lines: lines,
	})
	if err != nil {
		t.Fatalf("place: %v", err)
	}
	if order.DeliveryAddress != "5 Saved St" || order.DeliveryCity == nil || *order.DeliveryCity != "Istanbul" {
		t.Errorf("address not snapshotted: %+v", order)
	}
	if order.DeliveryLatitude == nil || *order.DeliveryLatitude != lat {
		t.Errorf("coordinates not snapshotted: %+v", order.DeliveryLatitude)
	}

	// Another user's saved address -> 403.
	if _, err := svc.PlaceOrder(ctx, 200, service.PlaceOrderInput{
		AddressID: &saved.ID, PaymentMethod: domain.PaymentMethodCash, Lines: lines,
	}); !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("foreign address = %v, want ErrForbidden", err)
	}

	// Unknown id -> not found.
	ghost := 999
	if _, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		AddressID: &ghost, PaymentMethod: domain.PaymentMethodCash, Lines: lines,
	}); !errors.Is(err, domain.ErrAddressNotFound) {
		t.Errorf("ghost address = %v, want ErrAddressNotFound", err)
	}

	// Both address_id and a one-off address -> validation error.
	if _, err := svc.PlaceOrder(ctx, 100, service.PlaceOrderInput{
		AddressID: &saved.ID, DeliveryAddress: "elsewhere",
		PaymentMethod: domain.PaymentMethodCash, Lines: lines,
	}); !isValidation(err) {
		t.Errorf("ambiguous address = %v, want ValidationError", err)
	}
}
