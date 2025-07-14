package repository

import (
	"testing"
	"time"

	"ecommerce/internal/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB - Test veritabanı kurulumu
func SetupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(
		&model.User{},
		&model.Chef{},
		&model.Meal{},
		&model.Order{},
		&model.SubOrder{},
		&model.OrderItem{},
		&model.Review{},
		&model.Cart{},
	)

	return db
}

// SeedTestData - Test verilerini ekle
func SeedTestData(db *gorm.DB) {
	// Users
	users := []model.User{
		{ID: 1, Name: "Test User 1", Email: "user1@test.com", Phone: "+905551111111"},
		{ID: 2, Name: "Test User 2", Email: "user2@test.com", Phone: "+905552222222"},
	}
	db.Create(&users)

	// Chefs
	chefs := []model.Chef{
		{ID: 1, UserID: 1, RestaurantName: "Chef1 Restaurant", Cuisine: "Turkish", IsActive: true},
		{ID: 2, UserID: 2, RestaurantName: "Chef2 Restaurant", Cuisine: "Italian", IsActive: true},
		{ID: 3, UserID: 3, RestaurantName: "Chef3 Restaurant", Cuisine: "Asian", IsActive: true},
	}
	db.Create(&chefs)

	// Meals
	meals := []model.Meal{
		{ID: 1, ChefID: 1, Name: "Turkish Kebab", Price: 25.50, AvailableQuantity: 10, IsActive: true},
		{ID: 2, ChefID: 1, Name: "Turkish Pilaf", Price: 15.00, AvailableQuantity: 15, IsActive: true},
		{ID: 3, ChefID: 2, Name: "Italian Pizza", Price: 30.00, AvailableQuantity: 8, IsActive: true},
		{ID: 4, ChefID: 2, Name: "Italian Pasta", Price: 22.50, AvailableQuantity: 12, IsActive: true},
		{ID: 5, ChefID: 3, Name: "Asian Noodles", Price: 20.00, AvailableQuantity: 20, IsActive: true},
	}
	db.Create(&meals)
}

// TestOrderRepository_CreateSingleChefOrder - Tek şef siparişi oluşturma testi
func TestOrderRepository_CreateSingleChefOrder(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	order := &model.Order{
		UserID:          1,
		OrderNumber:     "ORD-20250714-001",
		Total:           51.00,
		Currency:        "TRY",
		Status:          "pending",
		DeliveryType:    "delivery",
		Address:         "123 Test St",
		DeliveryAddress: "Apt 5",
		PaymentMethod:   "credit_card",
		PaymentStatus:   "pending",
		ChefCount:       1,
		Items: []model.OrderItem{
			{
				MealID:   1,
				ChefID:   1,
				Quantity: 2,
				Price:    25.50,
				Subtotal: 51.00,
			},
		},
	}

	err := orderRepo.Create(order)
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)
	assert.NotZero(t, order.CreatedAt)

	// Verify order was created
	createdOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.OrderNumber, createdOrder.OrderNumber)
	assert.Equal(t, order.ChefCount, createdOrder.ChefCount)
	assert.Equal(t, 1, len(createdOrder.Items))
}

// TestOrderRepository_CreateMultiVendorOrder - Multi-vendor sipariş oluşturma testi
func TestOrderRepository_CreateMultiVendorOrder(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	order := &model.Order{
		UserID:          1,
		OrderNumber:     "ORD-20250714-002",
		Total:           127.50,
		Currency:        "TRY",
		Status:          "pending",
		DeliveryType:    "delivery",
		Address:         "456 Test Ave",
		DeliveryAddress: "Suite 10",
		PaymentMethod:   "cash",
		PaymentStatus:   "pending",
		ChefCount:       3,
		SubOrders: []model.SubOrder{
			{
				ChefID:          1,
				ChefOrderNumber: "ORD-20250714-002-CHEF1",
				Subtotal:        40.50,
				DeliveryFee:     5.00,
				ServiceFee:      2.00,
				Total:           47.50,
				Status:          "pending",
				EstimatedTime:   30,
			},
			{
				ChefID:          2,
				ChefOrderNumber: "ORD-20250714-002-CHEF2",
				Subtotal:        52.50,
				DeliveryFee:     5.00,
				ServiceFee:      2.50,
				Total:           60.00,
				Status:          "pending",
				EstimatedTime:   45,
			},
			{
				ChefID:          3,
				ChefOrderNumber: "ORD-20250714-002-CHEF3",
				Subtotal:        18.00,
				DeliveryFee:     0.00,
				ServiceFee:      2.00,
				Total:           20.00,
				Status:          "pending",
				EstimatedTime:   25,
			},
		},
		Items: []model.OrderItem{
			{
				MealID:              1,
				ChefID:              1,
				Quantity:            1,
				Price:               25.50,
				Subtotal:            25.50,
				SpecialInstructions: "Medium spicy",
			},
			{
				MealID:   2,
				ChefID:   1,
				Quantity: 1,
				Price:    15.00,
				Subtotal: 15.00,
			},
			{
				MealID:              3,
				ChefID:              2,
				Quantity:            1,
				Price:               30.00,
				Subtotal:            30.00,
				SpecialInstructions: "Extra cheese",
			},
			{
				MealID:   4,
				ChefID:   2,
				Quantity: 1,
				Price:    22.50,
				Subtotal: 22.50,
			},
			{
				MealID:              5,
				ChefID:              3,
				Quantity: 1,
				Price:               20.00,
				Subtotal:            20.00,
				SpecialInstructions: "No seafood",
			},
		},
	}

	// Set sub_order_id for items
	order.Items[0].SubOrderID = 1
	order.Items[1].SubOrderID = 1
	order.Items[2].SubOrderID = 2
	order.Items[3].SubOrderID = 2
	order.Items[4].SubOrderID = 3

	err := orderRepo.Create(order)
	assert.NoError(t, err)
	assert.NotZero(t, order.ID)

	// Verify order was created with all sub-orders and items
	createdOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, order.OrderNumber, createdOrder.OrderNumber)
	assert.Equal(t, order.ChefCount, createdOrder.ChefCount)
	assert.Equal(t, 3, len(createdOrder.SubOrders))
	assert.Equal(t, 5, len(createdOrder.Items))

	// Verify sub-order details
	for i, subOrder := range createdOrder.SubOrders {
		assert.Equal(t, order.SubOrders[i].ChefID, subOrder.ChefID)
		assert.Equal(t, order.SubOrders[i].Total, subOrder.Total)
		assert.Equal(t, order.SubOrders[i].EstimatedTime, subOrder.EstimatedTime)
	}

	// Verify items are correctly linked to sub-orders
	chef1Items := 0
	chef2Items := 0
	chef3Items := 0
	for _, item := range createdOrder.Items {
		switch item.ChefID {
		case 1:
			chef1Items++
		case 2:
			chef2Items++
		case 3:
			chef3Items++
		}
	}
	assert.Equal(t, 2, chef1Items)
	assert.Equal(t, 2, chef2Items)
	assert.Equal(t, 1, chef3Items)
}

// TestOrderRepository_GetOrdersByUser - Kullanıcı siparişlerini getirme testi
func TestOrderRepository_GetOrdersByUser(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create test orders for user 1
	orders := []*model.Order{
		{
			UserID:      1,
			OrderNumber: "ORD-20250714-001",
			Total:       50.00,
			Status:      "delivered",
			ChefCount:   1,
			CreatedAt:   time.Now().AddDate(0, 0, -3),
		},
		{
			UserID:      1,
			OrderNumber: "ORD-20250714-002",
			Total:       75.50,
			Status:      "preparing",
			ChefCount:   2,
			CreatedAt:   time.Now().AddDate(0, 0, -1),
		},
		{
			UserID:      1,
			OrderNumber: "ORD-20250714-003",
			Total:       100.00,
			Status:      "pending",
			ChefCount:   1,
			CreatedAt:   time.Now(),
		},
		{
			UserID:      2,
			OrderNumber: "ORD-20250714-004",
			Total:       25.00,
			Status:      "confirmed",
			ChefCount:   1,
			CreatedAt:   time.Now(),
		},
	}

	for _, order := range orders {
		err := orderRepo.Create(order)
		assert.NoError(t, err)
	}

	// Get orders for user 1
	userOrders, err := orderRepo.GetByUserID(1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(userOrders))

	// Verify orders are sorted by creation date (newest first)
	for i := 1; i < len(userOrders); i++ {
		assert.True(t, userOrders[i-1].CreatedAt.After(userOrders[i].CreatedAt) || 
					userOrders[i-1].CreatedAt.Equal(userOrders[i].CreatedAt))
	}

	// Get orders for user 2
	user2Orders, err := orderRepo.GetByUserID(2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(user2Orders))

	// Get orders for non-existent user
	noOrders, err := orderRepo.GetByUserID(999)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(noOrders))
}

// TestOrderRepository_UpdateOrderStatus - Sipariş durumu güncelleme testi
func TestOrderRepository_UpdateOrderStatus(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create a test order
	order := &model.Order{
		UserID:      1,
		OrderNumber: "ORD-20250714-001",
		Total:       50.00,
		Status:      "pending",
		ChefCount:   1,
	}

	err := orderRepo.Create(order)
	assert.NoError(t, err)

	// Update status
	order.Status = "confirmed"
	err = orderRepo.Update(order)
	assert.NoError(t, err)

	// Verify status was updated
	updatedOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, "confirmed", updatedOrder.Status)
	assert.True(t, updatedOrder.UpdatedAt.After(updatedOrder.CreatedAt))
}

// TestOrderRepository_UpdateSubOrderStatus - Alt sipariş durumu güncelleme testi
func TestOrderRepository_UpdateSubOrderStatus(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create a test order with sub-orders
	order := &model.Order{
		UserID:      1,
		OrderNumber: "ORD-20250714-001",
		Total:       100.00,
		Status:      "pending",
		ChefCount:   2,
		SubOrders: []model.SubOrder{
			{
				ChefID:          1,
				ChefOrderNumber: "ORD-20250714-001-CHEF1",
				Total:           50.00,
				Status:          "pending",
				EstimatedTime:   30,
			},
			{
				ChefID:          2,
				ChefOrderNumber: "ORD-20250714-001-CHEF2",
				Total:           50.00,
				Status:          "pending",
				EstimatedTime:   45,
			},
		},
	}

	err := orderRepo.Create(order)
	assert.NoError(t, err)

	// Get created sub-orders
	createdOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(createdOrder.SubOrders))

	// Update first sub-order status
	subOrder1 := &createdOrder.SubOrders[0]
	subOrder1.Status = "confirmed"
	
	err = orderRepo.UpdateSubOrder(subOrder1)
	assert.NoError(t, err)

	// Verify sub-order status was updated
	updatedOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	
	var confirmedSubOrder *model.SubOrder
	for _, so := range updatedOrder.SubOrders {
		if so.ID == subOrder1.ID {
			confirmedSubOrder = &so
			break
		}
	}
	
	assert.NotNil(t, confirmedSubOrder)
	assert.Equal(t, "confirmed", confirmedSubOrder.Status)
}

// TestOrderRepository_DeleteOrder - Sipariş silme testi
func TestOrderRepository_DeleteOrder(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create a test order
	order := &model.Order{
		UserID:      1,
		OrderNumber: "ORD-20250714-001",
		Total:       50.00,
		Status:      "pending",
		ChefCount:   1,
		Items: []model.OrderItem{
			{
				MealID:   1,
				ChefID:   1,
				Quantity: 2,
				Price:    25.00,
				Subtotal: 50.00,
			},
		},
	}

	err := orderRepo.Create(order)
	assert.NoError(t, err)

	// Verify order exists
	createdOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.NotNil(t, createdOrder)

	// Delete order
	err = orderRepo.Delete(order.ID)
	assert.NoError(t, err)

	// Verify order is deleted
	deletedOrder, err := orderRepo.GetByID(order.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedOrder)
}

// TestOrderRepository_GetOrdersByStatus - Duruma göre sipariş getirme testi
func TestOrderRepository_GetOrdersByStatus(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create test orders with different statuses
	orders := []*model.Order{
		{UserID: 1, OrderNumber: "ORD-001", Total: 50.00, Status: "pending", ChefCount: 1},
		{UserID: 1, OrderNumber: "ORD-002", Total: 75.00, Status: "confirmed", ChefCount: 1},
		{UserID: 1, OrderNumber: "ORD-003", Total: 100.00, Status: "pending", ChefCount: 2},
		{UserID: 2, OrderNumber: "ORD-004", Total: 25.00, Status: "delivered", ChefCount: 1},
		{UserID: 2, OrderNumber: "ORD-005", Total: 60.00, Status: "pending", ChefCount: 1},
	}

	for _, order := range orders {
		err := orderRepo.Create(order)
		assert.NoError(t, err)
	}

	// Get pending orders
	pendingOrders, err := orderRepo.GetByStatus("pending")
	assert.NoError(t, err)
	assert.Equal(t, 3, len(pendingOrders))

	// Get confirmed orders
	confirmedOrders, err := orderRepo.GetByStatus("confirmed")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(confirmedOrders))

	// Get delivered orders
	deliveredOrders, err := orderRepo.GetByStatus("delivered")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(deliveredOrders))

	// Get orders with non-existent status
	noOrders, err := orderRepo.GetByStatus("non_existent")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(noOrders))
}

// TestOrderRepository_GetOrdersByDateRange - Tarih aralığına göre sipariş getirme
func TestOrderRepository_GetOrdersByDateRange(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	now := time.Now()
	
	// Create orders with different dates
	orders := []*model.Order{
		{
			UserID: 1, OrderNumber: "ORD-001", Total: 50.00, Status: "delivered", ChefCount: 1,
			CreatedAt: now.AddDate(0, 0, -5), // 5 days ago
		},
		{
			UserID: 1, OrderNumber: "ORD-002", Total: 75.00, Status: "delivered", ChefCount: 1,
			CreatedAt: now.AddDate(0, 0, -3), // 3 days ago
		},
		{
			UserID: 1, OrderNumber: "ORD-003", Total: 100.00, Status: "pending", ChefCount: 1,
			CreatedAt: now.AddDate(0, 0, -1), // 1 day ago
		},
		{
			UserID: 2, OrderNumber: "ORD-004", Total: 25.00, Status: "confirmed", ChefCount: 1,
			CreatedAt: now, // today
		},
	}

	for _, order := range orders {
		err := orderRepo.Create(order)
		assert.NoError(t, err)
	}

	// Get orders from last 7 days
	startDate := now.AddDate(0, 0, -7)
	endDate := now.AddDate(0, 0, 1)
	
	allOrders, err := orderRepo.GetByDateRange(startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(allOrders))

	// Get orders from last 2 days
	startDate = now.AddDate(0, 0, -2)
	recentOrders, err := orderRepo.GetByDateRange(startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(recentOrders))

	// Get orders from future (should be empty)
	startDate = now.AddDate(0, 0, 1)
	endDate = now.AddDate(0, 0, 7)
	futureOrders, err := orderRepo.GetByDateRange(startDate, endDate)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(futureOrders))
}

// TestOrderRepository_GetMultiVendorOrderStats - Multi-vendor sipariş istatistikleri
func TestOrderRepository_GetMultiVendorOrderStats(t *testing.T) {
	db := SetupTestDB()
	SeedTestData(db)
	
	orderRepo := NewOrderRepository(db)

	// Create mix of single and multi-vendor orders
	orders := []*model.Order{
		{UserID: 1, OrderNumber: "ORD-001", Total: 50.00, Status: "delivered", ChefCount: 1},
		{UserID: 1, OrderNumber: "ORD-002", Total: 75.00, Status: "delivered", ChefCount: 2},
		{UserID: 1, OrderNumber: "ORD-003", Total: 100.00, Status: "pending", ChefCount: 3},
		{UserID: 2, OrderNumber: "ORD-004", Total: 25.00, Status: "confirmed", ChefCount: 1},
		{UserID: 2, OrderNumber: "ORD-005", Total: 80.00, Status: "delivered", ChefCount: 2},
	}

	for _, order := range orders {
		err := orderRepo.Create(order)
		assert.NoError(t, err)
	}

	// Get all orders
	allOrders, err := orderRepo.GetAll()
	assert.NoError(t, err)
	assert.Equal(t, 5, len(allOrders))

	// Calculate stats
	singleVendorCount := 0
	multiVendorCount := 0
	
	for _, order := range allOrders {
		if order.ChefCount == 1 {
			singleVendorCount++
		} else {
			multiVendorCount++
		}
	}

	assert.Equal(t, 2, singleVendorCount)
	assert.Equal(t, 3, multiVendorCount)

	// Calculate multi-vendor ratio
	multiVendorRatio := float64(multiVendorCount) / float64(len(allOrders))
	assert.Equal(t, 0.6, multiVendorRatio) // 3/5 = 0.6
}
