package repository

import (
	"testing"
	"ecommerce/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Simple integration tests for UserRepository using SQLite in-memory database

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestUserRepository_Create(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	user := &model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "customer",
		IsActive: true,
	}

	err = repo.Create(user)
	if err != nil {
		t.Errorf("Failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("User ID should be assigned after creation")
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Create a test user
	user := &model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "customer",
		IsActive: true,
	}

	err = repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve the user
	retrievedUser, err := repo.GetByID(user.ID)
	if err != nil {
		t.Errorf("Failed to get user by ID: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Retrieved user should not be nil")
	} else {
		if retrievedUser.Name != user.Name {
			t.Errorf("Expected name %s, got %s", user.Name, retrievedUser.Name)
		}
		if retrievedUser.Email != user.Email {
			t.Errorf("Expected email %s, got %s", user.Email, retrievedUser.Email)
		}
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Create a test user
	user := &model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "customer",
		IsActive: true,
	}

	err = repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Retrieve the user by email
	retrievedUser, err := repo.GetByEmail(user.Email)
	if err != nil {
		t.Errorf("Failed to get user by email: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Retrieved user should not be nil")
	} else {
		if retrievedUser.Name != user.Name {
			t.Errorf("Expected name %s, got %s", user.Name, retrievedUser.Name)
		}
		if retrievedUser.ID != user.ID {
			t.Errorf("Expected ID %d, got %d", user.ID, retrievedUser.ID)
		}
	}
}

func TestUserRepository_Update(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Create a test user
	user := &model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "customer",
		IsActive: true,
	}

	err = repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update the user
	user.Name = "Updated User"
	user.IsActive = false

	err = repo.Update(user)
	if err != nil {
		t.Errorf("Failed to update user: %v", err)
	}

	// Retrieve and verify the update
	retrievedUser, err := repo.GetByID(user.ID)
	if err != nil {
		t.Errorf("Failed to get updated user: %v", err)
	}

	if retrievedUser == nil {
		t.Error("Retrieved user should not be nil")
	} else {
		if retrievedUser.Name != "Updated User" {
			t.Errorf("Expected updated name 'Updated User', got %s", retrievedUser.Name)
		}
		if retrievedUser.IsActive != false {
			t.Errorf("Expected IsActive to be false, got %t", retrievedUser.IsActive)
		}
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Create a test user
	user := &model.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "customer",
		IsActive: true,
	}

	err = repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	userID := user.ID

	// Delete the user
	err = repo.Delete(userID)
	if err != nil {
		t.Errorf("Failed to delete user: %v", err)
	}

	// Try to retrieve the deleted user
	retrievedUser, err := repo.GetByID(userID)
	if err == nil && retrievedUser != nil {
		t.Error("Deleted user should not be found")
	}
}

func TestUserRepository_GetAll(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Create multiple test users
	users := []*model.User{
		{
			Name:     "User 1",
			Email:    "user1@example.com",
			Password: "hashedpassword1",
			Role:     "customer",
			IsActive: true,
		},
		{
			Name:     "User 2",
			Email:    "user2@example.com",
			Password: "hashedpassword2",
			Role:     "chef",
			IsActive: true,
		},
		{
			Name:     "User 3",
			Email:    "user3@example.com",
			Password: "hashedpassword3",
			Role:     "admin",
			IsActive: false,
		},
	}

	for _, user := range users {
		err = repo.Create(user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Get all users
	allUsers, err := repo.GetAll()
	if err != nil {
		t.Errorf("Failed to get all users: %v", err)
	}

	if len(allUsers) != 3 {
		t.Errorf("Expected 3 users, got %d", len(allUsers))
	}

	// Verify users are correctly retrieved
	emailSet := make(map[string]bool)
	for _, user := range allUsers {
		emailSet[user.Email] = true
	}

	expectedEmails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for _, email := range expectedEmails {
		if !emailSet[email] {
			t.Errorf("Expected email %s not found in retrieved users", email)
		}
	}
}

func TestUserRepository_NonExistentUser(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewUserRepository(db)

	// Try to get a non-existent user
	user, err := repo.GetByID(9999)
	if err == nil && user != nil {
		t.Error("Non-existent user should not be found")
	}

	// Try to get by non-existent email
	user, err = repo.GetByEmail("nonexistent@example.com")
	if err == nil && user != nil {
		t.Error("Non-existent user should not be found")
	}
}
