package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Yasin4261/food-delivery/internal/metrics"
)

// scrape returns the /metrics text output.
func scrape(t *testing.T, m *metrics.Metrics) string {
	t.Helper()
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("scrape status = %d, want 200", rec.Code)
	}
	return rec.Body.String()
}

func TestMetrics_HTTPMiddleware(t *testing.T) {
	m := metrics.New()
	// A handler returning 500 for /boom, 200 otherwise.
	h := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/boom" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ok", nil))
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/boom", nil))

	body := scrape(t, m)
	// Labelled by method + status, never by raw path (cardinality guard).
	if !strings.Contains(body, `http_requests_total{method="GET",status="200"} 3`) {
		t.Errorf("missing 200 count:\n%s", body)
	}
	if !strings.Contains(body, `http_requests_total{method="GET",status="500"} 1`) {
		t.Errorf("missing 500 count:\n%s", body)
	}
	if strings.Contains(body, `path=`) || strings.Contains(body, "/boom") {
		t.Error("metrics must not carry per-path labels (cardinality risk)")
	}
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Error("missing latency histogram")
	}
}

func TestMetrics_BusinessCounters(t *testing.T) {
	m := metrics.New()
	m.OrderPlaced()
	m.OrderPlaced()
	m.OrderDelivered()
	m.PaymentCompleted(true)
	m.PaymentCompleted(false)
	m.PaymentCompleted(false)

	body := scrape(t, m)
	for _, want := range []string{
		"orders_placed_total 2",
		"orders_delivered_total 1",
		`payments_total{outcome="success"} 1`,
		`payments_total{outcome="failed"} 2`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestMetrics_NilSafe(t *testing.T) {
	// A nil *Metrics (metrics disabled) must not panic — handlers call these
	// unconditionally.
	var m *metrics.Metrics
	m.OrderPlaced()
	m.OrderDelivered()
	m.PaymentCompleted(true)
}

func TestMetrics_RuntimeCollectors(t *testing.T) {
	body := scrape(t, metrics.New())
	if !strings.Contains(body, "go_goroutines") {
		t.Error("Go runtime collector not registered")
	}
}
