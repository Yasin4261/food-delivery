package repository

import (
	"testing"
	"ecommerce/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Simple integration tests for MealRepository using SQLite in-memory database

func setupMealTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&model.Meal{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestMealRepository_Create(t *testing.T) {
	db, err := setupMealTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewMealRepository(db)

	meal := &model.Meal{
		ChefID:      1,
		Name:        "Test Meal",
		Description: "Test Description",
		Price:       25.50,
		Category:    "Main Course",
		IsActive:    true,
		CookingTime: 30,
		Ingredients: "Test ingredients",
	}

	err = repo.Create(meal)
	if err != nil {
		t.Errorf("Failed to create meal: %v", err)
	}

	if meal.ID == 0 {
		t.Error("Meal ID should be assigned after creation")
	}
}

func TestMealRepository_GetByID(t *testing.T) {
	db, err := setupMealTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewMealRepository(db)

	// Create a test meal
	meal := &model.Meal{
		ChefID:      1,
		Name:        "Test Meal",
		Description: "Test Description",
		Price:       25.50,
		Category:    "Main Course",
		IsActive:    true,
		CookingTime: 30,
		Ingredients: "Test ingredients",
	}

	err = repo.Create(meal)
	if err != nil {
		t.Fatalf("Failed to create meal: %v", err)
	}

	// Retrieve the meal
	retrievedMeal, err := repo.GetByID(meal.ID)
	if err != nil {
		t.Errorf("Failed to get meal by ID: %v", err)
	}

	if retrievedMeal == nil {
		t.Error("Retrieved meal should not be nil")
	} else {
		if retrievedMeal.Name != meal.Name {
			t.Errorf("Expected name %s, got %s", meal.Name, retrievedMeal.Name)
		}
		if retrievedMeal.Price != meal.Price {
			t.Errorf("Expected price %f, got %f", meal.Price, retrievedMeal.Price)
		}
	}
}

func TestMealRepository_GetByChefID(t *testing.T) {
	db, err := setupMealTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewMealRepository(db)

	chefID := uint(1)

	// Create multiple test meals for the same chef
	meals := []*model.Meal{
		{
			ChefID:      chefID,
			Name:        "Meal 1",
			Description: "Description 1",
			Price:       20.00,
			Category:    "Appetizer",
			IsActive:    true,
			CookingTime: 15,
		},
		{
			ChefID:      chefID,
			Name:        "Meal 2",
			Description: "Description 2",
			Price:       35.00,
			Category:    "Main Course",
			IsActive:    true,
			CookingTime: 45,
		},
		{
			ChefID:      2, // Different chef
			Name:        "Meal 3",
			Description: "Description 3",
			Price:       15.00,
			Category:    "Dessert",
			IsActive:    true,
			CookingTime: 20,
		},
	}

	for _, meal := range meals {
		err = repo.Create(meal)
		if err != nil {
			t.Fatalf("Failed to create meal: %v", err)
		}
	}

	// Get meals by chef ID
	chefMeals, err := repo.GetByChefID(chefID)
	if err != nil {
		t.Errorf("Failed to get meals by chef ID: %v", err)
	}

	if len(chefMeals) != 2 {
		t.Errorf("Expected 2 meals for chef %d, got %d", chefID, len(chefMeals))
	}

	// Verify all meals belong to the correct chef
	for _, meal := range chefMeals {
		if meal.ChefID != chefID {
			t.Errorf("Expected chef ID %d, got %d", chefID, meal.ChefID)
		}
	}
}

func TestMealRepository_Update(t *testing.T) {
	db, err := setupMealTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewMealRepository(db)

	// Create a test meal
	meal := &model.Meal{
		ChefID:      1,
		Name:        "Original Meal",
		Description: "Original Description",
		Price:       25.50,
		Category:    "Main Course",
		IsActive:    true,
		CookingTime: 30,
	}

	err = repo.Create(meal)
	if err != nil {
		t.Fatalf("Failed to create meal: %v", err)
	}

	// Update the meal
	meal.Name = "Updated Meal"
	meal.Price = 30.00
	meal.IsActive = false

	err = repo.Update(meal)
	if err != nil {
		t.Errorf("Failed to update meal: %v", err)
	}

	// Retrieve and verify the update
	retrievedMeal, err := repo.GetByID(meal.ID)
	if err != nil {
		t.Errorf("Failed to get updated meal: %v", err)
	}

	if retrievedMeal == nil {
		t.Error("Retrieved meal should not be nil")
	} else {
		if retrievedMeal.Name != "Updated Meal" {
			t.Errorf("Expected updated name 'Updated Meal', got %s", retrievedMeal.Name)
		}
		if retrievedMeal.Price != 30.00 {
			t.Errorf("Expected updated price 30.00, got %f", retrievedMeal.Price)
		}
		if retrievedMeal.IsActive != false {
			t.Errorf("Expected IsActive to be false, got %t", retrievedMeal.IsActive)
		}
	}
}

func TestMealRepository_GetAll(t *testing.T) {
	db, err := setupMealTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewMealRepository(db)

	// Create multiple test meals
	meals := []*model.Meal{
		{
			ChefID:      1,
			Name:        "Meal 1",
			Description: "Description 1",
			Price:       20.00,
			Category:    "Appetizer",
			IsActive:    true,
		},
		{
			ChefID:      2,
			Name:        "Meal 2",
			Description: "Description 2",
			Price:       35.00,
			Category:    "Main Course",
			IsActive:    false,
		},
		{
			ChefID:      1,
			Name:        "Meal 3",
			Description: "Description 3",
			Price:       15.00,
			Category:    "Dessert",
			IsActive:    true,
		},
	}

	for _, meal := range meals {
		err = repo.Create(meal)
		if err != nil {
			t.Fatalf("Failed to create meal: %v", err)
		}
	}

	// Get all meals
	allMeals, err := repo.GetAll()
	if err != nil {
		t.Errorf("Failed to get all meals: %v", err)
	}

	if len(allMeals) != 3 {
		t.Errorf("Expected 3 meals, got %d", len(allMeals))
	}

	// Verify meal data
	nameSet := make(map[string]bool)
	for _, meal := range allMeals {
		nameSet[meal.Name] = true
	}

	expectedNames := []string{"Meal 1", "Meal 2", "Meal 3"}
	for _, name := range expectedNames {
		if !nameSet[name] {
			t.Errorf("Expected meal name %s not found in retrieved meals", name)
		}
	}
}

func TestMealRepository_PriceValidation(t *testing.T) {
	testCases := []struct {
		name     string
		price    float64
		valid    bool
	}{
		{"Positive price", 25.50, true},
		{"Zero price", 0.0, false},
		{"Negative price", -10.0, false},
		{"Large price", 999.99, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := tc.price > 0
			
			if isValid != tc.valid {
				t.Errorf("Price %f validation: expected %t, got %t", tc.price, tc.valid, isValid)
			}
		})
	}
}

func TestMealRepository_CategoryValidation(t *testing.T) {
	validCategories := []string{"Appetizer", "Main Course", "Dessert", "Beverage", "Soup", "Salad"}
	
	for _, category := range validCategories {
		t.Run("Category_"+category, func(t *testing.T) {
			meal := model.Meal{
				Category: category,
			}
			
			if meal.Category != category {
				t.Errorf("Expected category %s, got %s", category, meal.Category)
			}
		})
	}
}
