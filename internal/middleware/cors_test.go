package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/middleware"
)

func corsHandler(origins ...string) http.Handler {
	return middleware.CORS(origins)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestCORS_AllowedOrigin(t *testing.T) {
	h := corsHandler("https://app.example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/chefs", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO = %q, want the allowed origin", got)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	h := corsHandler("https://app.example.com")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/chefs", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("disallowed origin must not receive an Access-Control-Allow-Origin header")
	}
}

func TestCORS_Preflight(t *testing.T) {
	h := corsHandler("https://app.example.com")

	req := httptest.NewRequest(http.MethodOptions, "/api/v2/chefs", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("preflight should advertise allowed methods")
	}
}

func TestCORS_Wildcard(t *testing.T) {
	h := corsHandler("*")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://anything.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("wildcard ACAO = %q, want *", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
