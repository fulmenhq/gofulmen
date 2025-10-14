package foundry

import (
	"context"
	"net/http"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

// correlationIDKey is the context key for correlation IDs.
const correlationIDKey contextKey = "correlation_id"

// WithCorrelationID returns a new context with the correlation ID attached.
//
// This is the standard Go pattern for propagating correlation IDs across
// service boundaries and through the call stack.
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    corrID := foundry.NewCorrelationIDValue()
//	    ctx := foundry.WithCorrelationID(r.Context(), corrID)
//
//	    // Pass context to downstream functions
//	    processRequest(ctx, data)
//	}
func WithCorrelationID(ctx context.Context, id CorrelationID) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationIDFromContext extracts the correlation ID from the context.
//
// Returns the correlation ID and true if present, or an empty ID and false
// if not found in the context.
//
// Example:
//
//	func processRequest(ctx context.Context, data interface{}) {
//	    corrID, ok := foundry.CorrelationIDFromContext(ctx)
//	    if !ok {
//	        log.Warn("No correlation ID in context")
//	        corrID = foundry.NewCorrelationIDValue()
//	    }
//
//	    log.Info("Processing request", "correlation_id", corrID)
//	}
func CorrelationIDFromContext(ctx context.Context) (CorrelationID, bool) {
	id, ok := ctx.Value(correlationIDKey).(CorrelationID)
	return id, ok
}

// MustCorrelationIDFromContext extracts the correlation ID or panics.
//
// Use this when the correlation ID is required and its absence indicates
// a programming error.
//
// Example:
//
//	func logEvent(ctx context.Context, event string) {
//	    corrID := foundry.MustCorrelationIDFromContext(ctx)
//	    logger.Info(event, "correlation_id", corrID)
//	}
func MustCorrelationIDFromContext(ctx context.Context) CorrelationID {
	id, ok := CorrelationIDFromContext(ctx)
	if !ok {
		panic("correlation ID not found in context")
	}
	return id
}

// CorrelationIDMiddleware is HTTP middleware that extracts or generates correlation IDs.
//
// This middleware:
//   - Checks for X-Correlation-ID header in incoming request
//   - Validates and uses it if present
//   - Generates a new correlation ID if not present or invalid
//   - Attaches correlation ID to request context
//   - Sets X-Correlation-ID header in response
//
// Example:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/data", handleData)
//
//	// Wrap with correlation ID middleware
//	handler := foundry.CorrelationIDMiddleware(mux)
//	http.ListenAndServe(":8080", handler)
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var corrID CorrelationID

		// Try to extract from X-Correlation-ID header
		headerValue := r.Header.Get("X-Correlation-ID")
		if headerValue != "" {
			parsed, err := ParseCorrelationIDValue(headerValue)
			if err == nil && parsed.IsValid() {
				corrID = parsed
			}
		}

		// Generate new ID if not present or invalid
		if corrID == "" {
			corrID = NewCorrelationIDValue()
		}

		// Set response header
		w.Header().Set("X-Correlation-ID", corrID.String())

		// Attach to context and continue
		ctx := WithCorrelationID(r.Context(), corrID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CorrelationIDHandler is a convenience wrapper for http.HandlerFunc.
//
// This is a shorthand for CorrelationIDMiddleware when you have a function
// rather than an http.Handler.
//
// Example:
//
//	http.Handle("/api/data", foundry.CorrelationIDHandler(handleData))
func CorrelationIDHandler(next http.HandlerFunc) http.Handler {
	return CorrelationIDMiddleware(next)
}
