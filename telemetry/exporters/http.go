package exporters

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"golang.org/x/time/rate"
)

// httpHandler wraps the Prometheus exporter with HTTP middleware
type httpHandler struct {
	exporter  *PrometheusExporter
	config    *PrometheusConfig
	limiter   *rate.Limiter
	quietMode atomic.Bool
}

// newHTTPHandler creates an HTTP handler with auth and rate limiting
func newHTTPHandler(exporter *PrometheusExporter, config *PrometheusConfig) *httpHandler {
	h := &httpHandler{
		exporter: exporter,
		config:   config,
	}

	// Initialize rate limiter if configured
	if config.RateLimitPerMinute > 0 {
		// Convert per-minute to per-second rate
		perSecond := float64(config.RateLimitPerMinute) / 60.0
		h.limiter = rate.NewLimiter(rate.Limit(perSecond), config.RateLimitBurst)
	}

	h.quietMode.Store(config.QuietMode)
	return h
}

// ServeHTTP implements http.Handler with auth, rate limiting, and metrics
func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	path := r.URL.Path

	// Extract client identifier (from auth header or IP)
	client := h.getClientIdentifier(r)

	tags := map[string]string{
		metrics.TagPath:   path,
		metrics.TagClient: client,
	}

	// Check bearer token authentication if configured
	if h.config.BearerToken != "" {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + h.config.BearerToken

		if authHeader != expectedAuth {
			h.emitHTTPError(w, http.StatusUnauthorized, "Unauthorized", tags)
			return
		}
	}

	// Apply rate limiting if configured
	if h.limiter != nil {
		if !h.limiter.Allow() {
			h.emitHTTPError(w, http.StatusTooManyRequests, "Rate limit exceeded", tags)
			return
		}
	}

	// Emit HTTP request metric
	tags[metrics.TagStatus] = fmt.Sprintf("%d", http.StatusOK)
	telemetry.EmitCounter(metrics.PrometheusExporterHTTPRequestsTotal, 1, tags)

	// Log request if not in quiet mode
	if !h.quietMode.Load() {
		fmt.Printf("[prometheus-exporter] %s %s from %s\n", r.Method, path, client)
	}

	// Serve metrics (delegate to the exporter's metrics handler)
	h.exporter.metricsHandler(w, r)

	// Measure total request duration
	_ = time.Since(start) // Reserved for future use
}

// emitHTTPError emits error metrics and sends HTTP error response
func (h *httpHandler) emitHTTPError(w http.ResponseWriter, statusCode int, message string, tags map[string]string) {
	tags[metrics.TagStatus] = fmt.Sprintf("%d", statusCode)

	// Emit both HTTP request and error counters
	telemetry.EmitCounter(metrics.PrometheusExporterHTTPRequestsTotal, 1, tags)
	telemetry.EmitCounter(metrics.PrometheusExporterHTTPErrorsTotal, 1, tags)

	// Log error if not in quiet mode
	if !h.quietMode.Load() {
		fmt.Printf("[prometheus-exporter] %d %s\n", statusCode, message)
	}

	http.Error(w, message, statusCode)
}

// getClientIdentifier extracts a client identifier from the request
func (h *httpHandler) getClientIdentifier(r *http.Request) string {
	// Try X-Forwarded-For first (proxy environments)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// SetQuietMode enables or disables request logging
func (e *PrometheusExporter) SetQuietMode(quiet bool) {
	if e.httpHandler != nil {
		e.httpHandler.quietMode.Store(quiet)
	}
}
