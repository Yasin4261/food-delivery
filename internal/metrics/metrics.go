// Package metrics exposes Prometheus instrumentation for the API (#73). It
// owns a private registry (no global state, so tests don't collide) carrying
// Go runtime, process, DB-pool and application collectors, plus the HTTP
// instrumentation middleware and the /metrics handler.
//
// Cardinality note: HTTP metrics are labelled by method and status code only —
// never by the raw request path, which contains IDs (/orders/123) and would
// blow up the time-series count. Route-level breakdown is deliberately traded
// away for a bounded, safe label set.
package metrics

import (
	"bufio"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds the registry and the application collectors.
type Metrics struct {
	registry *prometheus.Registry

	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec

	ordersPlaced    prometheus.Counter
	ordersDelivered prometheus.Counter
	payments        *prometheus.CounterVec
}

// New builds a Metrics with Go/process collectors registered. Call
// RegisterDB once the database handle exists to add pool stats.
func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := &Metrics{
		registry: reg,
		httpRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method and status code.",
		}, []string{"method", "status"}),
		httpDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency by method.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method"}),
		ordersPlaced: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "orders_placed_total",
			Help: "Orders successfully placed.",
		}),
		ordersDelivered: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "orders_delivered_total",
			Help: "Orders that reached the delivered state.",
		}),
		payments: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "payments_total",
			Help: "Card payment verification outcomes.",
		}, []string{"outcome"}),
	}
	reg.MustRegister(m.httpRequests, m.httpDuration, m.ordersPlaced, m.ordersDelivered, m.payments)
	return m
}

// RegisterDB adds connection-pool gauges (open/idle/in-use, waits) for db.
func (m *Metrics) RegisterDB(db *sql.DB) {
	m.registry.MustRegister(collectors.NewDBStatsCollector(db, "food_delivery"))
}

// Handler serves the metrics in the Prometheus text exposition format.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware records request count and latency. It reads the final status from
// a lightweight response wrapper.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		m.httpRequests.WithLabelValues(r.Method, strconv.Itoa(rec.status)).Inc()
		m.httpDuration.WithLabelValues(r.Method).Observe(time.Since(start).Seconds())
	})
}

// --- application counters (nil-safe so wiring is optional) ---

// OrderPlaced increments the placed-orders counter.
func (m *Metrics) OrderPlaced() {
	if m != nil {
		m.ordersPlaced.Inc()
	}
}

// OrderDelivered increments the delivered-orders counter.
func (m *Metrics) OrderDelivered() {
	if m != nil {
		m.ordersDelivered.Inc()
	}
}

// PaymentCompleted records a card payment verification outcome.
func (m *Metrics) PaymentCompleted(paid bool) {
	if m == nil {
		return
	}
	outcome := "failed"
	if paid {
		outcome = "success"
	}
	m.payments.WithLabelValues(outcome).Inc()
}

// statusRecorder captures the response status without buffering the body.
type statusRecorder struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if !r.wrote {
		r.status = code
		r.wrote = true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wrote = true
	return r.ResponseWriter.Write(b)
}

// Hijack forwards to the underlying ResponseWriter so WebSocket upgrades (chat
// /ws) keep working when this middleware is in the chain.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying ResponseWriter does not support hijacking")
	}
	return h.Hijack()
}
