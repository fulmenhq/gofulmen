package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"github.com/stretchr/testify/assert"
)

// TestChiIntegrationPattern demonstrates HTTP metrics middleware integration pattern with Chi router
// This test shows how to integrate with Chi without requiring Chi as a dependency
func TestChiIntegrationPattern(t *testing.T) {
	emitter := &mockEmitter{}

	// Example Chi router setup (would import "github.com/go-chi/chi/v5")
	/*
				r := chi.NewRouter()

				// Apply HTTP metrics middleware to the entire router
				r.Use(telemetry.HTTPMetricsMiddleware(
					emitter, // Pass emitter first
					telemetry.WithServiceName("chi-api"),
					telemetry.WithRouteNormalizer(func(method, path string) string {
						// Custom route normalization for Chi
						if path == "/api/users/{id}" {
							return "/api/users/{id}"
						}
						return telemetry.DefaultRouteNormalizer(method, path)
					}),
				))

				// Add routes
				r.Get("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					if _, err := w.Write([]byte(`{"id": "123", "name": "John"}`)); err != nil {
				t.Fatal(err)
			}
				})

				r.Post("/api/users", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					if _, err := w.Write([]byte(`{"id": "456", "name": "Jane"}`)); err != nil {
			t.Fatal(err)
		}
				})
	*/

	// For this test, we'll simulate the Chi middleware behavior
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("chi-api"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			// Simulate Chi's route parameter extraction
			if method == "GET" && path == "/api/users/123" {
				return "/api/users/{id}"
			}
			return telemetry.DefaultRouteNormalizer(method, path)
		}),
	)

	// Create a handler that simulates Chi route
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users/123" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"id": "123", "name": "John"}`)); err != nil {
				t.Fatal(err)
			}
		} else if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write([]byte(`{"id": "456", "name": "Jane"}`)); err != nil {
				t.Fatal(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Test GET /api/users/123
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, emitter.calledCounter, "Counter should be called")
	assert.Equal(t, metrics.HTTPRequestsTotal, emitter.counterName)
	assert.Equal(t, float64(1), emitter.counterValue)

	// Verify tags include route template
	assert.Equal(t, "GET", emitter.counterTags[metrics.TagMethod])
	assert.Equal(t, "/api/users/{id}", emitter.counterTags[metrics.TagRoute])
	assert.Equal(t, "200", emitter.counterTags[metrics.TagStatus])
	assert.Equal(t, "chi-api", emitter.counterTags[metrics.TagService])

	// Test POST /api/users
	emitter = &mockEmitter{} // Reset emitter
	middleware = telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("chi-api"),
	)

	handler = middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			if _, err := w.Write([]byte(`{"id": "456", "name": "Jane"}`)); err != nil {
				t.Fatal(err)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	req = httptest.NewRequest("POST", "/api/users", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "POST", emitter.counterTags[metrics.TagMethod])
	assert.Equal(t, "/api/users", emitter.counterTags[metrics.TagRoute])
	assert.Equal(t, "201", emitter.counterTags[metrics.TagStatus])
}

// TestChiMiddlewareOrder demonstrates correct middleware ordering with Chi
func TestChiMiddlewareOrderPattern(t *testing.T) {
	emitter := &mockEmitter{}

	// Example Chi middleware ordering
	/*
			r := chi.NewRouter()

			// Apply HTTP metrics middleware first (to wrap all requests)
			r.Use(telemetry.HTTPMetricsMiddleware(
				emitter,
				telemetry.WithServiceName("chi-ordered"),
			))

			// Apply other middleware (e.g., CORS, logging)
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("X-Custom-Header", "test")
					next.ServeHTTP(w, r)
				})
			})

			r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("ordered")); err != nil {
			t.Fatal(err)
		}
			})
	*/

	// Simulate the middleware chain
	metricsMiddleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("chi-ordered"),
	)

	customMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "test")
			next.ServeHTTP(w, r)
		})
	}

	// Apply in correct order: metrics first, then custom
	handler := customMiddleware(metricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ordered")); err != nil {
			t.Fatal(err)
		}
	})))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test", w.Header().Get("X-Custom-Header"))
	assert.True(t, emitter.calledCounter, "Metrics should be collected")
	assert.Equal(t, "chi-ordered", emitter.counterTags[metrics.TagService])
}
