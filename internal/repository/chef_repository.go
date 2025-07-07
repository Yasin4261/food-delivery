package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"time"
)

// ChefRepository - chef veritabanı işlemleri
type ChefRepository struct {
	db *sql.DB
}

func NewChefRepository(db *sql.DB) *ChefRepository {
	return &ChefRepository{db: db}
}

func (r *ChefRepository) Create(chef *model.Chef) error {
	query := `
		INSERT INTO chefs (user_id, kitchen_name, description, speciality, experience, 
			address, district, city, latitude, longitude, is_active, is_verified, 
			rating, total_orders, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, chef.UserID, chef.KitchenName, chef.Description, 
		chef.Speciality, chef.Experience, chef.Address, chef.District, chef.City,
		chef.Latitude, chef.Longitude, chef.IsActive, chef.IsVerified, 
		chef.Rating, chef.TotalOrders, now, now).Scan(&chef.ID)
	
	if err != nil {
		return err
	}
	
	chef.CreatedAt = now
	chef.UpdatedAt = now
	return nil
}

func (r *ChefRepository) GetByID(id uint) (*model.Chef, error) {
	chef := &model.Chef{}
	query := `
		SELECT id, user_id, kitchen_name, description, speciality, experience,
			address, district, city, latitude, longitude, is_active, is_verified,
			rating, total_orders, created_at, updated_at
		FROM chefs 
		WHERE id = $1`
	
	err := r.db.QueryRow(query, id).Scan(
		&chef.ID, &chef.UserID, &chef.KitchenName, &chef.Description, &chef.Speciality,
		&chef.Experience, &chef.Address, &chef.District, &chef.City, &chef.Latitude,
		&chef.Longitude, &chef.IsActive, &chef.IsVerified, &chef.Rating, &chef.TotalOrders,
		&chef.CreatedAt, &chef.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return chef, nil
}

func (r *ChefRepository) GetByUserID(userID uint) (*model.Chef, error) {
	chef := &model.Chef{}
	query := `
		SELECT id, user_id, kitchen_name, description, speciality, experience,
			address, district, city, latitude, longitude, is_active, is_verified,
			rating, total_orders, created_at, updated_at
		FROM chefs 
		WHERE user_id = $1`
	
	err := r.db.QueryRow(query, userID).Scan(
		&chef.ID, &chef.UserID, &chef.KitchenName, &chef.Description, &chef.Speciality,
		&chef.Experience, &chef.Address, &chef.District, &chef.City, &chef.Latitude,
		&chef.Longitude, &chef.IsActive, &chef.IsVerified, &chef.Rating, &chef.TotalOrders,
		&chef.CreatedAt, &chef.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return chef, nil
}

func (r *ChefRepository) Update(chef *model.Chef) error {
	query := `
		UPDATE chefs 
		SET kitchen_name = $1, description = $2, speciality = $3, experience = $4,
			address = $5, district = $6, city = $7, latitude = $8, longitude = $9,
			is_active = $10, is_verified = $11, rating = $12, total_orders = $13,
			updated_at = $14
		WHERE id = $15`
	
	chef.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, chef.KitchenName, chef.Description, chef.Speciality,
		chef.Experience, chef.Address, chef.District, chef.City, chef.Latitude,
		chef.Longitude, chef.IsActive, chef.IsVerified, chef.Rating, chef.TotalOrders,
		chef.UpdatedAt, chef.ID)
	return err
}

func (r *ChefRepository) Delete(id uint) error {
	query := `DELETE FROM chefs WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ChefRepository) GetByLocation(city, district string) ([]model.Chef, error) {
	query := `
		SELECT id, user_id, kitchen_name, description, speciality, experience,
			address, district, city, latitude, longitude, is_active, is_verified,
			rating, total_orders, created_at, updated_at
		FROM chefs 
		WHERE city = $1 AND district = $2 AND is_active = true
		ORDER BY rating DESC, total_orders DESC`
	
	rows, err := r.db.Query(query, city, district)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chefs []model.Chef
	for rows.Next() {
		var chef model.Chef
		err := rows.Scan(
			&chef.ID, &chef.UserID, &chef.KitchenName, &chef.Description, &chef.Speciality,
			&chef.Experience, &chef.Address, &chef.District, &chef.City, &chef.Latitude,
			&chef.Longitude, &chef.IsActive, &chef.IsVerified, &chef.Rating, &chef.TotalOrders,
			&chef.CreatedAt, &chef.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, nil
}

func (r *ChefRepository) GetActiveChefs() ([]model.Chef, error) {
	query := `
		SELECT id, user_id, kitchen_name, description, speciality, experience,
			address, district, city, latitude, longitude, is_active, is_verified,
			rating, total_orders, created_at, updated_at
		FROM chefs 
		WHERE is_active = true
		ORDER BY rating DESC, total_orders DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chefs []model.Chef
	for rows.Next() {
		var chef model.Chef
		err := rows.Scan(
			&chef.ID, &chef.UserID, &chef.KitchenName, &chef.Description, &chef.Speciality,
			&chef.Experience, &chef.Address, &chef.District, &chef.City, &chef.Latitude,
			&chef.Longitude, &chef.IsActive, &chef.IsVerified, &chef.Rating, &chef.TotalOrders,
			&chef.CreatedAt, &chef.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, nil
}

func (r *ChefRepository) GetWithMeals(chefID uint) (*model.Chef, error) {
	// Bu metod için meal repository ile join yapılması gerekir
	// Şu an basit implementasyon
	return r.GetByID(chefID)
}

func (r *ChefRepository) GetAll() ([]model.Chef, error) {
	query := `
		SELECT id, user_id, kitchen_name, description, speciality, experience,
			address, district, city, latitude, longitude, is_active, is_verified,
			rating, total_orders, created_at, updated_at
		FROM chefs 
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chefs []model.Chef
	for rows.Next() {
		var chef model.Chef
		err := rows.Scan(
			&chef.ID, &chef.UserID, &chef.KitchenName, &chef.Description, &chef.Speciality,
			&chef.Experience, &chef.Address, &chef.District, &chef.City, &chef.Latitude,
			&chef.Longitude, &chef.IsActive, &chef.IsVerified, &chef.Rating, &chef.TotalOrders,
			&chef.CreatedAt, &chef.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		chefs = append(chefs, chef)
	}
	
	return chefs, nil
}
