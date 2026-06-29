package handler

import (
	"errors"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// SearchHandler exposes catalogue search.
type SearchHandler struct {
	search *service.SearchService
}

// NewSearchHandler builds a SearchHandler.
func NewSearchHandler(search *service.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

// Search handles GET /api/v2/search?q=&type=&limit=&offset= (auth). type is one
// of chef, food or user; user search is restricted to admins.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	q := r.URL.Query().Get("q")
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)

	var (
		result any
		err    error
	)
	switch r.URL.Query().Get("type") {
	case "chef":
		result, err = h.search.Chefs(r.Context(), q, limit, offset)
	case "food":
		result, err = h.search.Foods(r.Context(), q, limit, offset)
	case "user":
		if claims.Role != domain.RoleAdmin {
			respondError(w, http.StatusForbidden, "user search is restricted to admins")
			return
		}
		result, err = h.search.Users(r.Context(), q, limit, offset)
	default:
		respondError(w, http.StatusBadRequest, "type must be one of chef, food or user")
		return
	}

	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}
