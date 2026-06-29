package handler

import (
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// FavoriteHandler exposes the customer favorite-chef endpoints.
type FavoriteHandler struct {
	favorites *service.FavoriteService
}

// NewFavoriteHandler builds a FavoriteHandler.
func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

// Add handles POST /api/v2/favorites/{chefId} (auth). Idempotent.
func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	chefID, err := strconv.Atoi(r.PathValue("chefId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	if err := h.favorites.Add(r.Context(), claims.UserID, chefID); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Remove handles DELETE /api/v2/favorites/{chefId} (auth). Idempotent.
func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	chefID, err := strconv.Atoi(r.PathValue("chefId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	if err := h.favorites.Remove(r.Context(), claims.UserID, chefID); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /api/v2/favorites (auth) — the caller's favorite chefs.
func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	chefs, err := h.favorites.List(r.Context(), claims.UserID, queryInt(r, "limit", 20), queryInt(r, "offset", 0))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chefs)
}
