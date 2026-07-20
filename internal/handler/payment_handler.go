package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Yasin4261/food-delivery/internal/metrics"
	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// PaymentHandler exposes the card-payment endpoints: starting a hosted
// checkout and receiving the gateway's browser callback.
type PaymentHandler struct {
	payments *service.PaymentService
	metrics  *metrics.Metrics // payment outcome counter; nil-safe
}

// NewPaymentHandler builds a PaymentHandler. m may be nil (metrics disabled).
func NewPaymentHandler(payments *service.PaymentService, m *metrics.Metrics) *PaymentHandler {
	return &PaymentHandler{payments: payments, metrics: m}
}

// payRequest is the optional body of POST /orders/{id}/pay. save_card opts the
// customer in to storing the card used this time (#67).
type payRequest struct {
	SaveCard bool `json:"save_card"`
}

// Pay handles POST /api/v2/orders/{id}/pay (auth, owner). It returns the
// hosted payment page the browser should navigate to.
func (h *PaymentHandler) Pay(w http.ResponseWriter, r *http.Request) {
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
	// The body is optional (a bare POST is a plain checkout).
	var req payRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	url, err := h.payments.StartCheckout(r.Context(), claims.UserID, id, req.SaveCard)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"payment_page_url": url})
}

// ListMethods handles GET /api/v2/payment-methods (auth). It returns the
// caller's saved cards (masked metadata + gateway token only).
func (h *PaymentHandler) ListMethods(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	cards, err := h.payments.ListSavedCards(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": cards})
}

// DeleteMethod handles DELETE /api/v2/payment-methods/{token} (auth, owner). It
// revokes the card at the gateway and removes the saved association.
func (h *PaymentHandler) DeleteMethod(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	token := r.PathValue("token")
	if token == "" {
		respondError(w, http.StatusBadRequest, "invalid card token")
		return
	}
	if err := h.payments.DeleteSavedCard(r.Context(), claims.UserID, token); err != nil {
		respondDomainError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "card removed"})
}

// Callback handles POST /api/v2/payments/callback. The gateway redirects the
// customer's browser here with a checkout token (form-encoded); the outcome is
// verified server-to-server, then the browser is sent back to the SPA. It is
// public by necessity (the browser carries no bearer token on this hop) and
// rate limited at the router.
func (h *PaymentHandler) Callback(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	token := r.FormValue("token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if token == "" {
		http.Redirect(w, r, h.payments.ErrorRedirectURL(), http.StatusSeeOther)
		return
	}

	orderID, paid, err := h.payments.CompleteCheckout(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, h.payments.ErrorRedirectURL(), http.StatusSeeOther)
		return
	}
	h.metrics.PaymentCompleted(paid)
	http.Redirect(w, r, h.payments.ResultRedirectURL(orderID, paid), http.StatusSeeOther)
}
