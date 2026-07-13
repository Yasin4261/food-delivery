package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// UserRepository is the PostgreSQL adapter for domain.UserRepository.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository builds a UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// userColumns is the shared SELECT list, kept in one place so every read scans
// the same fields in the same order.
const userColumns = `
	id, username, email, password_hash, phone_number,
	address, city, state, zip_code, latitude, longitude,
	role, is_verified, is_active, email_notifications, created_at, updated_at`

// Create inserts a user and back-fills its generated ID and timestamps.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			username, email, password_hash, phone_number,
			address, city, state, zip_code, latitude, longitude,
			role, is_verified, is_active, email_notifications, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		user.Username, user.Email, user.PasswordHash, user.PhoneNumber,
		user.Address, user.City, user.State, user.ZipCode, user.Latitude, user.Longitude,
		user.Role, user.IsVerified, user.IsActive, user.EmailNotifications, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// FindByID returns the user with the given id, or domain.ErrUserNotFound.
func (r *UserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	return r.findOne(ctx, `SELECT`+userColumns+` FROM users WHERE id = $1`, id)
}

// FindByEmail returns the user with the given email, or domain.ErrUserNotFound.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, `SELECT`+userColumns+` FROM users WHERE email = $1`, email)
}

// FindByUsername returns the user with the given username, or domain.ErrUserNotFound.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.findOne(ctx, `SELECT`+userColumns+` FROM users WHERE username = $1`, username)
}

// UpdatePassword sets a user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, passwordHash string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// UpdateProfile persists the editable contact/location fields.
func (r *UserRepository) UpdateProfile(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET phone_number = $2, address = $3, city = $4, state = $5, zip_code = $6,
		    latitude = $7, longitude = $8, email_notifications = $9, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		u.ID, u.PhoneNumber, u.Address, u.City, u.State, u.ZipCode, u.Latitude, u.Longitude,
		u.EmailNotifications,
	).Scan(&u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUserNotFound
	}
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	return nil
}

func scanUser(s interface{ Scan(...any) error }) (*domain.User, error) {
	u := &domain.User{}
	err := s.Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.PhoneNumber,
		&u.Address, &u.City, &u.State, &u.ZipCode, &u.Latitude, &u.Longitude,
		&u.Role, &u.IsVerified, &u.IsActive, &u.EmailNotifications, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func (r *UserRepository) findOne(ctx context.Context, query string, arg any) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRowContext(ctx, query, arg))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}
