package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChefRepository is the PostgreSQL adapter for domain.ChefRepository.
type ChefRepository struct {
	db *sql.DB
}

// NewChefRepository builds a ChefRepository.
func NewChefRepository(db *sql.DB) *ChefRepository {
	return &ChefRepository{db: db}
}

const chefColumns = `
	id, user_id, business_name, bio, specialty, experience_years, image_url,
	kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
	food_license_number, health_certificate_url, is_verified, verified_at,
	rating, total_reviews, total_orders, is_active, is_accepting_orders, is_online,
	created_at, updated_at`

func scanChef(s interface{ Scan(...any) error }) (*domain.Chef, error) {
	c := &domain.Chef{}
	err := s.Scan(
		&c.ID, &c.UserID, &c.BusinessName, &c.Bio, &c.Specialty, &c.ExperienceYears, &c.ImageURL,
		&c.KitchenAddress, &c.KitchenCity, &c.KitchenLatitude, &c.KitchenLongitude, &c.DeliveryRadius,
		&c.FoodLicenseNumber, &c.HealthCertificateURL, &c.IsVerified, &c.VerifiedAt,
		&c.Rating, &c.TotalReviews, &c.TotalOrders, &c.IsActive, &c.IsAcceptingOrders, &c.IsOnline,
		&c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

// Create inserts a chef and back-fills its generated ID and timestamps.
func (r *ChefRepository) Create(ctx context.Context, c *domain.Chef) error {
	query := `
		INSERT INTO chefs (
			user_id, business_name, bio, specialty, experience_years,
			kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
			food_license_number, health_certificate_url, is_verified, verified_at,
			rating, total_reviews, total_orders, is_active, is_accepting_orders, is_online,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		c.UserID, c.BusinessName, c.Bio, c.Specialty, c.ExperienceYears,
		c.KitchenAddress, c.KitchenCity, c.KitchenLatitude, c.KitchenLongitude, c.DeliveryRadius,
		c.FoodLicenseNumber, c.HealthCertificateURL, c.IsVerified, c.VerifiedAt,
		c.Rating, c.TotalReviews, c.TotalOrders, c.IsActive, c.IsAcceptingOrders, c.IsOnline,
		c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create chef: %w", err)
	}
	return nil
}

// FindByID returns the active chef with the given id, or ErrChefNotFound.
func (r *ChefRepository) FindByID(ctx context.Context, id int) (*domain.Chef, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+chefColumns+` FROM chefs WHERE id = $1 AND is_active = true`, id)
	c, err := scanChef(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrChefNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find chef: %w", err)
	}
	return c, nil
}

// FindByUserID returns the chef owned by userID, or ErrChefNotFound.
func (r *ChefRepository) FindByUserID(ctx context.Context, userID int) (*domain.Chef, error) {
	row := r.db.QueryRowContext(ctx, `SELECT`+chefColumns+` FROM chefs WHERE user_id = $1`, userID)
	c, err := scanChef(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrChefNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find chef by user: %w", err)
	}
	return c, nil
}

// List returns a page of active chefs narrowed/ordered by f. The sort value
// selects a fixed ORDER BY from the chefOrder whitelist (see
// search_repository.go) — it is never interpolated from input.
func (r *ChefRepository) List(ctx context.Context, f domain.ChefListFilters, limit, offset int) ([]*domain.Chef, int, error) {
	const where = ` WHERE is_active = true AND is_accepting_orders = true AND ($1 = false OR is_online = true) AND rating >= $2`
	query := `SELECT` + chefColumns + `
		FROM chefs` + where + `
		ORDER BY ` + chefOrder[f.Sort] + `
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, query, f.OnlineOnly, f.MinRating, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list chefs: %w", err)
	}
	defer rows.Close()

	chefs, err := collectChefs(rows)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM chefs`+where, f.OnlineOnly, f.MinRating).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chefs: %w", err)
	}
	return chefs, total, nil
}

// FindNearby returns active chefs whose delivery radius covers (lat, lng),
// nearest first. Distance is computed in SQL with the Haversine formula.
func (r *ChefRepository) FindNearby(ctx context.Context, lat, lng float64, limit int, onlineOnly bool) ([]*domain.Chef, error) {
	query := `
		SELECT` + chefColumns + `
		FROM chefs
		WHERE is_active = true
		  AND is_accepting_orders = true
		  AND ($4 = false OR is_online = true)
		  AND kitchen_latitude IS NOT NULL
		  AND kitchen_longitude IS NOT NULL
		  AND (6371 * acos(
		        cos(radians($1)) * cos(radians(kitchen_latitude)) *
		        cos(radians(kitchen_longitude) - radians($2)) +
		        sin(radians($1)) * sin(radians(kitchen_latitude))
		      )) <= delivery_radius
		ORDER BY (6371 * acos(
		        cos(radians($1)) * cos(radians(kitchen_latitude)) *
		        cos(radians(kitchen_longitude) - radians($2)) +
		        sin(radians($1)) * sin(radians(kitchen_latitude))
		      )) ASC
		LIMIT $3`

	rows, err := r.db.QueryContext(ctx, query, lat, lng, limit, onlineOnly)
	if err != nil {
		return nil, fmt.Errorf("find nearby chefs: %w", err)
	}
	defer rows.Close()

	return collectChefs(rows)
}

// SetOnline updates a chef's live presence flag.
func (r *ChefRepository) SetOnline(ctx context.Context, chefID int, online bool) error {
	res, err := r.db.ExecContext(ctx, `UPDATE chefs SET is_online = $2, updated_at = now() WHERE id = $1`, chefID, online)
	if err != nil {
		return fmt.Errorf("set chef online: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrChefNotFound
	}
	return nil
}

// SetAcceptingOrders updates the chef's availability (away / vacation mode).
func (r *ChefRepository) SetAcceptingOrders(ctx context.Context, chefID int, accepting bool) error {
	res, err := r.db.ExecContext(ctx, `UPDATE chefs SET is_accepting_orders = $2, updated_at = now() WHERE id = $1`, chefID, accepting)
	if err != nil {
		return fmt.Errorf("set chef accepting orders: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrChefNotFound
	}
	return nil
}

// SetImageURL updates the chef's kitchen photo URL.
func (r *ChefRepository) SetImageURL(ctx context.Context, chefID int, url string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE chefs SET image_url = $2, updated_at = now() WHERE id = $1`, chefID, url)
	if err != nil {
		return fmt.Errorf("set chef image: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrChefNotFound
	}
	return nil
}

// Update persists the chef's editable profile fields.
func (r *ChefRepository) Update(ctx context.Context, c *domain.Chef) error {
	query := `
		UPDATE chefs
		SET business_name = $2, bio = $3, specialty = $4,
		    kitchen_address = $5, kitchen_city = $6,
		    kitchen_latitude = $7, kitchen_longitude = $8,
		    delivery_radius = $9, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`

	err := r.db.QueryRowContext(
		ctx, query,
		c.ID, c.BusinessName, c.Bio, c.Specialty,
		c.KitchenAddress, c.KitchenCity, c.KitchenLatitude, c.KitchenLongitude,
		c.DeliveryRadius,
	).Scan(&c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrChefNotFound
	}
	if err != nil {
		return fmt.Errorf("update chef: %w", err)
	}
	return nil
}

func collectChefs(rows *sql.Rows) ([]*domain.Chef, error) {
	chefs := make([]*domain.Chef, 0)
	for rows.Next() {
		c, err := scanChef(rows)
		if err != nil {
			return nil, fmt.Errorf("scan chef: %w", err)
		}
		chefs = append(chefs, c)
	}
	return chefs, rows.Err()
}
