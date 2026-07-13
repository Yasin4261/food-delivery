package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// OrderHandler exposes customer ordering and chef order-management endpoints.
type OrderHandler struct {
	orders *service.OrderService
}

// NewOrderHandler builds an OrderHandler.
func NewOrderHandler(orders *service.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

type orderLineRequest struct {
	MenuItemID          int    `json:"menu_item_id"`
	Quantity            int    `json:"quantity"`
	SpecialInstructions string `json:"special_instructions"`
}

type placeOrderRequest struct {
	// AddressID selects a saved address-book entry (mutually exclusive with
	// DeliveryAddress).
	AddressID         *int               `json:"address_id"`
	DeliveryAddress   string             `json:"delivery_address"`
	DeliveryCity      string             `json:"delivery_city"`
	DeliveryLatitude  *float64           `json:"delivery_latitude"`
	DeliveryLongitude *float64           `json:"delivery_longitude"`
	PaymentMethod     string             `json:"payment_method"`
	CustomerNotes     string             `json:"customer_notes"`
	Items             []orderLineRequest `json:"items"`
}

func (req placeOrderRequest) toInput() service.PlaceOrderInput {
	lines := make([]service.OrderLineInput, 0, len(req.Items))
	for _, it := range req.Items {
		lines = append(lines, service.OrderLineInput{
			MenuItemID:          it.MenuItemID,
			Quantity:            it.Quantity,
			SpecialInstructions: it.SpecialInstructions,
		})
	}
	return service.PlaceOrderInput{
		AddressID:         req.AddressID,
		DeliveryAddress:   req.DeliveryAddress,
		DeliveryCity:      req.DeliveryCity,
		DeliveryLatitude:  req.DeliveryLatitude,
		DeliveryLongitude: req.DeliveryLongitude,
		PaymentMethod:     req.PaymentMethod,
		CustomerNotes:     req.CustomerNotes,
		Lines:             lines,
	}
}

// --- customer ---

// Place handles POST /api/v2/orders (auth).
func (h *OrderHandler) Place(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req placeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	order, err := h.orders.PlaceOrder(r.Context(), claims.UserID, req.toInput())
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, order)
}

// List handles GET /api/v2/orders (auth) — the caller's order history.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	orders, total, err := h.orders.ListForCustomer(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondPage(w, orders, limit, offset, total)
}

// Get handles GET /api/v2/orders/{id} (auth, owner only).
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.orders.GetForCustomer(r.Context(), claims.UserID, id)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, order)
}

// Cancel handles POST /api/v2/orders/{id}/cancel (auth, owner only).
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := h.orders.CancelForCustomer(r.Context(), claims.UserID, id)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, order)
}

// Summary handles GET /api/v2/notifications/summary (auth): the lightweight
// counts the SPA polls for its navbar badges.
func (h *OrderHandler) Summary(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	summary, err := h.orders.Summary(r.Context(), claims.UserID, claims.Role == domain.RoleChef)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

// --- chef ---

// ChefList handles GET /api/v2/chef/orders (chef) — orders with the chef's
// items, scoped to that chef.
func (h *OrderHandler) ChefList(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	orders, total, err := h.orders.ListForChef(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondPage(w, orders, limit, offset, total)
}

type chefStatusRequest struct {
	Action string `json:"action"`
}

// ChefAdvance handles POST /api/v2/chef/orders/{id}/status (chef). The body
// carries the action: confirm, preparing, ready, delivering, delivered or
// decline.
func (h *OrderHandler) ChefAdvance(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	var req chefStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	order, err := h.orders.AdvanceForChef(r.Context(), claims.UserID, id, req.Action)
	if err != nil {
		respondOrderError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, order)
}

// respondOrderError maps validation and order-specific errors, falling back to
// the shared domain-error mapping.
func respondOrderError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, ve.Msg)
		return
	}
	respondDomainError(w, err)
}
