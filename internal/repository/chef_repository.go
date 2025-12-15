package repository

import (
	"database/sql"
	"fmt"
	
	"github.com/Yasin4261/food-delivery/internal/domain"
)

// ChefRepository implements domain.ChefRepository interface
type ChefRepository struct {
	db *sql.DB
}

// NewChefRepository creates a new chef repository
func NewChefRepository(db *sql.DB) *ChefRepository {
	return &ChefRepository{db: db}
}

// Create creates a new chef profile
func (r *ChefRepository) Create(chef *domain.Chef) error {
	query := `
		INSERT INTO chefs (
			user_id, business_name, bio, specialty, experience_years,
			kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
			food_license_number, health_certificate_url, is_verified, verified_at,
			rating, total_reviews, total_orders, is_active, is_accepting_orders,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at
	`
	
	err := r.db.QueryRow(
		query,
		chef.UserID,
		chef.BusinessName,
		chef.Bio,
		chef.Specialty,
		chef.ExperienceYears,
		chef.KitchenAddress,
		chef.KitchenCity,
		chef.KitchenLatitude,
		chef.KitchenLongitude,
		chef.DeliveryRadius,
		chef.FoodLicenseNumber,
		chef.HealthCertificateURL,
		chef.IsVerified,
		chef.VerifiedAt,
		chef.Rating,
		chef.TotalReviews,
		chef.TotalOrders,
		chef.IsActive,
		chef.IsAcceptingOrders,
		chef.CreatedAt,
		chef.UpdatedAt,
	).Scan(&chef.ID, &chef.CreatedAt, &chef.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create chef: %w", err)
	}
	
	return nil
}

// FindByID finds a chef by ID
func (r *ChefRepository) FindByID(id int) (*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE id = $1 AND is_active = true
	`
	
	chef := &domain.Chef{}
	err := r.db.QueryRow(query, id).Scan(
		&chef.ID,
		&chef.UserID,
		&chef.BusinessName,
		&chef.Bio,
		&chef.Specialty,
		&chef.ExperienceYears,
		&chef.KitchenAddress,
		&chef.KitchenCity,
		&chef.KitchenLatitude,
		&chef.KitchenLongitude,
		&chef.DeliveryRadius,
		&chef.FoodLicenseNumber,
		&chef.HealthCertificateURL,
		&chef.IsVerified,
		&chef.VerifiedAt,
		&chef.Rating,
		&chef.TotalReviews,
		&chef.TotalOrders,
		&chef.IsActive,
		&chef.IsAcceptingOrders,
		&chef.CreatedAt,
		&chef.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chef not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find chef by id: %w", err)
	}
	
	return chef, nil
}

// FindByUserID finds a chef by user ID
func (r *ChefRepository) FindByUserID(userID int) (*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE user_id = $1
	`
	
	chef := &domain.Chef{}
	err := r.db.QueryRow(query, userID).Scan(
		&chef.ID,
		&chef.UserID,
		&chef.BusinessName,
		&chef.Bio,
		&chef.Specialty,
		&chef.ExperienceYears,
		&chef.KitchenAddress,
		&chef.KitchenCity,
		&chef.KitchenLatitude,
		&chef.KitchenLongitude,
		&chef.DeliveryRadius,
		&chef.FoodLicenseNumber,
		&chef.HealthCertificateURL,
		&chef.IsVerified,
		&chef.VerifiedAt,
		&chef.Rating,
		&chef.TotalReviews,
		&chef.TotalOrders,
		&chef.IsActive,
		&chef.IsAcceptingOrders,
		&chef.CreatedAt,
		&chef.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chef not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find chef by user id: %w", err)
	}
	
	return chef, nil
}

// Update updates chef information
func (r *ChefRepository) Update(chef *domain.Chef) error {
	query := `
		UPDATE chefs
		SET business_name = $1, bio = $2, specialty = $3, experience_years = $4,
		    kitchen_address = $5, kitchen_city = $6, kitchen_latitude = $7, kitchen_longitude = $8,
		    delivery_radius = $9, food_license_number = $10, health_certificate_url = $11,
		    is_verified = $12, verified_at = $13, rating = $14, total_reviews = $15,
		    total_orders = $16, is_active = $17, is_accepting_orders = $18, updated_at = $19
		WHERE id = $20
	`
	
	result, err := r.db.Exec(
		query,
		chef.BusinessName,
		chef.Bio,
		chef.Specialty,
		chef.ExperienceYears,
		chef.KitchenAddress,
		chef.KitchenCity,
		chef.KitchenLatitude,
		chef.KitchenLongitude,
		chef.DeliveryRadius,
		chef.FoodLicenseNumber,
		chef.HealthCertificateURL,
		chef.IsVerified,
		chef.VerifiedAt,
		chef.Rating,
		chef.TotalReviews,
		chef.TotalOrders,
		chef.IsActive,
		chef.IsAcceptingOrders,
		chef.UpdatedAt,
		chef.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update chef: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("chef not found")
	}
	
	return nil
}

// Delete soft deletes a chef profile
func (r *ChefRepository) Delete(id int) error {
	query := `UPDATE chefs SET is_active = false, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chef: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("chef not found")
	}
	
	return nil
}

// List returns paginated chefs
func (r *ChefRepository) List(offset, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE is_active = true
		ORDER BY rating DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chefs: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// FindByCity finds chefs in a city
func (r *ChefRepository) FindByCity(city string, offset, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE kitchen_city = $1 AND is_active = true AND is_verified = true
		ORDER BY rating DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, city, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find chefs by city: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// FindNearby finds chefs within delivery radius from a location
func (r *ChefRepository) FindNearby(lat, lng float64, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at,
		       (6371 * acos(
		           cos(radians($1)) * cos(radians(kitchen_latitude)) *
		           cos(radians(kitchen_longitude) - radians($2)) +
		           sin(radians($1)) * sin(radians(kitchen_latitude))
		       )) AS distance
		FROM chefs
		WHERE kitchen_latitude IS NOT NULL 
		  AND kitchen_longitude IS NOT NULL
		  AND is_active = true
		  AND is_verified = true
		  AND is_accepting_orders = true
		HAVING distance <= delivery_radius
		ORDER BY distance, rating DESC
		LIMIT $3
	`
	
	rows, err := r.db.Query(query, lat, lng, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby chefs: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		var distance float64
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
			&distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// FindByRating finds chefs with rating >= minRating
func (r *ChefRepository) FindByRating(minRating float64, offset, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE rating >= $1 AND is_active = true AND is_verified = true
		ORDER BY rating DESC, total_reviews DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(query, minRating, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find chefs by rating: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// FindVerified finds all verified chefs
func (r *ChefRepository) FindVerified(offset, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE is_verified = true AND is_active = true
		ORDER BY rating DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find verified chefs: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// FindAcceptingOrders finds chefs currently accepting orders
func (r *ChefRepository) FindAcceptingOrders(offset, limit int) ([]*domain.Chef, error) {
	query := `
		SELECT id, user_id, business_name, bio, specialty, experience_years,
		       kitchen_address, kitchen_city, kitchen_latitude, kitchen_longitude, delivery_radius,
		       food_license_number, health_certificate_url, is_verified, verified_at,
		       rating, total_reviews, total_orders, is_active, is_accepting_orders,
		       created_at, updated_at
		FROM chefs
		WHERE is_accepting_orders = true AND is_active = true AND is_verified = true
		ORDER BY rating DESC
		LIMIT $1 OFFSET $2
	`
	
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find accepting chefs: %w", err)
	}
	defer rows.Close()
	
	var chefs []*domain.Chef
	for rows.Next() {
		chef := &domain.Chef{}
		err := rows.Scan(
			&chef.ID,
			&chef.UserID,
			&chef.BusinessName,
			&chef.Bio,
			&chef.Specialty,
			&chef.ExperienceYears,
			&chef.KitchenAddress,
			&chef.KitchenCity,
			&chef.KitchenLatitude,
			&chef.KitchenLongitude,
			&chef.DeliveryRadius,
			&chef.FoodLicenseNumber,
			&chef.HealthCertificateURL,
			&chef.IsVerified,
			&chef.VerifiedAt,
			&chef.Rating,
			&chef.TotalReviews,
			&chef.TotalOrders,
			&chef.IsActive,
			&chef.IsAcceptingOrders,
			&chef.CreatedAt,
			&chef.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chef: %w", err)
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, rows.Err()
}

// UpdateRating updates chef rating and review count
func (r *ChefRepository) UpdateRating(id int, rating float64) error {
	query := `
		UPDATE chefs
		SET rating = (rating * total_reviews + $1) / (total_reviews + 1),
		    total_reviews = total_reviews + 1,
		    updated_at = NOW()
		WHERE id = $2
	`
	
	result, err := r.db.Exec(query, rating, id)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("chef not found")
	}
	
	return nil
}

// IncrementOrders increments total orders count
func (r *ChefRepository) IncrementOrders(id int) error {
	query := `
		UPDATE chefs
		SET total_orders = total_orders + 1, updated_at = NOW()
		WHERE id = $1
	`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to increment orders: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("chef not found")
	}
	
	return nil
}
