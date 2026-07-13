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
	auth     *service.AuthService
	denylist service.TokenRevoker
}

// NewAuthHandler builds an AuthHandler.
func NewAuthHandler(auth *service.AuthService, denylist service.TokenRevoker) *AuthHandler {
	return &AuthHandler{auth: auth, denylist: denylist}
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

// Logout handles POST /api/v2/auth/logout (auth). It revokes the presented
// token by adding its jti to the denylist until it would expire, so a stolen
// token can be invalidated before expiry.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	if h.denylist != nil && claims.ExpiresAt != nil {
		h.denylist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out; token revoked"})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// ForgotPassword handles POST /api/v2/auth/forgot-password. When the email is
// registered it emails a reset link; it always responds 202 with a generic
// message (and never the token) so it cannot be used to enumerate accounts.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.RequestPasswordReset(r.Context(), req.Email); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{
		"message": "if that email is registered, a reset link has been sent",
	})
}

// ResetPassword handles POST /api/v2/auth/reset-password.
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ResetPassword(r.Context(), req.Token, req.Password); err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "password updated; you can now log in"})
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

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword handles PUT /api/v2/auth/password (auth) — a logged-in user
// sets a new password after proving the current one.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.auth.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

type updateProfileRequest struct {
	PhoneNumber string   `json:"phone_number"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	State       string   `json:"state"`
	ZipCode     string   `json:"zip_code"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
}

// UpdateProfile handles PUT /api/v2/users/me (auth) — the caller edits their
// own contact/location fields. Identity fields (email, username, role) are
// not accepted here by design.
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.UpdateProfile(r.Context(), claims.UserID, service.UpdateProfileInput{
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		ZipCode:     req.ZipCode,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
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
	respondJSON(w, http.StatusOK, user)
}
