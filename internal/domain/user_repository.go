package domain

import "context"

// UserRepository is the port the core needs for user persistence. The Postgres
// adapter lives in internal/repository. Lookups return ErrUserNotFound when no
// row matches.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
}
