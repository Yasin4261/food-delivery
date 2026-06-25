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
)
