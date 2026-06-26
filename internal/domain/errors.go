package domain

import "errors"

// Domain errors. Services return these so handlers can map them to HTTP status
// codes without depending on storage or transport details.
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrAccountInactive       = errors.New("account is deactivated")

	ErrChefNotFound      = errors.New("chef not found")
	ErrChefProfileExists = errors.New("chef profile already exists for this user")

	ErrMenuNotFound     = errors.New("menu not found")
	ErrMenuItemNotFound = errors.New("menu item not found")

	// ErrForbidden marks an authenticated caller acting on a resource they do
	// not own. Handlers map it to HTTP 403.
	ErrForbidden = errors.New("you do not have permission to modify this resource")
)
