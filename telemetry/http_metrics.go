package telemetry

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// RouteNormalizer normalizes HTTP routes to prevent cardinality explosions
type RouteNormalizer func(method, path string) string

// HTTPMetricsConfig configures the HTTP metrics middleware
type HTTPMetricsConfig struct {
	RouteNormalizer RouteNormalizer
	SizeBuckets     []float64
	ServiceName     string
}

// Default HTTP metric buckets from taxonomy
var (
	DefaultHTTPSizeBuckets = []float64{1024, 10240, 102400, 1048576, 10485760, 104857600} // bytes
)

// HTTPMetricsOption configures the HTTP metrics middleware
type HTTPMetricsOption func(*HTTPMetricsConfig)

// WithRouteNormalizer sets a custom route normalizer function
func WithRouteNormalizer(fn RouteNormalizer) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.RouteNormalizer = fn
	}
}

// WithCustomSizeBuckets sets custom histogram buckets for size metrics
func WithCustomSizeBuckets(sizeBuckets []float64) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.SizeBuckets = sizeBuckets
	}
}

// WithServiceName sets the service name for metrics labels
func WithServiceName(name string) HTTPMetricsOption {
	return func(c *HTTPMetricsConfig) {
		c.ServiceName = name
	}
}

// Pre-compiled UUID pattern for better performance (global to avoid recompilation)
var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// DefaultRouteNormalizer provides optimized defaults for route normalization
func DefaultRouteNormalizer(method, path string) string {
	// Fast path: if no query or fragment, skip string operations
	idx := strings.IndexAny(path, "?#")
	if idx >= 0 {
		// Strip query parameters and fragments to prevent cardinality explosion
		path = path[:idx]
	}

	// Fast UUID detection using length and pattern matching
	if len(path) > 30 && strings.Contains(path, "-") {
		// Replace all UUID segments with {uuid} using regex replacement
		path = uuidPattern.ReplaceAllString(path, "{uuid}")
	}

	// Optimized numeric segment detection
	segments := strings.Split(strings.Trim(path, "/"), "/")
	needsNormalization := false
	for i, segment := range segments {
		if len(segment) > 0 && segment[0] >= '0' && segment[0] <= '9' {
			// Likely numeric, but verify to avoid false positives
			if _, err := strconv.Atoi(segment); err == nil {
				segments[i] = "{id}"
				needsNormalization = true
			}
		}
	}

	if needsNormalization {
		return "/" + strings.Join(segments, "/")
	}

	return path
}

// Histogram bucket pool to reduce allocations
var histogramBucketPool = sync.Pool{
	New: func() interface{} {
		buckets := make([]HistogramBucket, 0, 7) // DefaultHTTPSizeBuckets + 1
		return &buckets
	},
}

// This is a workaround until MetricsEmitter supports SizeHistogram method
func emitSizeHistogram(emitter MetricsEmitter, name string, size float64, buckets []float64, tags map[string]string) error {
	if len(buckets) == 0 {
		buckets = DefaultHTTPSizeBuckets
	}

	// Get bucket slice from pool to reduce allocations
	bucketCountsPtr := histogramBucketPool.Get().(*[]HistogramBucket)
	bucketCounts := *bucketCountsPtr
	defer func() {
		// Reset length and return to pool
		*bucketCountsPtr = bucketCounts[:0]
		histogramBucketPool.Put(bucketCountsPtr)
	}()

	// Create histogram buckets for this single observation with proper cumulative counts
	// For a single observation: 0 for buckets < size, 1 for first bucket ≥ size, 1 for +Inf
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

	// Pre-allocated tag maps to reduce allocations (sync.Map for concurrent safety)
	tagPool sync.Pool
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

	// Get tag map from pool to reduce allocations
	tags := h.getTagMap()
	defer h.putTagMap(tags)

	// Fill tags efficiently (pre-allocate status string to avoid allocation)
	statusStr := strconv.Itoa(rw.status)
	tags[metrics.TagMethod] = r.Method
	tags[metrics.TagRoute] = route
	tags[metrics.TagStatus] = statusStr
	if h.config.ServiceName != "" {
		tags[metrics.TagService] = h.config.ServiceName
	}

	// HTTP requests total (counter)
	_ = h.emitter.Counter(metrics.HTTPRequestsTotal, 1, tags) // Ignore errors to avoid request failure

	// HTTP request duration (histogram) - duration is already in nanoseconds for the interface
	_ = h.emitter.Histogram(metrics.HTTPRequestDurationSeconds, duration, tags) // Ignore errors to avoid request failure

	// HTTP request size (histogram) - always emit, use 0 if unknown
	requestSizeValue := float64(requestSize)
	if requestSize < 0 {
		requestSizeValue = 0 // Unknown size becomes 0
	}
	_ = emitSizeHistogram(h.emitter, metrics.HTTPRequestSizeBytes, requestSizeValue, h.config.SizeBuckets, tags) // Ignore errors to avoid request failure

	// HTTP response size (histogram) - always emit
	_ = emitSizeHistogram(h.emitter, metrics.HTTPResponseSizeBytes, float64(responseSize), h.config.SizeBuckets, tags) // Ignore errors to avoid request failure

	// HTTP active requests (gauge) - minimal tags per taxonomy (service only)
	if h.config.ServiceName != "" {
		// Reuse tags map for gauge with just service tag
		serviceOnlyTag := h.getTagMap()
		defer h.putTagMap(serviceOnlyTag)
		serviceOnlyTag[metrics.TagService] = h.config.ServiceName
		_ = h.emitter.Gauge(metrics.HTTPActiveRequests, float64(h.active), serviceOnlyTag)
	}
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

// getTagMap retrieves a tag map from the pool or creates a new one
func (h *httpMetricsHandler) getTagMap() map[string]string {
	if tags := h.tagPool.Get(); tags != nil {
		return tags.(map[string]string)
	}
	return make(map[string]string, 4) // Pre-allocate for common tags
}

// putTagMap returns a tag map to the pool after clearing it
func (h *httpMetricsHandler) putTagMap(tags map[string]string) {
	// Clear the map to avoid memory leaks
	for k := range tags {
		delete(tags, k)
	}
	h.tagPool.Put(tags)
}

// HTTPMetricsMiddleware creates middleware that collects HTTP server metrics
func HTTPMetricsMiddleware(emitter MetricsEmitter, opts ...HTTPMetricsOption) func(http.Handler) http.Handler {
	config := HTTPMetricsConfig{
		RouteNormalizer: DefaultRouteNormalizer,
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
			tagPool: sync.Pool{
				New: func() interface{} {
					return make(map[string]string, 4)
				},
			},
		}
	}
}
