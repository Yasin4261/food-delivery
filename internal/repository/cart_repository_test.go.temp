package repository

import (
	"testing"
	"ecommerce/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Simple integration tests for CartRepository using SQLite in-memory database

func setupCartTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&model.Cart{}, &model.CartItem{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func TestCartRepository_CreateCart(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Errorf("Failed to create cart: %v", err)
	}

	if cart.ID == 0 {
		t.Error("Cart ID should be assigned after creation")
	}
}

func TestCartRepository_GetCartByUserID(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Retrieve the cart by user ID
	retrievedCart, err := repo.GetCartByUserID(cart.UserID)
	if err != nil {
		t.Errorf("Failed to get cart by user ID: %v", err)
	}

	if retrievedCart == nil {
		t.Error("Retrieved cart should not be nil")
	} else {
		if retrievedCart.UserID != cart.UserID {
			t.Errorf("Expected user ID %d, got %d", cart.UserID, retrievedCart.UserID)
		}
		if retrievedCart.SessionID != cart.SessionID {
			t.Errorf("Expected session ID %s, got %s", cart.SessionID, retrievedCart.SessionID)
		}
	}
}

func TestCartRepository_AddCartItem(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add cart item
	cartItem := &model.CartItem{
		CartID:   cart.ID,
		MealID:   1,
		Quantity: 2,
		Price:    25.50,
	}

	err = repo.AddCartItem(cartItem)
	if err != nil {
		t.Errorf("Failed to add cart item: %v", err)
	}

	if cartItem.ID == 0 {
		t.Error("Cart item ID should be assigned after creation")
	}
}

func TestCartRepository_GetCartItems(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add multiple cart items
	cartItems := []*model.CartItem{
		{
			CartID:   cart.ID,
			MealID:   1,
			Quantity: 2,
			Price:    25.50,
		},
		{
			CartID:   cart.ID,
			MealID:   2,
			Quantity: 1,
			Price:    15.00,
		},
		{
			CartID:   cart.ID,
			MealID:   3,
			Quantity: 3,
			Price:    18.75,
		},
	}

	for _, item := range cartItems {
		err = repo.AddCartItem(item)
		if err != nil {
			t.Fatalf("Failed to add cart item: %v", err)
		}
	}

	// Get cart items
	retrievedItems, err := repo.GetCartItems(cart.ID)
	if err != nil {
		t.Errorf("Failed to get cart items: %v", err)
	}

	if len(retrievedItems) != 3 {
		t.Errorf("Expected 3 cart items, got %d", len(retrievedItems))
	}

	// Verify items belong to the correct cart
	for _, item := range retrievedItems {
		if item.CartID != cart.ID {
			t.Errorf("Expected cart ID %d, got %d", cart.ID, item.CartID)
		}
	}
}

func TestCartRepository_UpdateCartItem(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add cart item
	cartItem := &model.CartItem{
		CartID:   cart.ID,
		MealID:   1,
		Quantity: 2,
		Price:    25.50,
	}

	err = repo.AddCartItem(cartItem)
	if err != nil {
		t.Fatalf("Failed to add cart item: %v", err)
	}

	// Update cart item
	cartItem.Quantity = 5
	cartItem.Price = 30.00

	err = repo.UpdateCartItem(cartItem)
	if err != nil {
		t.Errorf("Failed to update cart item: %v", err)
	}

	// Retrieve and verify the update
	items, err := repo.GetCartItems(cart.ID)
	if err != nil {
		t.Errorf("Failed to get cart items after update: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 cart item, got %d", len(items))
	}

	updatedItem := items[0]
	if updatedItem.Quantity != 5 {
		t.Errorf("Expected updated quantity 5, got %d", updatedItem.Quantity)
	}
	if updatedItem.Price != 30.00 {
		t.Errorf("Expected updated price 30.00, got %f", updatedItem.Price)
	}
}

func TestCartRepository_RemoveCartItem(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add cart item
	cartItem := &model.CartItem{
		CartID:   cart.ID,
		MealID:   1,
		Quantity: 2,
		Price:    25.50,
	}

	err = repo.AddCartItem(cartItem)
	if err != nil {
		t.Fatalf("Failed to add cart item: %v", err)
	}

	// Remove cart item
	err = repo.RemoveCartItem(cartItem.ID)
	if err != nil {
		t.Errorf("Failed to remove cart item: %v", err)
	}

	// Verify item is removed
	items, err := repo.GetCartItems(cart.ID)
	if err != nil {
		t.Errorf("Failed to get cart items after removal: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 cart items after removal, got %d", len(items))
	}
}

func TestCartRepository_ClearCart(t *testing.T) {
	db, err := setupCartTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	repo := NewCartRepository(db)

	// Create a test cart
	cart := &model.Cart{
		UserID:    1,
		SessionID: "test-session-123",
		IsActive:  true,
	}

	err = repo.CreateCart(cart)
	if err != nil {
		t.Fatalf("Failed to create cart: %v", err)
	}

	// Add multiple cart items
	cartItems := []*model.CartItem{
		{CartID: cart.ID, MealID: 1, Quantity: 2, Price: 25.50},
		{CartID: cart.ID, MealID: 2, Quantity: 1, Price: 15.00},
		{CartID: cart.ID, MealID: 3, Quantity: 3, Price: 18.75},
	}

	for _, item := range cartItems {
		err = repo.AddCartItem(item)
		if err != nil {
			t.Fatalf("Failed to add cart item: %v", err)
		}
	}

	// Clear cart
	err = repo.ClearCart(cart.ID)
	if err != nil {
		t.Errorf("Failed to clear cart: %v", err)
	}

	// Verify cart is cleared
	items, err := repo.GetCartItems(cart.ID)
	if err != nil {
		t.Errorf("Failed to get cart items after clearing: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 cart items after clearing, got %d", len(items))
	}
}

func TestCartRepository_CartCalculations(t *testing.T) {
	// Test cart calculation logic
	cartItems := []model.CartItem{
		{MealID: 1, Quantity: 2, Price: 25.50}, // Total: 51.00
		{MealID: 2, Quantity: 1, Price: 15.00}, // Total: 15.00
		{MealID: 3, Quantity: 3, Price: 10.00}, // Total: 30.00
	}

	totalAmount := 0.0
	totalItems := 0

	for _, item := range cartItems {
		totalAmount += float64(item.Quantity) * item.Price
		totalItems += item.Quantity
	}

	expectedTotal := 96.00 // 51.00 + 15.00 + 30.00
	if totalAmount != expectedTotal {
		t.Errorf("Expected total amount %f, got %f", expectedTotal, totalAmount)
	}

	expectedItems := 6 // 2 + 1 + 3
	if totalItems != expectedItems {
		t.Errorf("Expected total items %d, got %d", expectedItems, totalItems)
	}
}
