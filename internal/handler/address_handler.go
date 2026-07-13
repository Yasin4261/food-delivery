package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// AddressHandler exposes the customer address-book endpoints. Every route is
// authenticated; ownership is enforced in the service layer.
type AddressHandler struct {
	addresses *service.AddressService
}

// NewAddressHandler builds an AddressHandler.
func NewAddressHandler(addresses *service.AddressService) *AddressHandler {
	return &AddressHandler{addresses: addresses}
}

type addressRequest struct {
	Label     string   `json:"label"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	IsDefault bool     `json:"is_default"`
}

func (req addressRequest) toInput() service.AddressInput {
	return service.AddressInput{
		Label:     req.Label,
		Address:   req.Address,
		City:      req.City,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsDefault: req.IsDefault,
	}
}

// Create handles POST /api/v2/addresses (auth).
func (h *AddressHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req addressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	address, err := h.addresses.Create(r.Context(), claims.UserID, req.toInput())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, address)
}

// List handles GET /api/v2/addresses (auth) — the caller's own book.
func (h *AddressHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	addresses, err := h.addresses.List(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, addresses)
}

// Update handles PUT /api/v2/addresses/{id} (auth, owner).
func (h *AddressHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid address id")
		return
	}
	var req addressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	address, err := h.addresses.Update(r.Context(), claims.UserID, id, req.toInput())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, address)
}

// Delete handles DELETE /api/v2/addresses/{id} (auth, owner).
func (h *AddressHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid address id")
		return
	}
	if err := h.addresses.Delete(r.Context(), claims.UserID, id); err != nil {
		respondDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
