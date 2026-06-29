package handler

import (
	"context"
	"net/http"
	"time"
)

// Pinger is the minimal contract the health check needs from the database.
// Keeping it an interface here means the handler does not depend on the
// concrete database package.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// HealthHandler reports service liveness and database reachability.
type HealthHandler struct {
	db Pinger
}

// NewHealthHandler builds a HealthHandler.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck returns 200 when the API and database are healthy, otherwise 503.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := map[string]string{"status": "ok", "database": "ok"}
	code := http.StatusOK

	if err := h.db.PingContext(ctx); err != nil {
		resp["status"] = "degraded"
		resp["database"] = "unreachable"
		code = http.StatusServiceUnavailable
	}

	respondJSON(w, code, resp)
}
