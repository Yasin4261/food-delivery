package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// MenuItemRepository is the PostgreSQL adapter for domain.MenuItemRepository.
type MenuItemRepository struct {
	db *sql.DB
}

// NewMenuItemRepository builds a MenuItemRepository.
func NewMenuItemRepository(db *sql.DB) *MenuItemRepository {
	return &MenuItemRepository{db: db}
}

// menuItemColumns lists every column in the exact order scanMenuItem expects.
const menuItemColumns = `
	id, menu_id, chef_id, name, description, category, cuisine,
	price, original_price, portion_size, preparation_time, serving_size,
	available_quantity, is_unlimited, daily_limit,
	is_vegetarian, is_vegan, is_gluten_free, is_halal, is_spicy, spice_level,
	calories, protein, carbs, fat, image_url, images,
	rating, total_reviews, total_orders, views,
	is_active, is_featured, is_available, created_at, updated_at`

func scanMenuItem(s interface{ Scan(...any) error }) (*domain.MenuItem, error) {
	m := &domain.MenuItem{}
	err := s.Scan(
		&m.ID, &m.MenuID, &m.ChefID, &m.Name, &m.Description, &m.Category, &m.Cuisine,
		&m.Price, &m.OriginalPrice, &m.PortionSize, &m.PreparationTime, &m.ServingSize,
		&m.AvailableQuantity, &m.IsUnlimited, &m.DailyLimit,
		&m.IsVegetarian, &m.IsVegan, &m.IsGlutenFree, &m.IsHalal, &m.IsSpicy, &m.SpiceLevel,
		&m.Calories, &m.Protein, &m.Carbs, &m.Fat, &m.ImageURL, &m.Images,
		&m.Rating, &m.TotalReviews, &m.TotalOrders, &m.Views,
		&m.IsActive, &m.IsFeatured, &m.IsAvailable, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

// Create inserts a dish and back-fills DB-managed columns.
func (r *MenuItemRepository) Create(ctx context.Context, m *domain.MenuItem) error {
	query := `
		INSERT INTO menu_items (
			menu_id, chef_id, name, description, category, cuisine,
			price, original_price, portion_size, preparation_time, serving_size,
			available_quantity, is_unlimited,
			is_vegetarian, is_vegan, is_gluten_free, is_halal, is_spicy, spice_level,
			image_url, is_active, is_featured, is_available
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13,
			$14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23
		)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		m.MenuID, m.ChefID, m.Name, m.Description, m.Category, m.Cuisine,
		m.Price, m.OriginalPrice, m.PortionSize, m.PreparationTime, m.ServingSize,
		m.AvailableQuantity, m.IsUnlimited,
		m.IsVegetarian, m.IsVegan, m.IsGlutenFree, m.IsHalal, m.IsSpicy, m.SpiceLevel,
		m.ImageURL, m.IsActive, m.IsFeatured, m.IsAvailable,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create menu item: %w", err)
	}
	return nil
}

// FindByID returns the dish with the given id, or ErrMenuItemNotFound.
func (r *MenuItemRepository) FindByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+menuItemColumns+` FROM menu_items WHERE id = $1`, id)
	m, err := scanMenuItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrMenuItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find menu item: %w", err)
	}
	return m, nil
}

// ListByMenu returns the active dishes in a menu, featured first.
func (r *MenuItemRepository) ListByMenu(ctx context.Context, menuID int) ([]*domain.MenuItem, error) {
	query := `SELECT` + menuItemColumns + `
		FROM menu_items WHERE menu_id = $1 AND is_active = true
		ORDER BY is_featured DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, menuID)
	if err != nil {
		return nil, fmt.Errorf("list menu items: %w", err)
	}
	defer rows.Close()

	return collectMenuItems(rows)
}

// ListByChef returns a page of a chef's active dishes across all menus plus the
// total.
func (r *MenuItemRepository) ListByChef(ctx context.Context, chefID, limit, offset int) ([]*domain.MenuItem, int, error) {
	query := `SELECT` + menuItemColumns + `
		FROM menu_items WHERE chef_id = $1 AND is_active = true
		ORDER BY is_featured DESC, rating DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, chefID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list chef items: %w", err)
	}
	defer rows.Close()

	items, err := collectMenuItems(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM menu_items WHERE chef_id = $1 AND is_active = true`, chefID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chef items: %w", err)
	}
	return items, total, nil
}

// Update writes the editable fields of a dish and refreshes updated_at.
func (r *MenuItemRepository) Update(ctx context.Context, m *domain.MenuItem) error {
	query := `
		UPDATE menu_items SET
			name = $2, description = $3, category = $4, cuisine = $5,
			price = $6, original_price = $7, portion_size = $8, preparation_time = $9, serving_size = $10,
			available_quantity = $11, is_unlimited = $12,
			is_vegetarian = $13, is_vegan = $14, is_gluten_free = $15, is_halal = $16, is_spicy = $17, spice_level = $18,
			image_url = $19, is_featured = $20, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		m.ID, m.Name, m.Description, m.Category, m.Cuisine,
		m.Price, m.OriginalPrice, m.PortionSize, m.PreparationTime, m.ServingSize,
		m.AvailableQuantity, m.IsUnlimited,
		m.IsVegetarian, m.IsVegan, m.IsGlutenFree, m.IsHalal, m.IsSpicy, m.SpiceLevel,
		m.ImageURL, m.IsFeatured,
	).Scan(&m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrMenuItemNotFound
	}
	if err != nil {
		return fmt.Errorf("update menu item: %w", err)
	}
	return nil
}

// Deactivate soft-deletes a dish.
func (r *MenuItemRepository) Deactivate(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `UPDATE menu_items SET is_active = false, updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deactivate menu item: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrMenuItemNotFound
	}
	return nil
}

// DecrementStock atomically reduces a limited item's available_quantity,
// refusing to go negative. It returns ErrItemOutOfStock when no row matches
// (insufficient stock, or the item is unlimited / has no tracked quantity).
func (r *MenuItemRepository) DecrementStock(ctx context.Context, id, qty int) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE menu_items
		SET available_quantity = available_quantity - $2, updated_at = now()
		WHERE id = $1 AND is_unlimited = false
		  AND available_quantity IS NOT NULL AND available_quantity >= $2`, id, qty)
	if err != nil {
		return fmt.Errorf("decrement stock: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrItemOutOfStock
	}
	return nil
}

func collectMenuItems(rows *sql.Rows) ([]*domain.MenuItem, error) {
	items := make([]*domain.MenuItem, 0)
	for rows.Next() {
		m, err := scanMenuItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}
