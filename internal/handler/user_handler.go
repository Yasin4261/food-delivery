package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register handles user registration
// POST /api/auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Register user
	authResp, err := h.userService.Register(r.Context(), req)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to register user: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, authResp)
}

// Login handles user authentication
// POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Authenticate user
	authResp, err := h.userService.Login(r.Context(), req)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "invalid credentials" {
			respondWithError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if err.Error() == "user account is deactivated" {
			respondWithError(w, http.StatusForbidden, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to login: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, authResp)
}

// GetProfile retrieves the authenticated user's profile
// GET /api/auth/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context (injected by auth middleware)
	claims, err := middleware.GetUserClaims(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get user profile
	user, err := h.userService.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
// PUT /api/auth/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, err := middleware.GetUserClaims(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update profile
	user, err := h.userService.UpdateProfile(r.Context(), claims.UserID, updates)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update profile: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateLocation updates the authenticated user's location
// PUT /api/auth/location
func (h *UserHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, err := middleware.GetUserClaims(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Update location
	if err := h.userService.UpdateLocation(r.Context(), claims.UserID, req.Latitude, req.Longitude); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update location: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "location updated successfully"})
}

// ChangePassword changes the authenticated user's password
// POST /api/auth/change-password
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user claims from context
	claims, err := middleware.GetUserClaims(r.Context())
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Parse request body
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Change password
	if err := h.userService.ChangePassword(r.Context(), claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		if err.Error() == "current password is incorrect" {
			respondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to change password: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "password changed successfully"})
}

// ListUsers retrieves a paginated list of users (admin only)
// GET /api/users?limit=10&offset=0&role=customer
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	role := r.URL.Query().Get("role")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := make(map[string]interface{})
	if role != "" {
		filters["role"] = role
	}

	// List users
	users, err := h.userService.ListUsers(r.Context(), filters, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to list users: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	})
}

// FindNearbyUsers finds users within a specified radius
// GET /api/users/nearby?lat=40.7128&lng=-74.0060&radius=5&limit=20
func (h *UserHandler) FindNearbyUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	radiusStr := r.URL.Query().Get("radius")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if latStr == "" || lngStr == "" {
		respondWithError(w, http.StatusBadRequest, "latitude and longitude are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid latitude")
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid longitude")
		return
	}

	radius := 10.0 // Default 10km
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil && r > 0 {
			radius = r
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Find nearby users
	users, err := h.userService.FindNearbyUsers(r.Context(), lat, lng, radius, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to find nearby users: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"users":  users,
		"lat":    lat,
		"lng":    lng,
		"radius": radius,
		"limit":  limit,
		"offset": offset,
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError sends a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
