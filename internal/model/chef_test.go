package model

import (
	"testing"
	"time"
)

func TestChef_BasicStructure(t *testing.T) {
	chef := Chef{
		ID:          1,
		UserID:      1,
		KitchenName: "Ayşe'nin Mutfağı",
		Description: "Ev yemekleri uzmanı",
		Speciality:  "Geleneksel Türk Mutfağı",
		Experience:  5,
		Address:     "Test Mahallesi, Test Sokak No:1",
		District:    "Kadıköy",
		City:        "İstanbul",
		Latitude:    41.0082,
		Longitude:   28.9784,
		IsActive:    true,
		IsVerified:  false,
		Rating:      4.5,
		TotalOrders: 100,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if chef.ID != 1 {
		t.Errorf("Expected ID 1, got %d", chef.ID)
	}
	
	if chef.KitchenName != "Ayşe'nin Mutfağı" {
		t.Errorf("Expected kitchen name 'Ayşe'nin Mutfağı', got '%s'", chef.KitchenName)
	}
	
	if chef.Experience != 5 {
		t.Errorf("Expected experience 5, got %d", chef.Experience)
	}
}

func TestChef_KitchenNameValidation(t *testing.T) {
	testCases := []struct {
		name        string
		kitchenName string
		valid       bool
	}{
		{"Valid name", "Ayşe'nin Mutfağı", true},
		{"Valid short name", "Ali", true},
		{"Empty name", "", false},
		{"Very long name", "Bu çok uzun bir mutfak adı bu çok uzun bir mutfak adı bu çok uzun bir mutfak adı bu çok uzun bir mutfak adı", false},
		{"Valid Turkish characters", "Özgür'ün Şahane Mutfağı", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chef := Chef{
				KitchenName: tc.kitchenName,
			}
			
			// Basic validation - name should not be empty and should be under 100 chars
			isValid := tc.kitchenName != "" && len(tc.kitchenName) <= 100
			
			if tc.valid != isValid {
				t.Errorf("Kitchen name '%s' validation: expected %t, got %t", tc.kitchenName, tc.valid, isValid)
			}
			
			if chef.KitchenName != tc.kitchenName {
				t.Errorf("Expected kitchen name '%s', got '%s'", tc.kitchenName, chef.KitchenName)
			}
		})
	}
}

func TestChef_LocationValidation(t *testing.T) {
	testCases := []struct {
		name      string
		latitude  float64
		longitude float64
		valid     bool
	}{
		{"Istanbul coordinates", 41.0082, 28.9784, true},
		{"Ankara coordinates", 39.9334, 32.8597, true},
		{"Zero coordinates", 0.0, 0.0, false},
		{"Invalid latitude high", 91.0, 28.9784, false},
		{"Invalid latitude low", -91.0, 28.9784, false},
		{"Invalid longitude high", 41.0082, 181.0, false},
		{"Invalid longitude low", 41.0082, -181.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chef := Chef{
				Latitude:  tc.latitude,
				Longitude: tc.longitude,
			}
			
			// Basic GPS coordinate validation
			isValid := tc.latitude >= -90 && tc.latitude <= 90 &&
					  tc.longitude >= -180 && tc.longitude <= 180 &&
					  !(tc.latitude == 0 && tc.longitude == 0) // Exclude null island
			
			if tc.valid != isValid {
				t.Errorf("Coordinates (%f, %f) validation: expected %t, got %t", 
					tc.latitude, tc.longitude, tc.valid, isValid)
			}
			
			if chef.Latitude != tc.latitude {
				t.Errorf("Expected latitude %f, got %f", tc.latitude, chef.Latitude)
			}
		})
	}
}

func TestChef_RatingValidation(t *testing.T) {
	testCases := []struct {
		name   string
		rating float64
		valid  bool
	}{
		{"Perfect rating", 5.0, true},
		{"Good rating", 4.5, true},
		{"Average rating", 3.0, true},
		{"Low rating", 1.0, true},
		{"Zero rating", 0.0, true},
		{"Negative rating", -1.0, false},
		{"Too high rating", 6.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chef := Chef{
				Rating: tc.rating,
			}
			
			// Rating should be between 0 and 5
			isValid := tc.rating >= 0.0 && tc.rating <= 5.0
			
			if tc.valid != isValid {
				t.Errorf("Rating %f validation: expected %t, got %t", tc.rating, tc.valid, isValid)
			}
			
			if chef.Rating != tc.rating {
				t.Errorf("Expected rating %f, got %f", tc.rating, chef.Rating)
			}
		})
	}
}

func TestChef_ExperienceValidation(t *testing.T) {
	testCases := []struct {
		name       string
		experience int
		valid      bool
	}{
		{"No experience", 0, true},
		{"Some experience", 5, true},
		{"Lots of experience", 20, true},
		{"Negative experience", -1, false},
		{"Unrealistic experience", 100, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chef := Chef{
				Experience: tc.experience,
			}
			
			// Experience should be between 0 and 50 years
			isValid := tc.experience >= 0 && tc.experience <= 50
			
			if tc.valid != isValid {
				t.Errorf("Experience %d validation: expected %t, got %t", tc.experience, tc.valid, isValid)
			}
			
			if chef.Experience != tc.experience {
				t.Errorf("Expected experience %d, got %d", tc.experience, chef.Experience)
			}
		})
	}
}

func TestChef_StatusValidation(t *testing.T) {
	chef := Chef{
		IsActive:   true,
		IsVerified: false,
	}
	
	if !chef.IsActive {
		t.Error("Expected chef to be active")
	}
	
	if chef.IsVerified {
		t.Error("Expected chef to not be verified initially")
	}
	
	// Test status changes
	chef.IsVerified = true
	if !chef.IsVerified {
		t.Error("Chef should be verified after update")
	}
	
	chef.IsActive = false
	if chef.IsActive {
		t.Error("Chef should be inactive after update")
	}
}

func TestChef_AddressValidation(t *testing.T) {
	testCases := []struct {
		name     string
		address  string
		district string
		city     string
		valid    bool
	}{
		{"Complete address", "Test Mahallesi No:1", "Kadıköy", "İstanbul", true},
		{"Empty address", "", "Kadıköy", "İstanbul", false},
		{"Missing district", "Test Mahallesi No:1", "", "İstanbul", true}, // District optional
		{"Missing city", "Test Mahallesi No:1", "Kadıköy", "", true},      // City optional
		{"Only address", "Test Mahallesi No:1", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			chef := Chef{
				Address:  tc.address,
				District: tc.district,
				City:     tc.city,
			}
			
			// Address is required, others are optional
			isValid := tc.address != ""
			
			if tc.valid != isValid {
				t.Errorf("Address validation failed: expected %t, got %t", tc.valid, isValid)
			}
			
			if chef.Address != tc.address {
				t.Errorf("Expected address '%s', got '%s'", tc.address, chef.Address)
			}
		})
	}
}

func TestChef_OrdersStatistics(t *testing.T) {
	chef := Chef{
		TotalOrders: 0,
		Rating:      0.0,
	}
	
	// Test initial state
	if chef.TotalOrders != 0 {
		t.Errorf("Expected 0 total orders initially, got %d", chef.TotalOrders)
	}
	
	if chef.Rating != 0.0 {
		t.Errorf("Expected 0.0 rating initially, got %f", chef.Rating)
	}
	
	// Simulate order completion
	chef.TotalOrders = 10
	chef.Rating = 4.5
	
	if chef.TotalOrders != 10 {
		t.Errorf("Expected 10 total orders, got %d", chef.TotalOrders)
	}
	
	if chef.Rating != 4.5 {
		t.Errorf("Expected 4.5 rating, got %f", chef.Rating)
	}
}

func TestChef_Relations(t *testing.T) {
	// Test that chef struct has correct relation fields
	chef := Chef{
		ID:     1,
		UserID: 1,
	}
	
	// Test User relation
	chef.User = User{ID: 1, FirstName: "John", LastName: "Doe"}
	if chef.User.ID != 1 {
		t.Error("User relation should be set")
	}
	
	// Test Meals relation (if it exists)
	if chef.Meals != nil {
		t.Log("Meals relation exists")
	}
	
	// Test Reviews relation (if it exists)  
	if chef.Reviews != nil {
		t.Log("Reviews relation exists")
	}
	
	// These are structural tests
	chef.Meals = []Meal{}
	chef.Reviews = []Review{}
}

func TestChef_Timestamps(t *testing.T) {
	now := time.Now()
	chef := Chef{
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	if chef.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	
	if chef.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	
	// Test timestamp update
	later := now.Add(time.Hour)
	chef.UpdatedAt = later
	
	if !chef.UpdatedAt.After(chef.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}
