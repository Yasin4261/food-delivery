package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder handles order creation
// POST /api/orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "create order endpoint - to be implemented",
	})
}

// GetOrder retrieves an order by ID
// GET /api/orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "get order endpoint - to be implemented",
	})
}

// ListOrders retrieves orders for authenticated user
// GET /api/orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusNotImplemented, map[string]string{
		"message": "list orders endpoint - to be implemented",
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
