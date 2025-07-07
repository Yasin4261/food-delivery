package repository

import (
	"database/sql"
	"ecommerce/internal/model"
	"fmt"
	"strings"
	"time"
)

// MealRepository - yemek veritabanı işlemleri
type MealRepository struct {
	db *sql.DB
}

func NewMealRepository(db *sql.DB) *MealRepository {
	return &MealRepository{db: db}
}

func (r *MealRepository) Create(meal *model.Meal) error {
	query := `
		INSERT INTO meals (chef_id, name, description, category, cuisine, price, currency,
			portion, serving_size, available_quantity, preparation_time, cooking_time,
			calories, ingredients, allergens, is_vegetarian, is_vegan, is_gluten_free,
			is_active, is_available, rating, total_orders, images, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		RETURNING id`
	
	now := time.Now()
	err := r.db.QueryRow(query, meal.ChefID, meal.Name, meal.Description, meal.Category,
		meal.Cuisine, meal.Price, meal.Currency, meal.Portion, meal.ServingSize,
		meal.AvailableQuantity, meal.PreparationTime, meal.CookingTime, meal.Calories,
		meal.Ingredients, meal.Allergens, meal.IsVegetarian, meal.IsVegan, meal.IsGlutenFree,
		meal.IsActive, meal.IsAvailable, meal.Rating, meal.TotalOrders, meal.Images,
		now, now).Scan(&meal.ID)
	
	if err != nil {
		return err
	}
	
	meal.CreatedAt = now
	meal.UpdatedAt = now
	return nil
}

func (r *MealRepository) GetByID(id uint) (*model.Meal, error) {
	meal := &model.Meal{}
	query := `
		SELECT id, chef_id, name, description, category, cuisine, price, currency,
			portion, serving_size, available_quantity, preparation_time, cooking_time,
			calories, ingredients, allergens, is_vegetarian, is_vegan, is_gluten_free,
			is_active, is_available, rating, total_orders, images, created_at, updated_at
		FROM meals 
		WHERE id = $1`
	
	err := r.db.QueryRow(query, id).Scan(
		&meal.ID, &meal.ChefID, &meal.Name, &meal.Description, &meal.Category,
		&meal.Cuisine, &meal.Price, &meal.Currency, &meal.Portion, &meal.ServingSize,
		&meal.AvailableQuantity, &meal.PreparationTime, &meal.CookingTime, &meal.Calories,
		&meal.Ingredients, &meal.Allergens, &meal.IsVegetarian, &meal.IsVegan, &meal.IsGlutenFree,
		&meal.IsActive, &meal.IsAvailable, &meal.Rating, &meal.TotalOrders, &meal.Images,
		&meal.CreatedAt, &meal.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return meal, nil
}

func (r *MealRepository) GetByChefID(chefID uint) ([]model.Meal, error) {
	query := `
		SELECT id, chef_id, name, description, category, cuisine, price, currency,
			portion, serving_size, available_quantity, preparation_time, cooking_time,
			calories, ingredients, allergens, is_vegetarian, is_vegan, is_gluten_free,
			is_active, is_available, rating, total_orders, images, created_at, updated_at
		FROM meals 
		WHERE chef_id = $1 AND is_active = true
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query, chefID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var meals []model.Meal
	for rows.Next() {
		var meal model.Meal
		err := rows.Scan(
			&meal.ID, &meal.ChefID, &meal.Name, &meal.Description, &meal.Category,
			&meal.Cuisine, &meal.Price, &meal.Currency, &meal.Portion, &meal.ServingSize,
			&meal.AvailableQuantity, &meal.PreparationTime, &meal.CookingTime, &meal.Calories,
			&meal.Ingredients, &meal.Allergens, &meal.IsVegetarian, &meal.IsVegan, &meal.IsGlutenFree,
			&meal.IsActive, &meal.IsAvailable, &meal.Rating, &meal.TotalOrders, &meal.Images,
			&meal.CreatedAt, &meal.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, meal)
	}
	
	return meals, nil
}

func (r *MealRepository) Update(meal *model.Meal) error {
	query := `
		UPDATE meals 
		SET name = $1, description = $2, category = $3, cuisine = $4, price = $5,
			portion = $6, serving_size = $7, available_quantity = $8, preparation_time = $9,
			cooking_time = $10, calories = $11, ingredients = $12, allergens = $13,
			is_vegetarian = $14, is_vegan = $15, is_gluten_free = $16, is_active = $17,
			is_available = $18, rating = $19, total_orders = $20, images = $21, updated_at = $22
		WHERE id = $23`
	
	meal.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, meal.Name, meal.Description, meal.Category, meal.Cuisine,
		meal.Price, meal.Portion, meal.ServingSize, meal.AvailableQuantity, meal.PreparationTime,
		meal.CookingTime, meal.Calories, meal.Ingredients, meal.Allergens, meal.IsVegetarian,
		meal.IsVegan, meal.IsGlutenFree, meal.IsActive, meal.IsAvailable, meal.Rating,
		meal.TotalOrders, meal.Images, meal.UpdatedAt, meal.ID)
	return err
}

func (r *MealRepository) Delete(id uint) error {
	query := `DELETE FROM meals WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *MealRepository) GetAvailableWithChef() ([]model.MealWithChef, error) {
	query := `
		SELECT m.id, m.chef_id, m.name, m.description, m.category, m.cuisine, m.price,
			m.currency, m.portion, m.serving_size, m.available_quantity, m.preparation_time,
			m.cooking_time, m.calories, m.ingredients, m.allergens, m.is_vegetarian,
			m.is_vegan, m.is_gluten_free, m.is_active, m.is_available, m.rating,
			m.total_orders, m.images, m.created_at, m.updated_at,
			CONCAT(u.first_name, ' ', u.last_name) as chef_name,
			c.kitchen_name, c.rating as chef_rating, c.district as chef_district, c.city as chef_city
		FROM meals m
		JOIN chefs c ON m.chef_id = c.id
		JOIN users u ON c.user_id = u.id
		WHERE m.is_active = true AND m.is_available = true AND c.is_active = true
		ORDER BY m.rating DESC, m.total_orders DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var meals []model.MealWithChef
	for rows.Next() {
		var meal model.MealWithChef
		err := rows.Scan(
			&meal.ID, &meal.ChefID, &meal.Name, &meal.Description, &meal.Category,
			&meal.Cuisine, &meal.Price, &meal.Currency, &meal.Portion, &meal.ServingSize,
			&meal.AvailableQuantity, &meal.PreparationTime, &meal.CookingTime, &meal.Calories,
			&meal.Ingredients, &meal.Allergens, &meal.IsVegetarian, &meal.IsVegan, &meal.IsGlutenFree,
			&meal.IsActive, &meal.IsAvailable, &meal.Rating, &meal.TotalOrders, &meal.Images,
			&meal.CreatedAt, &meal.UpdatedAt,
			&meal.ChefName, &meal.KitchenName, &meal.ChefRating, &meal.ChefDistrict, &meal.ChefCity,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, meal)
	}
	
	return meals, nil
}

func (r *MealRepository) SearchWithChef(filter *model.MealFilter) ([]model.MealWithChef, error) {
	query := `
		SELECT m.id, m.chef_id, m.name, m.description, m.category, m.cuisine, m.price,
			m.currency, m.portion, m.serving_size, m.available_quantity, m.preparation_time,
			m.cooking_time, m.calories, m.ingredients, m.allergens, m.is_vegetarian,
			m.is_vegan, m.is_gluten_free, m.is_active, m.is_available, m.rating,
			m.total_orders, m.images, m.created_at, m.updated_at,
			CONCAT(u.first_name, ' ', u.last_name) as chef_name,
			c.kitchen_name, c.rating as chef_rating, c.district as chef_district, c.city as chef_city
		FROM meals m
		JOIN chefs c ON m.chef_id = c.id
		JOIN users u ON c.user_id = u.id
		WHERE m.is_active = true AND m.is_available = true AND c.is_active = true`
	
	var conditions []string
	var args []interface{}
	argIndex := 1
	
	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("m.category = $%d", argIndex))
		args = append(args, filter.Category)
		argIndex++
	}
	
	if filter.Cuisine != "" {
		conditions = append(conditions, fmt.Sprintf("m.cuisine = $%d", argIndex))
		args = append(args, filter.Cuisine)
		argIndex++
	}
	
	if filter.MinPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("m.price >= $%d", argIndex))
		args = append(args, filter.MinPrice)
		argIndex++
	}
	
	if filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("m.price <= $%d", argIndex))
		args = append(args, filter.MaxPrice)
		argIndex++
	}
	
	if filter.IsVegetarian {
		conditions = append(conditions, "m.is_vegetarian = true")
	}
	
	if filter.IsVegan {
		conditions = append(conditions, "m.is_vegan = true")
	}
	
	if filter.IsGlutenFree {
		conditions = append(conditions, "m.is_gluten_free = true")
	}
	
	if filter.District != "" {
		conditions = append(conditions, fmt.Sprintf("c.district = $%d", argIndex))
		args = append(args, filter.District)
		argIndex++
	}
	
	if filter.City != "" {
		conditions = append(conditions, fmt.Sprintf("c.city = $%d", argIndex))
		args = append(args, filter.City)
		argIndex++
	}
	
	if filter.Rating > 0 {
		conditions = append(conditions, fmt.Sprintf("m.rating >= $%d", argIndex))
		args = append(args, filter.Rating)
		argIndex++
	}
	
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}
	
	query += " ORDER BY m.rating DESC, m.total_orders DESC"
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var meals []model.MealWithChef
	for rows.Next() {
		var meal model.MealWithChef
		err := rows.Scan(
			&meal.ID, &meal.ChefID, &meal.Name, &meal.Description, &meal.Category,
			&meal.Cuisine, &meal.Price, &meal.Currency, &meal.Portion, &meal.ServingSize,
			&meal.AvailableQuantity, &meal.PreparationTime, &meal.CookingTime, &meal.Calories,
			&meal.Ingredients, &meal.Allergens, &meal.IsVegetarian, &meal.IsVegan, &meal.IsGlutenFree,
			&meal.IsActive, &meal.IsAvailable, &meal.Rating, &meal.TotalOrders, &meal.Images,
			&meal.CreatedAt, &meal.UpdatedAt,
			&meal.ChefName, &meal.KitchenName, &meal.ChefRating, &meal.ChefDistrict, &meal.ChefCity,
		)
		if err != nil {
			return nil, err
		}
		meals = append(meals, meal)
	}
	
	return meals, nil
}

func (r *MealRepository) GetByCategoryWithChef(category string) ([]model.MealWithChef, error) {
	filter := &model.MealFilter{Category: category}
	return r.SearchWithChef(filter)
}

func (r *MealRepository) GetByLocationWithChef(city, district string) ([]model.MealWithChef, error) {
	filter := &model.MealFilter{City: city, District: district}
	return r.SearchWithChef(filter)
}

func (r *MealRepository) GetAll() ([]model.Meal, error) {
	query := `
		SELECT id, chef_id, name, description, category, cuisine, price, currency,
			portion, serving_size, available_quantity, preparation_time, cooking_time,
			calories, ingredients, allergens, is_vegetarian, is_vegan, is_gluten_free,
			is_active, is_available, rating, total_orders, images, created_at, updated_at
		FROM meals
		ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var meals []model.Meal
	for rows.Next() {
		var meal model.Meal
		err := rows.Scan(&meal.ID, &meal.ChefID, &meal.Name, &meal.Description, &meal.Category,
			&meal.Cuisine, &meal.Price, &meal.Currency, &meal.Portion, &meal.ServingSize,
			&meal.AvailableQuantity, &meal.PreparationTime, &meal.CookingTime, &meal.Calories,
			&meal.Ingredients, &meal.Allergens, &meal.IsVegetarian, &meal.IsVegan, &meal.IsGlutenFree,
			&meal.IsActive, &meal.IsAvailable, &meal.Rating, &meal.TotalOrders, &meal.Images,
			&meal.CreatedAt, &meal.UpdatedAt)
		if err != nil {
			return nil, err
		}
		meals = append(meals, meal)
	}
	
	return meals, nil
}
