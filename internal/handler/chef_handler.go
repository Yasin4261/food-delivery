package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/service"
)

type ChefHandler struct {
	chefService *service.ChefService
}

func NewChefHandler(chefService *service.ChefService) *ChefHandler {
	return &ChefHandler{
		chefService: chefService,
	}
}

// CreateChef handles chef profile creation
// POST /api/chefs
func (h *ChefHandler) CreateChef(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "chef creation endpoint - to be implemented",
	})
}

// GetChef retrieves a chef by ID
// GET /api/chefs/{id}
func (h *ChefHandler) GetChef(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "get chef endpoint - to be implemented",
	})
}

// ListChefs retrieves a list of chefs
// GET /api/chefs
func (h *ChefHandler) ListChefs(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "list chefs endpoint - to be implemented",
	})
}

// FindNearbyChefs finds chefs within delivery radius
// GET /api/chefs/nearby
func (h *ChefHandler) FindNearbyChefs(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "nearby chefs endpoint - to be implemented",
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
