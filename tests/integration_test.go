package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce/internal/model"
	"github.com/gin-gonic/gin"
)

// IntegrationTestSuite - Integration test suite
type IntegrationTestSuite struct {
	router   *gin.Engine
	userID   uint
	adminID  uint
	token    string
	adminToken string
}

// SetupTestSuite - Setup test suite
func SetupTestSuite(t *testing.T) *IntegrationTestSuite {
	gin.SetMode(gin.TestMode)
	
	// In a real integration test, you would initialize the actual services
	// For now, we'll create a mock setup
	router := gin.New()
	
	// Setup routes (in real test, this would be from your actual router setup)
	setupTestRoutes(router)
	
	return &IntegrationTestSuite{
		router: router,
	}
}

// setupTestRoutes - Setup test routes
func setupTestRoutes(router *gin.Engine) {
	// Mock routes for testing
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", mockRegister)
			auth.POST("/login", mockLogin)
			auth.POST("/logout", mockLogout)
		}
		
		products := v1.Group("/products")
		{
			products.GET("", mockGetProducts)
			products.GET("/:id", mockGetProduct)
		}
		
		// Protected routes
		authorized := v1.Group("/")
		authorized.Use(mockAuthMiddleware())
		{
			authorized.GET("/profile", mockGetProfile)
			authorized.PUT("/profile", mockUpdateProfile)
			
			cart := authorized.Group("/cart")
			{
				cart.GET("", mockGetCart)
				cart.POST("/items", mockAddToCart)
				cart.DELETE("/items/:id", mockRemoveFromCart)
			}
			
			orders := authorized.Group("/orders")
			{
				orders.GET("", mockGetOrders)
				orders.POST("", mockCreateOrder)
				orders.GET("/:id", mockGetOrder)
			}
		}
		
		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(mockAdminMiddleware())
		{
			admin.GET("/products", mockAdminGetProducts)
			admin.POST("/products", mockAdminCreateProduct)
			admin.PUT("/products/:id", mockAdminUpdateProduct)
			admin.DELETE("/products/:id", mockAdminDeleteProduct)
			
			admin.GET("/orders", mockAdminGetOrders)
			admin.PUT("/orders/:id/status", mockAdminUpdateOrderStatus)
		}
	}
}

// Mock handlers
func mockRegister(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	response := model.AuthResponse{
		Token: "mock-jwt-token",
		User: model.User{
			ID:        1,
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Role:      "customer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    response,
	})
}

func mockLogin(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	response := model.AuthResponse{
		Token: "mock-jwt-token",
		User: model.User{
			ID:        1,
			Email:     req.Email,
			FirstName: "John",
			LastName:  "Doe",
			Role:      "customer",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

func mockLogout(c *gin.Context) {
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Logout successful",
	})
}

func mockGetProducts(c *gin.Context) {
	products := []model.ProductWithCategory{
		{
			ID:           1,
			Name:         "Test Product 1",
			Description:  "Test Description 1",
			Price:        29.99,
			CategoryName: "Electronics",
			Stock:        100,
			IsActive:     true,
		},
		{
			ID:           2,
			Name:         "Test Product 2",
			Description:  "Test Description 2",
			Price:        49.99,
			CategoryName: "Books",
			Stock:        50,
			IsActive:     true,
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Products retrieved successfully",
		Data: model.ProductResponse{
			Products: products,
			Total:    len(products),
		},
	})
}

func mockGetProduct(c *gin.Context) {
	product := model.ProductWithCategory{
		ID:           1,
		Name:         "Test Product",
		Description:  "Test Description",
		Price:        29.99,
		CategoryName: "Electronics",
		Stock:        100,
		IsActive:     true,
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Product retrieved successfully",
		Data:    product,
	})
}

func mockGetProfile(c *gin.Context) {
	profile := model.UserProfileResponse{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    profile,
	})
}

func mockUpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	profile := model.UserProfileResponse{
		ID:        1,
		Email:     "test@example.com",
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    profile,
	})
}

func mockGetCart(c *gin.Context) {
	cart := model.CartResponse{
		Items: []model.CartItemWithProduct{
			{
				ID:           1,
				ProductName:  "Test Product",
				ProductPrice: 29.99,
				ProductImage: "https://example.com/image.jpg",
				Quantity:     2,
				Subtotal:     59.98,
			},
		},
		Total: 59.98,
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Cart retrieved successfully",
		Data:    cart,
	})
}

func mockAddToCart(c *gin.Context) {
	var req model.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Item added to cart",
	})
}

func mockRemoveFromCart(c *gin.Context) {
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Item removed from cart",
	})
}

func mockGetOrders(c *gin.Context) {
	orders := []model.OrderWithItems{
		{
			ID:              1,
			Total:           99.97,
			Status:          "pending",
			PaymentMethod:   "credit_card",
			PaymentStatus:   "pending",
			ShippingAddress: "123 Main St",
			Items:           []model.OrderItemWithProduct{},
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Orders retrieved successfully",
		Data: model.OrderResponse{
			Orders: orders,
			Total:  len(orders),
		},
	})
}

func mockCreateOrder(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	order := model.OrderWithItems{
		ID:              1,
		Total:           99.97,
		Status:          "pending",
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   "pending",
		ShippingAddress: req.ShippingAddress,
		Notes:           req.Notes,
		Items:           []model.OrderItemWithProduct{},
	}
	
	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Order created successfully",
		Data:    order,
	})
}

func mockGetOrder(c *gin.Context) {
	order := model.OrderWithItems{
		ID:              1,
		Total:           99.97,
		Status:          "pending",
		PaymentMethod:   "credit_card",
		PaymentStatus:   "pending",
		ShippingAddress: "123 Main St",
		Items:           []model.OrderItemWithProduct{},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Order retrieved successfully",
		Data:    order,
	})
}

func mockAdminGetProducts(c *gin.Context) {
	products := []model.ProductWithCategory{
		{
			ID:           1,
			Name:         "Admin Product 1",
			Description:  "Admin Description 1",
			Price:        29.99,
			CategoryName: "Electronics",
			Stock:        100,
			IsActive:     true,
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Admin products retrieved successfully",
		Data:    products,
	})
}

func mockAdminCreateProduct(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	product := model.Product{
		ID:          1,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		Stock:       req.Stock,
		IsActive:    true,
	}
	
	c.JSON(http.StatusCreated, model.APIResponse{
		Success: true,
		Message: "Product created successfully",
		Data:    product,
	})
}

func mockAdminUpdateProduct(c *gin.Context) {
	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Product updated successfully",
	})
}

func mockAdminDeleteProduct(c *gin.Context) {
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
}

func mockAdminGetOrders(c *gin.Context) {
	orders := []model.OrderWithItems{
		{
			ID:              1,
			Total:           99.97,
			Status:          "pending",
			PaymentMethod:   "credit_card",
			PaymentStatus:   "pending",
			ShippingAddress: "123 Main St",
			Items:           []model.OrderItemWithProduct{},
		},
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Admin orders retrieved successfully",
		Data:    orders,
	})
}

func mockAdminUpdateOrderStatus(c *gin.Context) {
	var req model.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "Order status updated successfully",
	})
}

// Middleware mocks
func mockAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}
		
		// Mock token validation
		if authHeader != "Bearer mock-jwt-token" {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Message: "Invalid token",
			})
			c.Abort()
			return
		}
		
		c.Set("user_id", uint(1))
		c.Set("user_email", "test@example.com")
		c.Set("user_role", "customer")
		c.Next()
	}
}

func mockAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}
		
		// Mock admin token validation
		if authHeader != "Bearer mock-admin-token" {
			c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Message: "Admin access required",
			})
			c.Abort()
			return
		}
		
		c.Set("user_id", uint(1))
		c.Set("user_email", "admin@example.com")
		c.Set("user_role", "admin")
		c.Next()
	}
}

// Integration tests
func TestIntegration_AuthFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	
	// Test user registration
	t.Run("Register User", func(t *testing.T) {
		registerRequest := model.RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
		}
		
		body, _ := json.Marshal(registerRequest)
		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		
		var response model.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response.Success {
			t.Error("Expected successful registration")
		}
	})
	
	// Test user login
	t.Run("Login User", func(t *testing.T) {
		loginRequest := model.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		
		body, _ := json.Marshal(loginRequest)
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response model.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response.Success {
			t.Error("Expected successful login")
		}
	})
}

func TestIntegration_ProductFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	
	// Test get products
	t.Run("Get Products", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/products", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response model.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response.Success {
			t.Error("Expected successful product retrieval")
		}
	})
	
	// Test get single product
	t.Run("Get Single Product", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/products/1", nil)
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		
		var response model.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		if !response.Success {
			t.Error("Expected successful product retrieval")
		}
	})
}

func TestIntegration_CartFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	
	// Test authenticated cart operations
	t.Run("Get Cart", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/cart", nil)
		req.Header.Set("Authorization", "Bearer mock-jwt-token")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
	
	// Test add to cart
	t.Run("Add to Cart", func(t *testing.T) {
		addRequest := model.AddToCartRequest{
			ProductID: 1,
			Quantity:  2,
		}
		
		body, _ := json.Marshal(addRequest)
		req, _ := http.NewRequest("POST", "/api/v1/cart/items", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-jwt-token")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestIntegration_OrderFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	
	// Test create order
	t.Run("Create Order", func(t *testing.T) {
		orderRequest := model.CreateOrderRequest{
			PaymentMethod:   "credit_card",
			ShippingAddress: "123 Main St, City, State 12345",
			Notes:           "Please deliver after 6 PM",
		}
		
		body, _ := json.Marshal(orderRequest)
		req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-jwt-token")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
	
	// Test get orders
	t.Run("Get Orders", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/orders", nil)
		req.Header.Set("Authorization", "Bearer mock-jwt-token")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestIntegration_AdminFlow(t *testing.T) {
	suite := SetupTestSuite(t)
	
	// Test admin product creation
	t.Run("Admin Create Product", func(t *testing.T) {
		productRequest := model.CreateProductRequest{
			Name:        "New Product",
			Description: "New Product Description",
			Price:       99.99,
			CategoryID:  1,
			Stock:       100,
		}
		
		body, _ := json.Marshal(productRequest)
		req, _ := http.NewRequest("POST", "/api/v1/admin/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-admin-token")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
	})
	
	// Test admin access control
	t.Run("Admin Access Control", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/admin/products", nil)
		req.Header.Set("Authorization", "Bearer mock-jwt-token") // Non-admin token
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

func TestIntegration_AuthenticationRequired(t *testing.T) {
	suite := SetupTestSuite(t)
	
	protectedEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/profile"},
		{"PUT", "/api/v1/profile"},
		{"GET", "/api/v1/cart"},
		{"POST", "/api/v1/cart/items"},
		{"GET", "/api/v1/orders"},
		{"POST", "/api/v1/orders"},
	}
	
	for _, endpoint := range protectedEndpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req, _ := http.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status %d for %s %s, got %d", 
					http.StatusUnauthorized, endpoint.method, endpoint.path, w.Code)
			}
		})
	}
}
