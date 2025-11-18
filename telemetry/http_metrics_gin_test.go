package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	"github.com/stretchr/testify/assert"
)

// TestGinIntegrationPattern demonstrates HTTP metrics middleware integration pattern with Gin router
// This test shows how to integrate with Gin without requiring Gin as a dependency
func TestGinIntegrationPattern(t *testing.T) {
	emitter := &mockEmitter{}

	// Example Gin router setup (would import "github.com/gin-gonic/gin")
	/*
		r := gin.Default()

		// Apply HTTP metrics middleware to the entire router
		r.Use(telemetry.HTTPMetricsMiddleware(
			emitter, // Pass emitter first
			telemetry.WithServiceName("gin-api"),
			telemetry.WithRouteNormalizer(func(method, path string) string {
				// Custom route normalization for Gin
				// Gin uses :param syntax, so we convert to {param} for consistency
				if strings.Contains(path, ":id") {
					return strings.Replace(path, ":id", "{id}", 1)
				}
				return telemetry.DefaultRouteNormalizer(method, path)
			}),
		))

		// Add routes
		r.GET("/api/users/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "name": "John"})
		})

		r.POST("/api/users", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "456", "name": "Jane"})
		})
	*/

	// For this test, we'll simulate the Gin middleware behavior
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("gin-api"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			// Simulate Gin's route parameter extraction
			if method == "GET" && path == "/api/users/123" {
				return "/api/users/{id}"
			}
			return telemetry.DefaultRouteNormalizer(method, path)
		}),
	)

	// Create a handler that simulates Gin route
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users/123" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"id": "123", "name": "John"}`)); err != nil {
				t.Fatal(err)
			}
		} else if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
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
	assert.Equal(t, "gin-api", emitter.counterTags[metrics.TagService])

	// Test POST /api/users
	emitter = &mockEmitter{} // Reset emitter
	middleware = telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("gin-api"),
	)

	handler = middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
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

// TestGinMiddlewareOrder demonstrates correct middleware ordering with Gin
func TestGinMiddlewareOrderPattern(t *testing.T) {
	emitter := &mockEmitter{}

	// Example Gin middleware ordering
	/*
		r := gin.Default()

		// Apply HTTP metrics middleware first (to wrap all requests)
		r.Use(telemetry.HTTPMetricsMiddleware(
			emitter,
			telemetry.WithServiceName("gin-ordered"),
		))

		// Apply other middleware (e.g., CORS, logging)
		r.Use(func(c *gin.Context) {
			c.Header("X-Custom-Header", "test")
			c.Next()
		})

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ordered")
		})
	*/

	// Simulate the middleware chain
	metricsMiddleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("gin-ordered"),
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
	assert.Equal(t, "gin-ordered", emitter.counterTags[metrics.TagService])
}

// TestGinRouteNormalization demonstrates Gin-specific route normalization
func TestGinRouteNormalizationPattern(t *testing.T) {
	emitter := &mockEmitter{}

	// Gin uses :param syntax, so we need to normalize to {param} for consistency
	middleware := telemetry.HTTPMetricsMiddleware(
		emitter,
		telemetry.WithServiceName("gin-api"),
		telemetry.WithRouteNormalizer(func(method, path string) string {
			// Convert Gin's :param to {param} for consistency
			if strings.Contains(path, ":") {
				// Simple conversion for demonstration
				if strings.Contains(path, ":id") {
					return strings.Replace(path, ":id", "{id}", 1)
				}
				if strings.Contains(path, ":userId") {
					return strings.Replace(path, ":userId", "{userId}", 1)
				}
			}
			return telemetry.DefaultRouteNormalizer(method, path)
		}),
	)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Fatal(err)
		}
	}))

	// Test route with parameter
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, "/api/users/{id}", emitter.counterTags[metrics.TagRoute])
	assert.Equal(t, "gin-api", emitter.counterTags[metrics.TagService])
}
