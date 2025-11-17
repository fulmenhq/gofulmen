package telemetry

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// RouteNormalizer normalizes HTTP routes to prevent cardinality explosions
type RouteNormalizer func(method, path string) string

// HTTPMetricsConfig configures the HTTP metrics middleware
type HTTPMetricsConfig struct {
	RouteNormalizer RouteNormalizer
	DurationBuckets []float64
	SizeBuckets     []float64
	ServiceName     string
}

// Default HTTP metric buckets from taxonomy
var (
	DefaultHTTPDurationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10} // seconds
	DefaultHTTPSizeBuckets     = []float64{1024, 10240, 102400, 1048576, 10485760, 104857600}       // bytes
)

// HTTPMetricsOption configures the HTTP metrics middleware
type HTTPMetricsOption func(*HTTPMetricsConfig)

// WithRouteNormalizer sets a custom route normalizer function
func WithRouteNormalizer(fn RouteNormalizer) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.RouteNormalizer = fn
	}
}

// WithCustomBuckets sets custom histogram buckets for duration and size metrics
func WithCustomBuckets(durationBuckets, sizeBuckets []float64) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.DurationBuckets = durationBuckets
		c.SizeBuckets = sizeBuckets
	}
}

// WithServiceName sets the service name for metrics labels
func WithServiceName(name string) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.ServiceName = name
	}
}

// DefaultRouteNormalizer provides basic route normalization by stripping numeric segments, UUIDs, and query parameters
func DefaultRouteNormalizer(method, path string) string {
	// Strip query parameters and fragments to prevent cardinality explosion
	if idx := strings.IndexAny(path, "?#"); idx >= 0 {
		path = path[:idx]
	}

	// Split path into segments
	segments := strings.Split(strings.Trim(path, "/"), "/")
	normalized := make([]string, 0, len(segments))

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// Replace numeric segments with {id}
		if _, err := strconv.Atoi(segment); err == nil {
			normalized = append(normalized, "{id}")
			continue
		}

		// Replace UUID-like segments (8-4-4-4-12 hex chars) with {uuid}
		if len(segment) == 36 && strings.Count(segment, "-") == 4 {
			parts := strings.Split(segment, "-")
			if len(parts) == 5 && len(parts[0]) == 8 && len(parts[1]) == 4 && len(parts[2]) == 4 && len(parts[3]) == 4 && len(parts[4]) == 12 {
				if isHexString(parts[0]) && isHexString(parts[1]) && isHexString(parts[2]) && isHexString(parts[3]) && isHexString(parts[4]) {
					normalized = append(normalized, "{uuid}")
					continue
				}
			}
		}

		normalized = append(normalized, segment)
	}

	result := "/" + strings.Join(normalized, "/")
	if result == "//" {
		return "/"
	}
	return result
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// emitSizeHistogram emits a size-based histogram using HistogramSummary
// This is a workaround until MetricsEmitter supports SizeHistogram method
func emitSizeHistogram(emitter MetricsEmitter, name string, size float64, buckets []float64, tags map[string]string) error {
	if len(buckets) == 0 {
		buckets = DefaultHTTPSizeBuckets
	}

	// Create histogram buckets for this single observation with proper cumulative counts
	// For a single observation: 0 for buckets < size, 1 for first bucket ≥ size, 1 for +Inf
	bucketCounts := make([]HistogramBucket, 0, len(buckets)+1)
	observationFound := false

	for _, bucket := range buckets {
		// Once we find the bucket that contains our observation, all subsequent buckets should have count=1
		if !observationFound {
			if size <= bucket {
				// This is the first bucket that contains our observation
				bucketCounts = append(bucketCounts, HistogramBucket{
					LE:    bucket,
					Count: 1,
				})
				observationFound = true
			} else {
				// This bucket is too small for our observation
				bucketCounts = append(bucketCounts, HistogramBucket{
					LE:    bucket,
					Count: 0,
				})
			}
		} else {
			// We've already found the containing bucket, so this is cumulative
			bucketCounts = append(bucketCounts, HistogramBucket{
				LE:    bucket,
				Count: 1,
			})
		}
	}

	// Add +Inf bucket - always 1 for single observation (cumulative)
	bucketCounts = append(bucketCounts, HistogramBucket{
		LE:    math.Inf(1),
		Count: 1,
	})

	summary := HistogramSummary{
		Count:   1,
		Sum:     size,
		Buckets: bucketCounts,
	}

	return emitter.HistogramSummary(name, summary, tags)
}

// httpMetricsHandler wraps an HTTP handler to collect metrics
type httpMetricsHandler struct {
	handler http.Handler
	emitter MetricsEmitter
	config  HTTPMetricsConfig
	active  int64 // Active request counter - must be updated atomically
}

// ServeHTTP implements http.Handler
func (h *httpMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Normalize route
	route := h.config.RouteNormalizer(r.Method, r.URL.Path)

	// Track active requests (increment atomically)
	atomic.AddInt64(&h.active, 1)

	// Create response writer wrapper to capture status and size
	rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

	// Decrement active requests when done (atomically)
	defer func() {
		atomic.AddInt64(&h.active, -1)
	}()

	// Call the handler
	h.handler.ServeHTTP(rw, r)

	// Calculate metrics
	duration := time.Since(start)
	requestSize := int64(r.ContentLength) // May be -1 for unknown
	responseSize := rw.size

	// Emit metrics
	tags := map[string]string{
		metrics.TagMethod: r.Method,
		metrics.TagRoute:  route,
		metrics.TagStatus: strconv.Itoa(rw.status),
	}

	if h.config.ServiceName != "" {
		tags[metrics.TagService] = h.config.ServiceName
	}

	// HTTP requests total (counter)
	h.emitter.Counter(metrics.HTTPRequestsTotal, 1, tags)

	// HTTP request duration (histogram) - duration is already in nanoseconds for the interface
	h.emitter.Histogram(metrics.HTTPRequestDurationSeconds, duration, tags)

	// HTTP request size (histogram) - always emit, use 0 if unknown
	requestSizeValue := float64(requestSize)
	if requestSize < 0 {
		requestSizeValue = 0 // Unknown size becomes 0
	}
	emitSizeHistogram(h.emitter, metrics.HTTPRequestSizeBytes, requestSizeValue, h.config.SizeBuckets, tags)

	// HTTP response size (histogram) - always emit
	emitSizeHistogram(h.emitter, metrics.HTTPResponseSizeBytes, float64(responseSize), h.config.SizeBuckets, tags)

	// HTTP active requests (gauge) - minimal tags per taxonomy (service only)
	activeTags := make(map[string]string)
	if h.config.ServiceName != "" {
		activeTags[metrics.TagService] = h.config.ServiceName
	}

	h.emitter.Gauge(metrics.HTTPActiveRequests, float64(h.active), activeTags)
}

// responseWriter wraps http.ResponseWriter to capture status and response size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.size += int64(n)
	return n, err
}

// HTTPMetricsMiddleware creates middleware that collects HTTP server metrics
func HTTPMetricsMiddleware(emitter MetricsEmitter, opts ...HTTPMetricsOption) func(http.Handler) http.Handler {
	config := HTTPMetricsConfig{
		RouteNormalizer: DefaultRouteNormalizer,
		DurationBuckets: DefaultHTTPDurationBuckets,
		SizeBuckets:     DefaultHTTPSizeBuckets,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return func(next http.Handler) http.Handler {
		return &httpMetricsHandler{
			handler: next,
			emitter: emitter,
			config:  config,
		}
	}
}
