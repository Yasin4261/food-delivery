package domain

// MenuItemRepository defines the interface for menu item data operations
type MenuItemRepository interface {
	// Create creates a new menu item
	Create(item *MenuItem) error
	
	// FindByID finds a menu item by ID
	FindByID(id int) (*MenuItem, error)
	
	// Update updates menu item information
	Update(item *MenuItem) error
	
	// Delete deletes a menu item
	Delete(id int) error
	
	// FindByMenuID finds menu items by menu ID
	FindByMenuID(menuID int) ([]*MenuItem, error)
	
	// FindByChefID finds menu items by chef ID
	FindByChefID(chefID int, offset, limit int) ([]*MenuItem, error)
	
	// FindByCategory finds menu items by category
	FindByCategory(category string, offset, limit int) ([]*MenuItem, error)
	
	// FindByCuisine finds menu items by cuisine
	FindByCuisine(cuisine string, offset, limit int) ([]*MenuItem, error)
	
	// FindAvailable finds available menu items
	FindAvailable(offset, limit int) ([]*MenuItem, error)
	
	// FindFeatured finds featured menu items
	FindFeatured(offset, limit int) ([]*MenuItem, error)
	
	// FindByPriceRange finds menu items within price range
	FindByPriceRange(minPrice, maxPrice float64, offset, limit int) ([]*MenuItem, error)
	
	// FindByDiet finds menu items by dietary preferences
	FindByDiet(vegetarian, vegan, glutenFree, halal bool, offset, limit int) ([]*MenuItem, error)
	
	// UpdateStock updates available quantity
	UpdateStock(id int, quantity int) error
	
	// UpdateRating updates menu item rating
	UpdateRating(id int, rating float64) error
	
	// IncrementViews increments view count
	IncrementViews(id int) error
	
	// IncrementOrders increments order count
	IncrementOrders(id int) error
	
	// FindPopular finds most ordered items
	FindPopular(limit int) ([]*MenuItem, error)
}
