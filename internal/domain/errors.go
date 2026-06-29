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

	ErrResetTokenNotFound = errors.New("reset token not found")
	ErrInvalidResetToken  = errors.New("invalid, expired or already-used reset token")

	ErrChefNotFound      = errors.New("chef not found")
	ErrChefProfileExists = errors.New("chef profile already exists for this user")

	ErrMenuNotFound     = errors.New("menu not found")
	ErrMenuItemNotFound = errors.New("menu item not found")

	ErrOrderNotFound            = errors.New("order not found")
	ErrEmptyOrder               = errors.New("an order must contain at least one item")
	ErrItemNotOrderable         = errors.New("item is not available to order")
	ErrItemOutOfStock           = errors.New("item is out of stock")
	ErrInvalidStatusTransition  = errors.New("invalid order status transition")
	ErrInvalidPaymentTransition = errors.New("invalid payment status transition")

	ErrInvalidRating          = errors.New("rating must be between 1 and 5")
	ErrInvalidReviewTarget    = errors.New("a review must target exactly one of a chef or a dish")
	ErrOrderNotReviewable     = errors.New("only delivered orders can be reviewed")
	ErrReviewTargetNotInOrder = errors.New("you can only review a chef or dish from your own order")
	ErrReviewExists           = errors.New("you have already reviewed this for this order")

	// ErrForbidden marks an authenticated caller acting on a resource they do
	// not own. Handlers map it to HTTP 403.
	ErrForbidden = errors.New("you do not have permission to modify this resource")
)
