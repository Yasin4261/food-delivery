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

	minRating, ok1 := queryFloat(r, "min_rating")
	minPrice, ok2 := queryFloat(r, "min_price")
	maxPrice, ok3 := queryFloat(r, "max_price")
	if !ok1 || !ok2 || !ok3 {
		respondError(w, http.StatusBadRequest, "min_rating, min_price and max_price must be numbers")
		return
	}
	filters := domain.SearchFilters{
		MinRating:  minRating,
		MinPrice:   minPrice,
		MaxPrice:   maxPrice,
		Cuisine:    r.URL.Query().Get("cuisine"),
		Sort:       r.URL.Query().Get("sort"),
		Vegetarian: queryBool(r, "vegetarian"),
		Vegan:      queryBool(r, "vegan"),
		GlutenFree: queryBool(r, "gluten_free"),
		Halal:      queryBool(r, "halal"),
	}

	switch r.URL.Query().Get("type") {
	case "chef":
		chefs, total, err := h.search.Chefs(r.Context(), q, filters, limit, offset)
		if err != nil {
			respondSearchError(w, err)
			return
		}
		respondPage(w, chefs, limit, offset, total)
	case "food":
		foods, total, err := h.search.Foods(r.Context(), q, filters, limit, offset)
		if err != nil {
			respondSearchError(w, err)
			return
		}
		respondPage(w, foods, limit, offset, total)
	case "user":
		if claims.Role != domain.RoleAdmin {
			respondError(w, http.StatusForbidden, "user search is restricted to admins")
			return
		}
		users, total, err := h.search.Users(r.Context(), q, limit, offset)
		if err != nil {
			respondSearchError(w, err)
			return
		}
		respondPage(w, users, limit, offset, total)
	default:
		respondError(w, http.StatusBadRequest, "type must be one of chef, food or user")
	}
}

func respondSearchError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, ve.Msg)
		return
	}
	respondDomainError(w, err)
}
