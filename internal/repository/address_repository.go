package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// AddressRepository is the PostgreSQL adapter for domain.AddressRepository.
// Default handling spans two statements (clear the old default, set the new
// one), so those paths run in a transaction; a partial unique index
// (idx_addresses_one_default) backs the invariant at the database level.
type AddressRepository struct {
	db *sql.DB
}

// NewAddressRepository builds an AddressRepository.
func NewAddressRepository(db *sql.DB) *AddressRepository {
	return &AddressRepository{db: db}
}

const addressColumns = `
	id, user_id, label, address, city, latitude, longitude, is_default,
	created_at, updated_at`

func scanAddress(s interface{ Scan(...any) error }) (*domain.Address, error) {
	a := &domain.Address{}
	err := s.Scan(
		&a.ID, &a.UserID, &a.Label, &a.Address, &a.City, &a.Latitude, &a.Longitude,
		&a.IsDefault, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

// Create persists an address, clearing any previous default first when the
// new one is marked default.
func (r *AddressRepository) Create(ctx context.Context, a *domain.Address) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if a.IsDefault {
		if err := clearDefault(ctx, tx, a.UserID); err != nil {
			return err
		}
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO addresses (user_id, label, address, city, latitude, longitude, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		a.UserID, a.Label, a.Address, a.City, a.Latitude, a.Longitude, a.IsDefault,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create address: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit address: %w", err)
	}
	return nil
}

// FindByID returns one address, or domain.ErrAddressNotFound.
func (r *AddressRepository) FindByID(ctx context.Context, id int) (*domain.Address, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+addressColumns+` FROM addresses WHERE id = $1`, id)
	a, err := scanAddress(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAddressNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find address: %w", err)
	}
	return a, nil
}

// ListByUser returns a user's addresses, default first, then newest.
func (r *AddressRepository) ListByUser(ctx context.Context, userID int) ([]*domain.Address, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT`+addressColumns+`
		FROM addresses
		WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list addresses: %w", err)
	}
	defer rows.Close()

	out := make([]*domain.Address, 0)
	for rows.Next() {
		a, err := scanAddress(rows)
		if err != nil {
			return nil, fmt.Errorf("scan address: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// Update persists the editable fields, clearing any previous default first
// when this address becomes the default.
func (r *AddressRepository) Update(ctx context.Context, a *domain.Address) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if a.IsDefault {
		if err := clearDefault(ctx, tx, a.UserID); err != nil {
			return err
		}
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE addresses
		SET label = $2, address = $3, city = $4, latitude = $5, longitude = $6,
		    is_default = $7, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`,
		a.ID, a.Label, a.Address, a.City, a.Latitude, a.Longitude, a.IsDefault,
	).Scan(&a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrAddressNotFound
	}
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit address: %w", err)
	}
	return nil
}

// Delete removes an address.
func (r *AddressRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM addresses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete address: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAddressNotFound
	}
	return nil
}

func clearDefault(ctx context.Context, tx *sql.Tx, userID int) error {
	if _, err := tx.ExecContext(ctx,
		`UPDATE addresses SET is_default = false, updated_at = now() WHERE user_id = $1 AND is_default`, userID); err != nil {
		return fmt.Errorf("clear default address: %w", err)
	}
	return nil
}
