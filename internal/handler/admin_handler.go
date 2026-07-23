package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// AdminHandler exposes the admin/moderation endpoints. Every route is guarded
// by RequireRole(admin) at the router.
type AdminHandler struct {
	admin *service.AdminService
}

// NewAdminHandler builds an AdminHandler.
func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

type activeRequest struct {
	Active bool   `json:"active"`
	Reason string `json:"reason"`
}

// actorID returns the authenticated admin's user id, or false when unauthenticated.
func actorID(r *http.Request) (int, bool) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

// queryTriBool parses an optional boolean filter: absent (or unparseable) means
// "either", so the caller can distinguish "only active", "only inactive" and
// "both" with one parameter.
func queryTriBool(r *http.Request, key string) *bool {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}
	return &v
}

// respondAdminError maps a service validation failure (an unknown filter value)
// to 400 and everything else through the shared domain mapping.
func respondAdminError(w http.ResponseWriter, err error) {
	var ve service.ValidationError
	if errors.As(err, &ve) {
		respondError(w, http.StatusBadRequest, ve.Msg)
		return
	}
	respondDomainError(w, err)
}

// ListUsers handles GET /api/v2/admin/users (admin), narrowed by
// ?q=&role=&active= — the support console's "find this person" query.
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	f := domain.AdminUserFilters{
		Query:  r.URL.Query().Get("q"),
		Role:   r.URL.Query().Get("role"),
		Active: queryTriBool(r, "active"),
	}
	users, total, err := h.admin.ListUsers(r.Context(), f, limit, offset)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondPage(w, users, limit, offset, total)
}

// SetUserActive handles PATCH /api/v2/admin/users/{id}/active (admin).
func (h *AdminHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req activeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	// An admin deactivating their own account would lock everyone out of
	// moderation on a single-admin platform; forbid self-deactivation.
	if actor == id && !req.Active {
		respondError(w, http.StatusUnprocessableEntity, "you cannot deactivate your own account")
		return
	}
	if err := h.admin.SetUserActive(r.Context(), actor, id, req.Active, req.Reason); err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"active": req.Active})
}

// ListChefs handles GET /api/v2/admin/chefs (admin) — all chefs incl inactive,
// narrowed by ?q=&active=.
func (h *AdminHandler) ListChefs(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	f := domain.AdminChefFilters{
		Query:  r.URL.Query().Get("q"),
		Active: queryTriBool(r, "active"),
	}
	chefs, total, err := h.admin.ListChefs(r.Context(), f, limit, offset)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondPage(w, chefs, limit, offset, total)
}

// SetChefActive handles PATCH /api/v2/admin/chefs/{id}/active (admin).
func (h *AdminHandler) SetChefActive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	var req activeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	if err := h.admin.SetChefActive(r.Context(), actor, id, req.Active, req.Reason); err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"active": req.Active})
}

type adminChefStatusRequest struct {
	Online bool   `json:"online"`
	Reason string `json:"reason"`
}

// SetChefStatus handles PATCH /api/v2/admin/chefs/{id}/status (admin) — drive a
// chef's online presence on their behalf.
func (h *AdminHandler) SetChefStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "chef")
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req adminChefStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.admin.SetChefOnline(r.Context(), actor, id, req.Online, req.Reason); err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"online": req.Online})
}

type chefAvailabilityRequest struct {
	AcceptingOrders bool   `json:"accepting_orders"`
	Reason          string `json:"reason"`
}

// SetChefAvailability handles PATCH /api/v2/admin/chefs/{id}/availability
// (admin) — pause/reopen a chef's orders on their behalf.
func (h *AdminHandler) SetChefAvailability(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "chef")
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req chefAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.admin.SetChefAcceptingOrders(r.Context(), actor, id, req.AcceptingOrders, req.Reason); err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"accepting_orders": req.AcceptingOrders})
}

// ListOrders handles GET /api/v2/admin/orders (admin) — platform overview,
// narrowed by ?status=&payment_status=&user_id= (the last scopes to one
// customer's orders, for support drill-in).
func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 20), queryInt(r, "offset", 0)
	f := domain.AdminOrderFilters{
		Status:        r.URL.Query().Get("status"),
		PaymentStatus: r.URL.Query().Get("payment_status"),
		UserID:        queryInt(r, "user_id", 0),
	}
	orders, total, err := h.admin.ListOrders(r.Context(), f, limit, offset)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondPage(w, orders, limit, offset, total)
}

// Stats handles GET /api/v2/admin/stats (admin).
func (h *AdminHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.admin.Stats(r.Context())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

type createPromoRequest struct {
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinOrder      float64    `json:"min_order"`
	ValidFrom     *time.Time `json:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until"`
	UsageLimit    int        `json:"usage_limit"`
}

// ListPromos handles GET /api/v2/admin/promos (admin).
func (h *AdminHandler) ListPromos(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 50), queryInt(r, "offset", 0)
	promos, total, err := h.admin.ListPromos(r.Context(), limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, promos, limit, offset, total)
}

func (r *createPromoRequest) input() service.PromoInput {
	return service.PromoInput{
		Code:          r.Code,
		DiscountType:  r.DiscountType,
		DiscountValue: r.DiscountValue,
		MinOrder:      r.MinOrder,
		ValidFrom:     r.ValidFrom,
		ValidUntil:    r.ValidUntil,
		UsageLimit:    r.UsageLimit,
	}
}

// CreatePromo handles POST /api/v2/admin/promos (admin).
func (h *AdminHandler) CreatePromo(w http.ResponseWriter, r *http.Request) {
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req createPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	promo, err := h.admin.CreatePromo(r.Context(), actor, req.input())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, promo)
}

// UpdatePromo handles PUT /api/v2/admin/promos/{id} (admin) — edit a code's
// definition (never its usage counter).
func (h *AdminHandler) UpdatePromo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "promo")
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req createPromoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	promo, err := h.admin.UpdatePromo(r.Context(), actor, id, req.input())
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, promo)
}

// DeletePromo handles DELETE /api/v2/admin/promos/{id} (admin).
func (h *AdminHandler) DeletePromo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "promo")
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	if err := h.admin.DeletePromo(r.Context(), actor, id); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "promo deleted"})
}

// SetPromoActive handles PATCH /api/v2/admin/promos/{id}/active (admin).
func (h *AdminHandler) SetPromoActive(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "promo")
	if !ok {
		return
	}
	actor, ok := actorID(r)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req activeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.admin.SetPromoActive(r.Context(), actor, id, req.Active); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"active": req.Active})
}

// ListAudit handles GET /api/v2/admin/audit (admin) — the read-only audit log,
// narrowed by ?action=&target_type=&target_id=&actor_id=.
func (h *AdminHandler) ListAudit(w http.ResponseWriter, r *http.Request) {
	limit, offset := queryInt(r, "limit", 50), queryInt(r, "offset", 0)
	f := domain.AuditFilters{
		Action:     r.URL.Query().Get("action"),
		TargetType: r.URL.Query().Get("target_type"),
		TargetID:   queryInt(r, "target_id", 0),
		ActorID:    queryInt(r, "actor_id", 0),
	}
	entries, total, err := h.admin.ListAudit(r.Context(), f, limit, offset)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondPage(w, entries, limit, offset, total)
}

// pathID parses the {id} path value, replying 400 with a caller-specific
// message when it is not a number.
func pathID(w http.ResponseWriter, r *http.Request, what string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid "+what+" id")
		return 0, false
	}
	return id, true
}

// UserDetail handles GET /api/v2/admin/users/{id} (admin) — the support
// console's drill-in on one account.
func (h *AdminHandler) UserDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "user")
	if !ok {
		return
	}
	detail, err := h.admin.UserDetail(r.Context(), id)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, detail)
}

// OrderDetail handles GET /api/v2/admin/orders/{id} (admin).
func (h *AdminHandler) OrderDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "order")
	if !ok {
		return
	}
	detail, err := h.admin.OrderDetail(r.Context(), id)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, detail)
}

// ChefDetail handles GET /api/v2/admin/chefs/{id} (admin). Unlike the public
// chef endpoint this resolves deactivated kitchens too.
func (h *AdminHandler) ChefDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "chef")
	if !ok {
		return
	}
	detail, err := h.admin.ChefDetail(r.Context(), id)
	if err != nil {
		respondAdminError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, detail)
}
