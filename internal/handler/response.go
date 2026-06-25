package handler

import (
	"encoding/json"
	"errors"
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
	case errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrUsernameAlreadyExists):
		respondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrAccountInactive):
		respondError(w, http.StatusUnauthorized, err.Error())
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}
