package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// AdminHandler exposes the admin/moderation endpoints. Every route is guarded
// by RequireRole(admin) at the router.
type AdminHandler struct {
	admin *service.AdminService
}

// NewAdminHandler builds an AdminHandler.
func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

type activeRequest struct {
	Active bool `json:"active"`
}

// ListUsers handles GET /api/v2/admin/users (admin).
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	users, total, err := h.admin.ListUsers(r.Context(), limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, users, limit, offset, total)
}

// SetUserActive handles PATCH /api/v2/admin/users/{id}/active (admin).
func (h *AdminHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req activeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// An admin deactivating their own account would lock everyone out of
	// moderation on a single-admin platform; forbid self-deactivation.
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok && claims.UserID == id && !req.Active {
		respondError(w, http.StatusUnprocessableEntity, "you cannot deactivate your own account")
		return
	}
	if err := h.admin.SetUserActive(r.Context(), id, req.Active); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"active": req.Active})
}

// ListChefs handles GET /api/v2/admin/chefs (admin) — all chefs incl inactive.
func (h *AdminHandler) ListChefs(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	chefs, total, err := h.admin.ListChefs(r.Context(), limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, chefs, limit, offset, total)
}

// SetChefActive handles PATCH /api/v2/admin/chefs/{id}/active (admin).
func (h *AdminHandler) SetChefActive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	var req activeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.admin.SetChefActive(r.Context(), id, req.Active); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"active": req.Active})
}

// ListOrders handles GET /api/v2/admin/orders (admin) — platform overview.
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	orders, total, err := h.admin.ListOrders(r.Context(), limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, orders, limit, offset, total)
}

// Stats handles GET /api/v2/admin/stats (admin).
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.admin.Stats(r.Context())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
