package handler

import (
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// EarningsHandler exposes the chef earnings endpoint.
type EarningsHandler struct {
	earnings *service.EarningsService
}

// NewEarningsHandler builds an EarningsHandler.
func NewEarningsHandler(earnings *service.EarningsService) *EarningsHandler {
	return &EarningsHandler{earnings: earnings}
}

// Get handles GET /api/v2/chefs/me/earnings?days=N (chef). days omitted or <= 0
// means all-time.
func (h *EarningsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	earnings, err := h.earnings.ForChef(r.Context(), claims.UserID, queryInt(r, "days", 0))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, earnings)
}
