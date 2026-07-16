package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Yasin4261/food-delivery/internal/domain"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// ChefHandler exposes chef-profile endpoints.
type ChefHandler struct {
	chefs *service.ChefService
}

// NewChefHandler builds a ChefHandler.
func NewChefHandler(chefs *service.ChefService) *ChefHandler {
	return &ChefHandler{chefs: chefs}
}

type createChefRequest struct {
	BusinessName   string   `json:"business_name"`
	KitchenAddress string   `json:"kitchen_address"`
	Bio            string   `json:"bio"`
	Specialty      string   `json:"specialty"`
	KitchenCity    string   `json:"kitchen_city"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	DeliveryRadius int      `json:"delivery_radius"`
}

// Create handles POST /api/v2/chefs (auth required). The owner is taken from
// the token, never the request body.
func (h *ChefHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req createChefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	chef, err := h.chefs.CreateProfile(r.Context(), claims.UserID, service.CreateProfileInput{
		BusinessName:   req.BusinessName,
		KitchenAddress: req.KitchenAddress,
		Bio:            req.Bio,
		Specialty:      req.Specialty,
		KitchenCity:    req.KitchenCity,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		DeliveryRadius: req.DeliveryRadius,
	})
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, chef)
}

// Me handles GET /api/v2/chefs/me (chef) — the caller's own chef profile. The
// UI uses it to discover the chef id and current status; 404 means "no profile
// yet" and drives the onboarding flow.
func (h *ChefHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	chef, err := h.chefs.MyProfile(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chef)
}

// Get handles GET /api/v2/chefs/{id}.
func (h *ChefHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}

	chef, err := h.chefs.Get(r.Context(), id)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chef)
}

// List handles GET /api/v2/chefs?limit=&offset=.
func (h *ChefHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	minRating, ok := queryFloat(r, "min_rating")
	if !ok {
		respondError(w, http.StatusBadRequest, "min_rating must be a number")
		return
	}

	chefs, total, err := h.chefs.List(r.Context(), domain.ChefListFilters{
		OnlineOnly: queryBool(r, "online"),
		MinRating:  minRating,
		Sort:       r.URL.Query().Get("sort"),
	}, limit, offset)
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}
	respondPage(w, chefs, limit, offset, total)
}

// Nearby handles GET /api/v2/chefs/nearby?lat=&lng=&limit=&online=.
func (h *ChefHandler) Nearby(w http.ResponseWriter, r *http.Request) {
	lat, errLat := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, errLng := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	if errLat != nil || errLng != nil {
		respondError(w, http.StatusBadRequest, "lat and lng query parameters are required")
		return
	}

	limit := queryInt(r, "limit", 20)
	chefs, err := h.chefs.Nearby(r.Context(), lat, lng, limit, queryBool(r, "online"))
	if err != nil {
		respondDomainError(w, err)
		return
	}
	// Nearby is a proximity query (limit only); total is the page size.
	respondPage(w, chefs, limit, 0, len(chefs))
}

type setStatusRequest struct {
	IsOnline bool `json:"is_online"`
}

// SetStatus handles PATCH /api/v2/chefs/me/status (chef) — toggle presence.
func (h *ChefHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req setStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	chef, err := h.chefs.SetOnline(r.Context(), claims.UserID, req.IsOnline)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chef)
}

type setAvailabilityRequest struct {
	AcceptingOrders bool `json:"accepting_orders"`
}

// SetAvailability handles PATCH /api/v2/chefs/me/availability (chef) — toggle
// away / vacation mode. When off the chef is hidden from browse/search and
// cannot receive new orders.
func (h *ChefHandler) SetAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req setAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	chef, err := h.chefs.SetAcceptingOrders(r.Context(), claims.UserID, req.AcceptingOrders)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chef)
}

// queryInt reads an integer query parameter, falling back to def.
func queryInt(r *http.Request, key string, def int) int {
	if v, err := strconv.Atoi(r.URL.Query().Get(key)); err == nil {
		return v
	}
	return def
}

// queryFloat parses an optional float query parameter. A missing parameter
// yields 0 with ok=true; garbage yields ok=false so handlers can 400.
func queryFloat(r *http.Request, key string) (float64, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return 0, true
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// queryBool reports whether a query parameter is set to "true".
func queryBool(r *http.Request, key string) bool {
	return r.URL.Query().Get(key) == "true"
}

// UpdateMe handles PUT /api/v2/chefs/me (chef) — the caller edits their own
// kitchen profile. Ownership comes from the token: the service resolves the
// chef by user id, so no other chef's profile is reachable.
func (h *ChefHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req createChefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	chef, err := h.chefs.UpdateProfile(r.Context(), claims.UserID, service.CreateProfileInput{
		BusinessName:   req.BusinessName,
		KitchenAddress: req.KitchenAddress,
		Bio:            req.Bio,
		Specialty:      req.Specialty,
		KitchenCity:    req.KitchenCity,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		DeliveryRadius: req.DeliveryRadius,
	})
	if err != nil {
		var ve service.ValidationError
		if errors.As(err, &ve) {
			respondError(w, http.StatusBadRequest, ve.Msg)
			return
		}
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, chef)
}

// hoursEntry is the wire shape of one working-hours window: HH:MM strings,
// friendlier for the editor UI than raw minutes.
type hoursEntry struct {
	Weekday int    `json:"weekday"`
	Opens   string `json:"opens"`
	Closes  string `json:"closes"`
}

func parseClock(s string) (int, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

func formatClock(minutes int) string {
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

func toHoursEntries(hours []*domain.ChefHours) []hoursEntry {
	out := make([]hoursEntry, 0, len(hours))
	for _, h := range hours {
		out = append(out, hoursEntry{Weekday: h.Weekday, Opens: formatClock(h.OpensAt), Closes: formatClock(h.ClosesAt)})
	}
	return out
}

// SetHours handles PUT /api/v2/chefs/me/hours (chef role) — full replace of
// the weekly schedule; an empty list clears it (always open).
func (h *ChefHandler) SetHours(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req []hoursEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	hours := make([]*domain.ChefHours, 0, len(req))
	for _, e := range req {
		opens, err := parseClock(e.Opens)
		if err != nil {
			respondError(w, http.StatusBadRequest, "opens must be HH:MM")
			return
		}
		closes, err := parseClock(e.Closes)
		if err != nil {
			respondError(w, http.StatusBadRequest, "closes must be HH:MM")
			return
		}
		hours = append(hours, &domain.ChefHours{Weekday: e.Weekday, OpensAt: opens, ClosesAt: closes})
	}

	saved, err := h.chefs.SetHours(r.Context(), claims.UserID, hours)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, toHoursEntries(saved))
}

// Hours handles GET /api/v2/chefs/{id}/hours (public).
func (h *ChefHandler) Hours(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid chef id")
		return
	}
	hours, err := h.chefs.HoursFor(r.Context(), id)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, toHoursEntries(hours))
}
