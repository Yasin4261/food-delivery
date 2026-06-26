package service

import (
	"context"
	"strings"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// MenuService implements the menu/dish use cases. Chefs own their menus and
// items; mutating use cases resolve the caller's chef profile and enforce that
// the caller owns the resource. It depends only on domain ports.
type MenuService struct {
	chefs domain.ChefRepository
	menus domain.MenuRepository
	items domain.MenuItemRepository
}

// NewMenuService builds a MenuService.
func NewMenuService(chefs domain.ChefRepository, menus domain.MenuRepository, items domain.MenuItemRepository) *MenuService {
	return &MenuService{chefs: chefs, menus: menus, items: items}
}

// CreateMenuInput is the data needed to create or replace a menu's editable
// fields.
type CreateMenuInput struct {
	Name          string
	Description   string
	MenuType      string
	AvailableDays string
	IsFeatured    bool
}

// CreateItemInput is the data needed to create or replace a dish's editable
// fields. MenuID is only read on creation.
type CreateItemInput struct {
	MenuID int

	Name        string
	Description string
	Category    string
	Cuisine     string
	PortionSize string
	ImageURL    string

	Price           float64
	OriginalPrice   *float64
	PreparationTime *int
	ServingSize     *int

	AvailableQuantity *int
	IsUnlimited       bool

	IsVegetarian bool
	IsVegan      bool
	IsGlutenFree bool
	IsHalal      bool
	IsSpicy      bool
	SpiceLevel   *int
}

// chefForUser resolves the chef profile owned by userID. Callers without a
// profile get ErrChefNotFound.
func (s *MenuService) chefForUser(ctx context.Context, userID int) (*domain.Chef, error) {
	return s.chefs.FindByUserID(ctx, userID)
}

// CreateMenu opens a new menu for the caller's chef profile.
func (s *MenuService) CreateMenu(ctx context.Context, userID int, in CreateMenuInput) (*domain.Menu, error) {
	chef, err := s.chefForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := validateMenu(&in); err != nil {
		return nil, err
	}

	menu := domain.NewMenu(chef.ID, in.Name)
	applyMenuInput(menu, in)

	if err := s.menus.Create(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// UpdateMenu replaces the editable fields of a menu the caller owns.
func (s *MenuService) UpdateMenu(ctx context.Context, userID, menuID int, in CreateMenuInput) (*domain.Menu, error) {
	menu, err := s.ownedMenu(ctx, userID, menuID)
	if err != nil {
		return nil, err
	}
	if err := validateMenu(&in); err != nil {
		return nil, err
	}

	applyMenuInput(menu, in)
	if err := s.menus.Update(ctx, menu); err != nil {
		return nil, err
	}
	return menu, nil
}

// DeactivateMenu soft-deletes a menu the caller owns.
func (s *MenuService) DeactivateMenu(ctx context.Context, userID, menuID int) error {
	if _, err := s.ownedMenu(ctx, userID, menuID); err != nil {
		return err
	}
	return s.menus.Deactivate(ctx, menuID)
}

// GetMenu returns an active menu by id (public).
func (s *MenuService) GetMenu(ctx context.Context, id int) (*domain.Menu, error) {
	menu, err := s.menus.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !menu.IsActive {
		return nil, domain.ErrMenuNotFound
	}
	return menu, nil
}

// ListChefMenus returns a chef's active menus (public).
func (s *MenuService) ListChefMenus(ctx context.Context, chefID, limit, offset int) ([]*domain.Menu, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.menus.ListByChef(ctx, chefID, limit, offset)
}

// CreateItem adds a dish to one of the caller's menus.
func (s *MenuService) CreateItem(ctx context.Context, userID int, in CreateItemInput) (*domain.MenuItem, error) {
	menu, err := s.ownedMenu(ctx, userID, in.MenuID)
	if err != nil {
		return nil, err
	}
	if err := validateItem(&in); err != nil {
		return nil, err
	}

	item := domain.NewMenuItem(menu.ID, menu.ChefID, in.Name, in.Price)
	applyItemInput(item, in)

	if err := s.items.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// UpdateItem replaces the editable fields of a dish the caller owns.
func (s *MenuService) UpdateItem(ctx context.Context, userID, itemID int, in CreateItemInput) (*domain.MenuItem, error) {
	item, err := s.ownedItem(ctx, userID, itemID)
	if err != nil {
		return nil, err
	}
	if err := validateItem(&in); err != nil {
		return nil, err
	}

	applyItemInput(item, in)
	if err := s.items.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// DeactivateItem soft-deletes a dish the caller owns.
func (s *MenuService) DeactivateItem(ctx context.Context, userID, itemID int) error {
	if _, err := s.ownedItem(ctx, userID, itemID); err != nil {
		return err
	}
	return s.items.Deactivate(ctx, itemID)
}

// ListMenuItems returns the active dishes in a menu (public).
func (s *MenuService) ListMenuItems(ctx context.Context, menuID int) ([]*domain.MenuItem, error) {
	return s.items.ListByMenu(ctx, menuID)
}

// ListChefItems returns a chef's active dishes across all menus (public).
func (s *MenuService) ListChefItems(ctx context.Context, chefID, limit, offset int) ([]*domain.MenuItem, error) {
	limit, offset = normalisePaging(limit, offset)
	return s.items.ListByChef(ctx, chefID, limit, offset)
}

// ownedMenu fetches a menu and verifies the caller's chef profile owns it.
func (s *MenuService) ownedMenu(ctx context.Context, userID, menuID int) (*domain.Menu, error) {
	chef, err := s.chefForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	menu, err := s.menus.FindByID(ctx, menuID)
	if err != nil {
		return nil, err
	}
	if menu.ChefID != chef.ID {
		return nil, domain.ErrForbidden
	}
	return menu, nil
}

// ownedItem fetches a dish and verifies the caller's chef profile owns it.
func (s *MenuService) ownedItem(ctx context.Context, userID, itemID int) (*domain.MenuItem, error) {
	chef, err := s.chefForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	item, err := s.items.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.ChefID != chef.ID {
		return nil, domain.ErrForbidden
	}
	return item, nil
}

func validateMenu(in *CreateMenuInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.MenuType = strings.TrimSpace(in.MenuType)
	if in.Name == "" {
		return ValidationError{Msg: "name is required"}
	}
	if in.MenuType != "" && !domain.ValidMenuType(in.MenuType) {
		return ValidationError{Msg: "invalid menu_type: must be regular, daily_special, seasonal or weekend"}
	}
	return nil
}

func applyMenuInput(menu *domain.Menu, in CreateMenuInput) {
	menu.Name = in.Name
	menu.Description = optional(in.Description)
	if in.MenuType != "" {
		menu.MenuType = in.MenuType
	}
	menu.AvailableDays = optional(in.AvailableDays)
	menu.IsFeatured = in.IsFeatured
}

func validateItem(in *CreateItemInput) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return ValidationError{Msg: "name is required"}
	}
	if in.Price <= 0 {
		return ValidationError{Msg: "price must be greater than zero"}
	}
	if in.OriginalPrice != nil && *in.OriginalPrice < 0 {
		return ValidationError{Msg: "original_price cannot be negative"}
	}
	if in.AvailableQuantity != nil && *in.AvailableQuantity < 0 {
		return ValidationError{Msg: "available_quantity cannot be negative"}
	}
	if in.SpiceLevel != nil && (*in.SpiceLevel < 0 || *in.SpiceLevel > 5) {
		return ValidationError{Msg: "spice_level must be between 0 and 5"}
	}
	return nil
}

func applyItemInput(item *domain.MenuItem, in CreateItemInput) {
	item.Name = in.Name
	item.Price = in.Price
	item.Description = optional(in.Description)
	item.Category = optional(in.Category)
	item.Cuisine = optional(in.Cuisine)
	item.PortionSize = optional(in.PortionSize)
	item.ImageURL = optional(in.ImageURL)
	item.OriginalPrice = in.OriginalPrice
	item.PreparationTime = in.PreparationTime
	if in.ServingSize != nil {
		item.ServingSize = *in.ServingSize
	}
	item.AvailableQuantity = in.AvailableQuantity
	item.IsUnlimited = in.IsUnlimited
	item.IsVegetarian = in.IsVegetarian
	item.IsVegan = in.IsVegan
	item.IsGlutenFree = in.IsGlutenFree
	item.IsHalal = in.IsHalal
	item.IsSpicy = in.IsSpicy
	item.SpiceLevel = in.SpiceLevel
}
