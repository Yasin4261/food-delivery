package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// MenuHandler exposes menu and dish endpoints. Mutating endpoints require the
// chef role (enforced by the router) and resolve the owner from the token.
type MenuHandler struct {
	menus *service.MenuService
}

// NewMenuHandler builds a MenuHandler.
func NewMenuHandler(menus *service.MenuService) *MenuHandler {
	return &MenuHandler{menus: menus}
}

type menuRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	MenuType      string `json:"menu_type"`
	AvailableDays string `json:"available_days"`
	IsFeatured    bool   `json:"is_featured"`
}

func (req menuRequest) toInput() service.CreateMenuInput {
	return service.CreateMenuInput{
		Name:          req.Name,
		Description:   req.Description,
		MenuType:      req.MenuType,
		AvailableDays: req.AvailableDays,
		IsFeatured:    req.IsFeatured,
	}
}

type itemRequest struct {
	MenuID int `json:"menu_id"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Cuisine     string `json:"cuisine"`
	PortionSize string `json:"portion_size"`
	ImageURL    string `json:"image_url"`

	Price           float64  `json:"price"`
	OriginalPrice   *float64 `json:"original_price"`
	PreparationTime *int     `json:"preparation_time"`
	ServingSize     *int     `json:"serving_size"`

	AvailableQuantity *int `json:"available_quantity"`
	IsUnlimited       bool `json:"is_unlimited"`

	IsVegetarian bool `json:"is_vegetarian"`
	IsVegan      bool `json:"is_vegan"`
	IsGlutenFree bool `json:"is_gluten_free"`
	IsHalal      bool `json:"is_halal"`
	IsSpicy      bool `json:"is_spicy"`
	SpiceLevel   *int `json:"spice_level"`
}

func (req itemRequest) toInput() service.CreateItemInput {
	return service.CreateItemInput{
		MenuID:            req.MenuID,
		Name:              req.Name,
		Description:       req.Description,
		Category:          req.Category,
		Cuisine:           req.Cuisine,
		PortionSize:       req.PortionSize,
		ImageURL:          req.ImageURL,
		Price:             req.Price,
		OriginalPrice:     req.OriginalPrice,
		PreparationTime:   req.PreparationTime,
		ServingSize:       req.ServingSize,
		AvailableQuantity: req.AvailableQuantity,
		IsUnlimited:       req.IsUnlimited,
		IsVegetarian:      req.IsVegetarian,
		IsVegan:           req.IsVegan,
		IsGlutenFree:      req.IsGlutenFree,
		IsHalal:           req.IsHalal,
		IsSpicy:           req.IsSpicy,
		SpiceLevel:        req.SpiceLevel,
	}
}

// --- menus ---

// CreateMenu handles POST /api/v2/menus (chef).
func (h *MenuHandler) CreateMenu(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req menuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	menu, err := h.menus.CreateMenu(r.Context(), claims.UserID, req.toInput())
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, menu)
}

// UpdateMenu handles PUT /api/v2/menus/{id} (chef).
func (h *MenuHandler) UpdateMenu(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	var req menuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	menu, err := h.menus.UpdateMenu(r.Context(), claims.UserID, id, req.toInput())
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, menu)
}

// DeleteMenu handles DELETE /api/v2/menus/{id} (chef). It soft-deletes.
func (h *MenuHandler) DeleteMenu(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	if err := h.menus.DeactivateMenu(r.Context(), claims.UserID, id); err != nil {
		respondMenuError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetMenu handles GET /api/v2/menus/{id} (public).
func (h *MenuHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	menu, err := h.menus.GetMenu(r.Context(), id)
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, menu)
}

// ListChefMenus handles GET /api/v2/chefs/{id}/menus (public).
func (h *MenuHandler) ListChefMenus(w http.ResponseWriter, r *http.Request) {
	chefID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	menus, total, err := h.menus.ListChefMenus(r.Context(), chefID, limit, offset)
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondPage(w, menus, limit, offset, total)
}

// --- dishes ---

// CreateItem handles POST /api/v2/menu-items (chef). menu_id is in the body.
func (h *MenuHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req itemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.menus.CreateItem(r.Context(), claims.UserID, req.toInput())
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

// UpdateItem handles PUT /api/v2/menu-items/{id} (chef).
func (h *MenuHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	var req itemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	item, err := h.menus.UpdateItem(r.Context(), claims.UserID, id, req.toInput())
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

// DeleteItem handles DELETE /api/v2/menu-items/{id} (chef). It soft-deletes.
func (h *MenuHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if err := h.menus.DeactivateItem(r.Context(), claims.UserID, id); err != nil {
		respondMenuError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListMenuItems handles GET /api/v2/menus/{id}/items (public).
func (h *MenuHandler) ListMenuItems(w http.ResponseWriter, r *http.Request) {
	menuID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu id")
		return
	}
	items, err := h.menus.ListMenuItems(r.Context(), menuID)
	if err != nil {
		respondMenuError(w, err)
		return
	}
	// All active items in the menu are returned (not offset-paginated).
	respondPage(w, items, len(items), 0, len(items))
}

// ListChefItems handles GET /api/v2/chefs/{id}/menu-items (public).
func (h *MenuHandler) ListChefItems(w http.ResponseWriter, r *http.Request) {
	chefID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	items, total, err := h.menus.ListChefItems(r.Context(), chefID, limit, offset)
	if err != nil {
		respondMenuError(w, err)
		return
	}
	respondPage(w, items, limit, offset, total)
}

// respondMenuError maps validation errors to 400 and otherwise defers to the
// shared domain-error mapping.
func respondMenuError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, ve.Msg)
		return
	}
	respondDomainError(w, err)
}
