package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// AuthHandler exposes the authentication endpoints. It is a thin transport
// layer: parse the request, call the service, translate the result to HTTP.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles POST /api/v2/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Register(r.Context(), service.RegisterInput{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		Role:        req.Role,
	})
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, result)
}

// Login handles POST /api/v2/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// Logout handles POST /api/v2/auth/logout. JWTs are stateless, so logout is a
// client-side action (discard the token). This endpoint exists to give the
// client a definite call to make; with a token blocklist it would revoke here.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out; discard your token"})
}

// Me handles GET /api/v2/auth/me. Requires authentication.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	user, err := h.auth.Profile(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}
