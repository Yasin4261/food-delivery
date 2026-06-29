package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/middleware"
)

func TestRequestLogger_LogsStatusAndLatency(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := middleware.RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v2/chefs", nil))

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("expected an X-Request-ID response header")
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not JSON: %v (%s)", err, buf.String())
	}
	if entry["msg"] != "http_request" {
		t.Errorf("msg = %v, want http_request", entry["msg"])
	}
	if entry["status"] != float64(http.StatusTeapot) {
		t.Errorf("status = %v, want 418", entry["status"])
	}
	if entry["method"] != http.MethodGet || entry["path"] != "/api/v2/chefs" {
		t.Errorf("method/path wrong: %v %v", entry["method"], entry["path"])
	}
	if _, ok := entry["duration"]; !ok {
		t.Error("expected a duration field")
	}
	if entry["request_id"] == "" || entry["request_id"] == nil {
		t.Error("expected a request_id field")
	}
}

func TestRequestLogger_HonoursIncomingRequestID(t *testing.T) {
	h := middleware.RequestLogger(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := middleware.RequestIDFromContext(r.Context()); got != "abc123" {
				t.Errorf("request id in context = %q, want abc123", got)
			}
		}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "abc123")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") != "abc123" {
		t.Errorf("response request id = %q, want abc123", rec.Header().Get("X-Request-ID"))
	}
}
