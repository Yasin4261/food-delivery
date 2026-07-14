package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/domain"
)

// respondJSON writes payload as a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// page is the standard envelope for list endpoints: the items plus the paging
// window and the total number of matching rows.
type page[T any] struct {
	Data   []T `json:"data"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// respondPage writes a paginated list envelope. A nil slice is serialised as an
// empty array, never null.
func respondPage[T any](w http.ResponseWriter, items []T, limit, offset, total int) {
	if items == nil {
		items = []T{}
	}
	respondJSON(w, http.StatusOK, page[T]{Data: items, Limit: limit, Offset: offset, Total: total})
}

// respondError writes a JSON {"error": msg} body with the given status code.
func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

// respondDomainError maps a domain error to the right HTTP status. Unknown
// errors become 500 without leaking internal detail.
func respondDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrChefNotFound),
		errors.Is(err, domain.ErrMenuNotFound),
		errors.Is(err, domain.ErrMenuItemNotFound),
		errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrConversationNotFound),
		errors.Is(err, domain.ErrPaymentSessionNotFound),
		errors.Is(err, domain.ErrAddressNotFound):
		respondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		respondError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, domain.ErrInvalidRating),
		errors.Is(err, domain.ErrInvalidReviewTarget),
		errors.Is(err, domain.ErrInvalidResetToken),
		errors.Is(err, domain.ErrEmptyMessage),
		errors.Is(err, domain.ErrUnsupportedImage),
		errors.Is(err, domain.ErrAddressLabelRequired),
		errors.Is(err, domain.ErrAddressLabelTooLong),
		errors.Is(err, domain.ErrAddressRequired),
		errors.Is(err, domain.ErrCoordinatesIncomplete):
		respondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrEmptyOrder),
		errors.Is(err, domain.ErrItemNotOrderable),
		errors.Is(err, domain.ErrItemOutOfStock),
		errors.Is(err, domain.ErrInvalidStatusTransition),
		errors.Is(err, domain.ErrInvalidPaymentTransition),
		errors.Is(err, domain.ErrOrderNotReviewable),
		errors.Is(err, domain.ErrReviewTargetNotInOrder),
		errors.Is(err, domain.ErrOrderNotPayable):
		respondError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrUsernameAlreadyExists),
		errors.Is(err, domain.ErrChefProfileExists),
		errors.Is(err, domain.ErrReviewExists):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrAccountInactive):
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		// Unexpected errors are logged server-side (the client only ever sees
		// the generic message) — otherwise 500s are undiagnosable in prod.
		slog.Error("unhandled error in request", "error", err)
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
