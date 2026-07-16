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

	ErrConversationNotFound = errors.New("conversation not found")
	ErrEmptyMessage         = errors.New("message body cannot be empty")

	ErrPaymentSessionNotFound = errors.New("payment session not found")
	ErrOrderNotPayable        = errors.New("only pending card orders can be paid online")

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

	// ErrUnsupportedImage marks an upload that is not a decodable JPEG or PNG.
	ErrUnsupportedImage = errors.New("image must be a valid JPEG or PNG")
	// ErrGalleryFull rejects a gallery upload past the per-dish cap.
	ErrGalleryFull = errors.New("this dish already has the maximum number of photos")

	ErrInvalidWeekday     = errors.New("weekday must be between 0 (Sunday) and 6 (Saturday)")
	ErrInvalidHoursWindow = errors.New("opening hours must be within the day and not empty")
	// ErrChefClosed rejects orders placed outside a chef's working hours.
	ErrChefClosed = errors.New("the chef is currently closed")

	ErrPromoNotFound      = errors.New("promo code not found")
	ErrPromoCodeRequired  = errors.New("promo code is required")
	ErrPromoInvalid       = errors.New("invalid promo code definition")
	ErrPromoExists        = errors.New("a promo code with that name already exists")
	ErrPromoNotRedeemable = errors.New("this promo code is not valid")
	ErrPromoExpired       = errors.New("this promo code has expired")
	ErrPromoUsedUp        = errors.New("this promo code has reached its usage limit")
	ErrPromoMinOrder      = errors.New("your order is below the minimum for this promo code")

	ErrAddressNotFound       = errors.New("address not found")
	ErrAddressLabelRequired  = errors.New("address label is required")
	ErrAddressLabelTooLong   = errors.New("address label must be at most 50 characters")
	ErrAddressRequired       = errors.New("address is required")
	ErrCoordinatesIncomplete = errors.New("latitude and longitude must be provided together")

	ErrInvalidRating          = errors.New("rating must be between 1 and 5")
	ErrInvalidReviewTarget    = errors.New("a review must target exactly one of a chef or a dish")
	ErrOrderNotReviewable     = errors.New("only delivered orders can be reviewed")
	ErrReviewTargetNotInOrder = errors.New("you can only review a chef or dish from your own order")
	ErrReviewExists           = errors.New("you have already reviewed this for this order")

	// ErrForbidden marks an authenticated caller acting on a resource they do
	// not own. Handlers map it to HTTP 403.
	ErrForbidden = errors.New("you do not have permission to modify this resource")
)
