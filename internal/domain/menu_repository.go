package domain

// MenuRepository defines the interface for menu data operations
type MenuRepository interface {
	// Create creates a new menu
	Create(menu *Menu) error
	
	// FindByID finds a menu by ID
	FindByID(id int) (*Menu, error)
	
	// Update updates menu information
	Update(menu *Menu) error
	
	// Delete deletes a menu
	Delete(id int) error
	
	// FindByChefID finds menus by chef ID
	FindByChefID(chefID int) ([]*Menu, error)
	
	// FindActiveByChefID finds active menus by chef ID
	FindActiveByChefID(chefID int) ([]*Menu, error)
	
	// FindByType finds menus by type
	FindByType(menuType string, offset, limit int) ([]*Menu, error)
	
	// FindFeatured finds featured menus
	FindFeatured(offset, limit int) ([]*Menu, error)
}
