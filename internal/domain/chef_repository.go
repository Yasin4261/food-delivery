package domain

// ChefRepository defines the interface for chef data operations
type ChefRepository interface {
	// Create creates a new chef profile
	Create(chef *Chef) error
	
	// FindByID finds a chef by ID
	FindByID(id int) (*Chef, error)
	
	// FindByUserID finds a chef by user ID
	FindByUserID(userID int) (*Chef, error)
	
	// Update updates chef information
	Update(chef *Chef) error
	
	// Delete deletes a chef profile (soft delete)
	Delete(id int) error
	
	// List returns paginated chefs
	List(offset, limit int) ([]*Chef, error)
	
	// FindByCity finds chefs in a city
	FindByCity(city string, offset, limit int) ([]*Chef, error)
	
	// FindNearby finds chefs within delivery radius from a location
	FindNearby(lat, lng float64, limit int) ([]*Chef, error)
	
	// FindByRating finds chefs with rating >= minRating
	FindByRating(minRating float64, offset, limit int) ([]*Chef, error)
	
	// FindVerified finds all verified chefs
	FindVerified(offset, limit int) ([]*Chef, error)
	
	// FindAcceptingOrders finds chefs currently accepting orders
	FindAcceptingOrders(offset, limit int) ([]*Chef, error)
	
	// UpdateRating updates chef rating and review count
	UpdateRating(id int, rating float64) error
	
	// IncrementOrders increments total orders count
	IncrementOrders(id int) error
}
