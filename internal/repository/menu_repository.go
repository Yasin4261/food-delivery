package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// MenuRepository is the PostgreSQL adapter for domain.MenuRepository.
type MenuRepository struct {
	db *sql.DB
}

// NewMenuRepository builds a MenuRepository.
func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

const menuColumns = `
	id, chef_id, name, description, menu_type, available_days,
	is_active, is_featured, created_at, updated_at`

func scanMenu(s interface{ Scan(...any) error }) (*domain.Menu, error) {
	m := &domain.Menu{}
	err := s.Scan(
		&m.ID, &m.ChefID, &m.Name, &m.Description, &m.MenuType, &m.AvailableDays,
		&m.IsActive, &m.IsFeatured, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

// Create inserts a menu and back-fills its generated id and timestamps.
func (r *MenuRepository) Create(ctx context.Context, m *domain.Menu) error {
	query := `
		INSERT INTO menus (chef_id, name, description, menu_type, available_days, is_active, is_featured)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		m.ChefID, m.Name, m.Description, m.MenuType, m.AvailableDays, m.IsActive, m.IsFeatured,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create menu: %w", err)
	}
	return nil
}

// FindByID returns the menu with the given id, or ErrMenuNotFound.
func (r *MenuRepository) FindByID(ctx context.Context, id int) (*domain.Menu, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+menuColumns+` FROM menus WHERE id = $1`, id)
	m, err := scanMenu(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrMenuNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find menu: %w", err)
	}
	return m, nil
}

// ListByChef returns a page of a chef's active menus (featured first then
// newest) plus the total.
func (r *MenuRepository) ListByChef(ctx context.Context, chefID, limit, offset int) ([]*domain.Menu, int, error) {
	query := `SELECT` + menuColumns + `
		FROM menus WHERE chef_id = $1 AND is_active = true
		ORDER BY is_featured DESC, created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, chefID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list menus: %w", err)
	}
	defer rows.Close()

	menus := make([]*domain.Menu, 0)
	for rows.Next() {
		m, err := scanMenu(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan menu: %w", err)
		}
		menus = append(menus, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM menus WHERE chef_id = $1 AND is_active = true`, chefID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count menus: %w", err)
	}
	return menus, total, nil
}

// Update writes the editable fields of a menu and refreshes updated_at.
func (r *MenuRepository) Update(ctx context.Context, m *domain.Menu) error {
	query := `
		UPDATE menus
		SET name = $2, description = $3, menu_type = $4, available_days = $5,
		    is_featured = $6, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		m.ID, m.Name, m.Description, m.MenuType, m.AvailableDays, m.IsFeatured,
	).Scan(&m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrMenuNotFound
	}
	if err != nil {
		return fmt.Errorf("update menu: %w", err)
	}
	return nil
}

// Deactivate soft-deletes a menu.
func (r *MenuRepository) Deactivate(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `UPDATE menus SET is_active = false, updated_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deactivate menu: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrMenuNotFound
	}
	return nil
}
