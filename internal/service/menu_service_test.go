package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// --- in-memory fakes for the menu ports ---

type fakeMenuRepo struct {
	menus  map[int]*domain.Menu
	nextID int
}

func newFakeMenuRepo() *fakeMenuRepo {
	return &fakeMenuRepo{menus: map[int]*domain.Menu{}, nextID: 1}
}

func (f *fakeMenuRepo) Create(_ context.Context, m *domain.Menu) error {
	m.ID = f.nextID
	f.nextID++
	cp := *m
	f.menus[m.ID] = &cp
	return nil
}
func (f *fakeMenuRepo) FindByID(_ context.Context, id int) (*domain.Menu, error) {
	if m, ok := f.menus[id]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, domain.ErrMenuNotFound
}
func (f *fakeMenuRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.Menu, error) {
	out := make([]*domain.Menu, 0)
	for _, m := range f.menus {
		if m.ChefID == chefID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeMenuRepo) Update(_ context.Context, m *domain.Menu) error {
	if _, ok := f.menus[m.ID]; !ok {
		return domain.ErrMenuNotFound
	}
	cp := *m
	f.menus[m.ID] = &cp
	return nil
}
func (f *fakeMenuRepo) Deactivate(_ context.Context, id int) error {
	m, ok := f.menus[id]
	if !ok {
		return domain.ErrMenuNotFound
	}
	m.IsActive = false
	return nil
}

type fakeMenuItemRepo struct {
	items  map[int]*domain.MenuItem
	nextID int
}

func newFakeMenuItemRepo() *fakeMenuItemRepo {
	return &fakeMenuItemRepo{items: map[int]*domain.MenuItem{}, nextID: 1}
}

func (f *fakeMenuItemRepo) Create(_ context.Context, m *domain.MenuItem) error {
	m.ID = f.nextID
	f.nextID++
	cp := *m
	f.items[m.ID] = &cp
	return nil
}
func (f *fakeMenuItemRepo) FindByID(_ context.Context, id int) (*domain.MenuItem, error) {
	if m, ok := f.items[id]; ok {
		cp := *m
		return &cp, nil
	}
	return nil, domain.ErrMenuItemNotFound
}
func (f *fakeMenuItemRepo) ListByMenu(_ context.Context, menuID int) ([]*domain.MenuItem, error) {
	out := make([]*domain.MenuItem, 0)
	for _, m := range f.items {
		if m.MenuID == menuID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeMenuItemRepo) ListByChef(_ context.Context, chefID, limit, offset int) ([]*domain.MenuItem, error) {
	out := make([]*domain.MenuItem, 0)
	for _, m := range f.items {
		if m.ChefID == chefID && m.IsActive {
			cp := *m
			out = append(out, &cp)
		}
	}
	return out, nil
}
func (f *fakeMenuItemRepo) Update(_ context.Context, m *domain.MenuItem) error {
	if _, ok := f.items[m.ID]; !ok {
		return domain.ErrMenuItemNotFound
	}
	cp := *m
	f.items[m.ID] = &cp
	return nil
}
func (f *fakeMenuItemRepo) Deactivate(_ context.Context, id int) error {
	m, ok := f.items[id]
	if !ok {
		return domain.ErrMenuItemNotFound
	}
	m.IsActive = false
	return nil
}
func (f *fakeMenuItemRepo) DecrementStock(_ context.Context, id, qty int) error {
	m, ok := f.items[id]
	if !ok || m.IsUnlimited || m.AvailableQuantity == nil || *m.AvailableQuantity < qty {
		return domain.ErrItemOutOfStock
	}
	*m.AvailableQuantity -= qty
	return nil
}

// menuFixture wires a MenuService over fakes and seeds chef profiles for the
// given user ids (chef.ID is assigned in order). It returns the service and the
// chef repo so tests can map user -> chef id.
func menuFixture(t *testing.T, userIDs ...int) (*service.MenuService, *fakeChefRepo) {
	t.Helper()
	chefRepo := newFakeChefRepo()
	for _, uid := range userIDs {
		if err := chefRepo.Create(context.Background(), &domain.Chef{UserID: uid, IsActive: true}); err != nil {
			t.Fatalf("seed chef: %v", err)
		}
	}
	svc := service.NewMenuService(chefRepo, newFakeMenuRepo(), newFakeMenuItemRepo())
	return svc, chefRepo
}

func TestMenuService_CreateMenu_Success(t *testing.T) {
	svc, _ := menuFixture(t, 1)
	menu, err := svc.CreateMenu(context.Background(), 1, service.CreateMenuInput{Name: "Dinner"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menu.ID == 0 || menu.ChefID != 1 || menu.MenuType != domain.MenuTypeRegular {
		t.Errorf("unexpected menu: %+v", menu)
	}
}

func TestMenuService_CreateMenu_RequiresProfile(t *testing.T) {
	svc, _ := menuFixture(t) // no chef profiles
	_, err := svc.CreateMenu(context.Background(), 99, service.CreateMenuInput{Name: "Dinner"})
	if !errors.Is(err, domain.ErrChefNotFound) {
		t.Errorf("err = %v, want ErrChefNotFound", err)
	}
}

func TestMenuService_CreateMenu_Validation(t *testing.T) {
	svc, _ := menuFixture(t, 1)
	ctx := context.Background()
	if _, err := svc.CreateMenu(ctx, 1, service.CreateMenuInput{Name: " "}); !isValidation(err) {
		t.Errorf("blank name err = %v, want ValidationError", err)
	}
	if _, err := svc.CreateMenu(ctx, 1, service.CreateMenuInput{Name: "X", MenuType: "brunch"}); !isValidation(err) {
		t.Errorf("bad menu_type err = %v, want ValidationError", err)
	}
}

func TestMenuService_UpdateMenu_OwnershipEnforced(t *testing.T) {
	svc, _ := menuFixture(t, 1, 2) // user1->chef1, user2->chef2
	ctx := context.Background()
	menu, err := svc.CreateMenu(ctx, 1, service.CreateMenuInput{Name: "Dinner"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Owner can update.
	if _, err := svc.UpdateMenu(ctx, 1, menu.ID, service.CreateMenuInput{Name: "Supper"}); err != nil {
		t.Fatalf("owner update failed: %v", err)
	}
	// A different chef cannot.
	_, err = svc.UpdateMenu(ctx, 2, menu.ID, service.CreateMenuInput{Name: "Hijack"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}
}

func TestMenuService_CreateItem_OwnershipAndStamp(t *testing.T) {
	svc, _ := menuFixture(t, 1, 2)
	ctx := context.Background()
	menu, err := svc.CreateMenu(ctx, 1, service.CreateMenuInput{Name: "Dinner"})
	if err != nil {
		t.Fatalf("create menu: %v", err)
	}

	// Owner adds an item; chef_id is stamped from the menu, not trusted input.
	item, err := svc.CreateItem(ctx, 1, service.CreateItemInput{MenuID: menu.ID, Name: "Soup", Price: 4.5})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if item.ChefID != 1 || item.MenuID != menu.ID {
		t.Errorf("unexpected item: %+v", item)
	}

	// Another chef cannot add to this menu.
	_, err = svc.CreateItem(ctx, 2, service.CreateItemInput{MenuID: menu.ID, Name: "X", Price: 1})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("err = %v, want ErrForbidden", err)
	}

	// Price must be positive.
	if _, err := svc.CreateItem(ctx, 1, service.CreateItemInput{MenuID: menu.ID, Name: "Free", Price: 0}); !isValidation(err) {
		t.Errorf("zero price err = %v, want ValidationError", err)
	}
}

func TestMenuService_DeactivateMenu(t *testing.T) {
	svc, _ := menuFixture(t, 1)
	ctx := context.Background()
	menu, _ := svc.CreateMenu(ctx, 1, service.CreateMenuInput{Name: "Dinner"})

	if err := svc.DeactivateMenu(ctx, 1, menu.ID); err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	// Now hidden from the public getter.
	if _, err := svc.GetMenu(ctx, menu.ID); !errors.Is(err, domain.ErrMenuNotFound) {
		t.Errorf("get after deactivate err = %v, want ErrMenuNotFound", err)
	}
}

func isValidation(err error) bool {
	var ve service.ValidationError
	return errors.As(err, &ve)
}
