package repository

import (
	"database/sql"
	"fmt"
	
	"github.com/Yasin4261/food-delivery/internal/domain"
)

// UserRepository implements domain.UserRepository interface
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (
			username, email, password_hash, phone_number,
			address, city, state, zip_code, latitude, longitude,
			role, is_verified, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.PhoneNumber,
		user.Address,
		user.City,
		user.State,
		user.ZipCode,
		user.Latitude,
		user.Longitude,
		user.Role,
		user.IsVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone_number,
		       address, city, state, zip_code, latitude, longitude,
		       role, is_verified, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND is_active = true
	`
	
	user := &domain.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.PhoneNumber,
		&user.Address,
		&user.City,
		&user.State,
		&user.ZipCode,
		&user.Latitude,
		&user.Longitude,
		&user.Role,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	
	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone_number,
		       address, city, state, zip_code, latitude, longitude,
		       role, is_verified, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	
	user := &domain.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.PhoneNumber,
		&user.Address,
		&user.City,
		&user.State,
		&user.ZipCode,
		&user.Latitude,
		&user.Longitude,
		&user.Role,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	
	return user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone_number,
		       address, city, state, zip_code, latitude, longitude,
		       role, is_verified, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	
	user := &domain.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.PhoneNumber,
		&user.Address,
		&user.City,
		&user.State,
		&user.ZipCode,
		&user.Latitude,
		&user.Longitude,
		&user.Role,
		&user.IsVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	
	return user, nil
}

// Update updates user information
func (r *UserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, phone_number = $3,
		    address = $4, city = $5, state = $6, zip_code = $7,
		    latitude = $8, longitude = $9, role = $10,
		    is_verified = $11, is_active = $12, updated_at = $13
		WHERE id = $14
	`
	
	result, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.PhoneNumber,
		user.Address,
		user.City,
		user.State,
		user.ZipCode,
		user.Latitude,
		user.Longitude,
		user.Role,
		user.IsVerified,
		user.IsActive,
		user.UpdatedAt,
		user.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id int) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// UpdateLocation updates user's location
func (r *UserRepository) UpdateLocation(id int, lat, lng float64, address, city, state, zipCode string) error {
	query := `
		UPDATE users
		SET latitude = $1, longitude = $2, address = $3,
		    city = $4, state = $5, zip_code = $6, updated_at = NOW()
		WHERE id = $7
	`
	
	result, err := r.db.Exec(query, lat, lng, address, city, state, zipCode, id)
	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

// List returns paginated users
func (r *UserRepository) List(offset, limit int) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone_number,
		       address, city, state, zip_code, latitude, longitude,
		       role, is_verified, is_active, created_at, updated_at
		FROM users
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()
	
	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.PhoneNumber,
			&user.Address,
			&user.City,
			&user.State,
			&user.ZipCode,
			&user.Latitude,
			&user.Longitude,
			&user.Role,
			&user.IsVerified,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return users, nil
}

// CountByRole counts users by role
func (r *UserRepository) CountByRole(role string) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE role = $1 AND is_active = true`
	
	var count int
	err := r.db.QueryRow(query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	
	return count, nil
}

// FindNearby finds users within radius (km) from a location using Haversine formula
func (r *UserRepository) FindNearby(lat, lng, radiusKm float64, limit int) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, phone_number,
		       address, city, state, zip_code, latitude, longitude,
		       role, is_verified, is_active, created_at, updated_at,
		       (6371 * acos(
		           cos(radians($1)) * cos(radians(latitude)) *
		           cos(radians(longitude) - radians($2)) +
		           sin(radians($1)) * sin(radians(latitude))
		       )) AS distance
		FROM users
		WHERE latitude IS NOT NULL 
		  AND longitude IS NOT NULL
		  AND is_active = true
		HAVING distance <= $3
		ORDER BY distance
		LIMIT $4
	`
	
	rows, err := r.db.Query(query, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby users: %w", err)
	}
	defer rows.Close()
	
	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		var distance float64
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.PhoneNumber,
			&user.Address,
			&user.City,
			&user.State,
			&user.ZipCode,
			&user.Latitude,
			&user.Longitude,
			&user.Role,
			&user.IsVerified,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
			&distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return users, nil
}
