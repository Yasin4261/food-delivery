package domain

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(user *User) error
	
	// FindByID finds a user by ID
	FindByID(id int) (*User, error)
	
	// FindByEmail finds a user by email
	FindByEmail(email string) (*User, error)
	
	// FindByUsername finds a user by username
	FindByUsername(username string) (*User, error)
	
	// Update updates user information
	Update(user *User) error
	
	// Delete deletes a user (soft delete)
	Delete(id int) error
	
	// UpdateLocation updates user's location
	UpdateLocation(id int, lat, lng float64, address, city, state, zipCode string) error
	
	// List returns paginated users
	List(offset, limit int) ([]*User, error)
	
	// CountByRole counts users by role
	CountByRole(role string) (int, error)
	
	// FindNearby finds users within radius (km) from a location
	FindNearby(lat, lng, radiusKm float64, limit int) ([]*User, error)
}
