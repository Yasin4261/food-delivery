package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Yasin4261/food-delivery/internal/middleware"
	"github.com/Yasin4261/food-delivery/internal/service"
)

// AccountHandler exposes the account data-rights endpoints (#107): exporting
// the caller's data and deleting their account.
type AccountHandler struct {
	account  *service.AccountService
	denylist service.TokenRevoker
}

// NewAccountHandler builds an AccountHandler.
func NewAccountHandler(account *service.AccountService, denylist service.TokenRevoker) *AccountHandler {
	return &AccountHandler{account: account, denylist: denylist}
}

// Export handles GET /api/v2/users/me/export (auth) — a JSON dump of the
// caller's own data, offered as a download.
func (h *AccountHandler) Export(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	export, err := h.account.Export(r.Context(), claims.UserID)
	if err != nil {
		respondDomainError(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="my-data.json"`)
	respondJSON(w, http.StatusOK, export)
}

type deleteAccountRequest struct {
	Password string `json:"password"`
}

// Delete handles DELETE /api/v2/users/me (auth) — irreversibly anonymises the
// caller's account after verifying their password, then revokes the token.
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	var req deleteAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.account.Delete(r.Context(), claims.UserID, req.Password); err != nil {
		respondDomainError(w, err)
		return
	}
	// The account is gone — invalidate the presented token immediately.
	if h.denylist != nil && claims.ExpiresAt != nil {
		h.denylist.Revoke(claims.ID, claims.ExpiresAt.Time)
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "account deleted"})
}
