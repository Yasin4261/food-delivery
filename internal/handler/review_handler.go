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

// ReviewHandler exposes the rating endpoints.
type ReviewHandler struct {
	reviews *service.ReviewService
}

// NewReviewHandler builds a ReviewHandler.
func NewReviewHandler(reviews *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviews: reviews}
}

type createReviewRequest struct {
	OrderID    int    `json:"order_id"`
	ChefID     *int   `json:"chef_id"`
	MenuItemID *int   `json:"menu_item_id"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

// Create handles POST /api/v2/reviews (auth).
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req createReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	review, err := h.reviews.Create(r.Context(), claims.UserID, service.CreateReviewInput{
		OrderID:    req.OrderID,
		ChefID:     req.ChefID,
		MenuItemID: req.MenuItemID,
		Rating:     req.Rating,
		Comment:    req.Comment,
	})
	if err != nil {
		respondReviewError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, review)
}

// ListForChef handles GET /api/v2/chefs/{id}/reviews (public).
func (h *ReviewHandler) ListForChef(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	reviews, total, err := h.reviews.ListForChef(r.Context(), id, limit, offset)
	if err != nil {
		respondReviewError(w, err)
		return
	}
	respondPage(w, reviews, limit, offset, total)
}

// ListForMenuItem handles GET /api/v2/menu-items/{id}/reviews (public).
func (h *ReviewHandler) ListForMenuItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid menu item id")
		return
	}
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	reviews, total, err := h.reviews.ListForMenuItem(r.Context(), id, limit, offset)
	if err != nil {
		respondReviewError(w, err)
		return
	}
	respondPage(w, reviews, limit, offset, total)
}

// respondReviewError maps validation errors to 400 and otherwise defers to the
// shared domain-error mapping.
func respondReviewError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, ve.Msg)
		return
	}
	respondDomainError(w, err)
}

// ListForOrder handles GET /api/v2/orders/{id}/reviews (auth) — the caller's
// own rating history for that order. The query is user-scoped, so asking
// about someone else's order simply returns an empty list.
func (h *ReviewHandler) ListForOrder(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	orderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	reviews, err := h.reviews.ListForOrder(r.Context(), claims.UserID, orderID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	if reviews == nil {
		reviews = []*domain.Review{}
	}
	respondJSON(w, http.StatusOK, reviews)
}
