package foundry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestWithCorrelationID tests attaching correlation ID to context
func TestWithCorrelationID(t *testing.T) {
	ctx := context.Background()
	id := NewCorrelationIDValue()

	newCtx := WithCorrelationID(ctx, id)

	// Verify context is not nil
	if newCtx == nil {
		t.Fatal("WithCorrelationID() returned nil context")
	}

	// Verify correlation ID can be extracted
	extracted, ok := CorrelationIDFromContext(newCtx)
	if !ok {
		t.Fatal("Failed to extract correlation ID from context")
	}

	if extracted.String() != id.String() {
		t.Errorf("Extracted ID = %q, want %q", extracted, id)
	}
}

// TestCorrelationIDFromContext tests extracting correlation ID from context
func TestCorrelationIDFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setupCtx  func() context.Context
		wantOK    bool
		wantEmpty bool
	}{
		{
			name: "WithID",
			setupCtx: func() context.Context {
				return WithCorrelationID(context.Background(), NewCorrelationIDValue())
			},
			wantOK:    true,
			wantEmpty: false,
		},
		{
			name: "WithoutID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantOK:    false,
			wantEmpty: true,
		},
		{
			name: "WithEmptyID",
			setupCtx: func() context.Context {
				return WithCorrelationID(context.Background(), "")
			},
			wantOK:    true,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			id, ok := CorrelationIDFromContext(ctx)

			if ok != tt.wantOK {
				t.Errorf("CorrelationIDFromContext() ok = %v, want %v", ok, tt.wantOK)
			}

			isEmpty := id == ""
			if isEmpty != tt.wantEmpty {
				t.Errorf("CorrelationIDFromContext() empty = %v, want %v", isEmpty, tt.wantEmpty)
			}
		})
	}
}

// TestMustCorrelationIDFromContext tests the Must variant
func TestMustCorrelationIDFromContext(t *testing.T) {
	t.Run("WithID", func(t *testing.T) {
		id := NewCorrelationIDValue()
		ctx := WithCorrelationID(context.Background(), id)

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustCorrelationIDFromContext() panicked: %v", r)
			}
		}()

		extracted := MustCorrelationIDFromContext(ctx)
		if extracted.String() != id.String() {
			t.Errorf("MustCorrelationIDFromContext() = %q, want %q", extracted, id)
		}
	})

	t.Run("WithoutID", func(t *testing.T) {
		ctx := context.Background()

		// Should panic
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustCorrelationIDFromContext() should panic when ID not present")
			}
		}()

		MustCorrelationIDFromContext(ctx)
	})
}

// TestCorrelationIDMiddleware_NewID tests middleware generates new ID
func TestCorrelationIDMiddleware_NewID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correlation ID is in context
		id, ok := CorrelationIDFromContext(r.Context())
		if !ok {
			t.Error("Correlation ID not found in context")
			return
		}

		if !id.IsValid() {
			t.Error("Correlation ID is invalid")
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Verify response header is set
	responseID := rec.Header().Get("X-Correlation-ID")
	if responseID == "" {
		t.Error("X-Correlation-ID header not set in response")
	}

	if !IsValidCorrelationID(responseID) {
		t.Errorf("Invalid correlation ID in response header: %s", responseID)
	}
}

// TestCorrelationIDMiddleware_ExistingID tests middleware uses existing ID from header
func TestCorrelationIDMiddleware_ExistingID(t *testing.T) {
	existingID := NewCorrelationIDValue()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify same correlation ID is in context
		id, ok := CorrelationIDFromContext(r.Context())
		if !ok {
			t.Error("Correlation ID not found in context")
			return
		}

		if id.String() != existingID.String() {
			t.Errorf("Context ID = %q, want %q", id, existingID)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", existingID.String())
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Verify same ID in response header
	responseID := rec.Header().Get("X-Correlation-ID")
	if responseID != existingID.String() {
		t.Errorf("Response ID = %q, want %q", responseID, existingID)
	}
}

// TestCorrelationIDMiddleware_InvalidID tests middleware generates new ID when existing is invalid
func TestCorrelationIDMiddleware_InvalidID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify a valid correlation ID is in context (not the invalid one)
		id, ok := CorrelationIDFromContext(r.Context())
		if !ok {
			t.Error("Correlation ID not found in context")
			return
		}

		if !id.IsValid() {
			t.Error("Correlation ID is invalid")
		}

		if id.String() == "invalid-id" {
			t.Error("Middleware should have generated new ID, not used invalid one")
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", "invalid-id")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Verify new valid ID in response header
	responseID := rec.Header().Get("X-Correlation-ID")
	if responseID == "invalid-id" {
		t.Error("Middleware should have generated new ID, not used invalid one")
	}

	if !IsValidCorrelationID(responseID) {
		t.Errorf("Invalid correlation ID in response header: %s", responseID)
	}
}

// TestCorrelationIDMiddleware_PropagatesContext tests context is properly propagated
func TestCorrelationIDMiddleware_PropagatesContext(t *testing.T) {
	var capturedID CorrelationID

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate calling downstream function
		processRequest(r.Context(), &capturedID)
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Verify downstream function received correlation ID
	if capturedID == "" {
		t.Error("Downstream function did not receive correlation ID")
	}

	if !capturedID.IsValid() {
		t.Error("Downstream function received invalid correlation ID")
	}

	// Verify it matches response header
	responseID := rec.Header().Get("X-Correlation-ID")
	if capturedID.String() != responseID {
		t.Errorf("Downstream ID = %q, response header = %q", capturedID, responseID)
	}
}

// Helper function for testing context propagation
func processRequest(ctx context.Context, capturedID *CorrelationID) {
	id, ok := CorrelationIDFromContext(ctx)
	if ok {
		*capturedID = id
	}
}

// TestCorrelationIDHandler tests the HandlerFunc wrapper
func TestCorrelationIDHandler(t *testing.T) {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		id, ok := CorrelationIDFromContext(r.Context())
		if !ok {
			t.Error("Correlation ID not found in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !id.IsValid() {
			t.Error("Correlation ID is invalid")
		}

		w.WriteHeader(http.StatusOK)
	}

	handler := CorrelationIDHandler(handlerFunc)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Handler returned %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify response header
	responseID := rec.Header().Get("X-Correlation-ID")
	if responseID == "" {
		t.Error("X-Correlation-ID header not set")
	}
}

// TestCorrelationIDMiddleware_ChainedHandlers tests middleware with chained handlers
func TestCorrelationIDMiddleware_ChainedHandlers(t *testing.T) {
	var ids []CorrelationID

	// Create second handler first
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := CorrelationIDFromContext(r.Context())
		if ok {
			ids = append(ids, id)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create first handler that calls handler2
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := CorrelationIDFromContext(r.Context())
		if ok {
			ids = append(ids, id)
		}
		// Call next handler
		handler2.ServeHTTP(w, r)
	})

	middleware := CorrelationIDMiddleware(handler1)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	// Verify both handlers received same correlation ID
	if len(ids) != 2 {
		t.Fatalf("Expected 2 IDs, got %d", len(ids))
	}

	if ids[0].String() != ids[1].String() {
		t.Errorf("Handler IDs don't match: %q != %q", ids[0], ids[1])
	}
}

// TestCorrelationIDMiddleware_Integration demonstrates real-world usage
func TestCorrelationIDMiddleware_Integration(t *testing.T) {
	// Simulate a service with correlation ID middleware
	mux := http.NewServeMux()

	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// Extract correlation ID for logging
		corrID, ok := CorrelationIDFromContext(r.Context())
		if !ok {
			t.Error("No correlation ID in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Simulate logging with correlation ID
		t.Logf("Processing request: correlation_id=%s", corrID)

		// Simulate calling downstream service with correlation ID
		req, _ := http.NewRequest("GET", "http://downstream/api", nil)
		req.Header.Set("X-Correlation-ID", corrID.String())

		// Response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with middleware
	handler := CorrelationIDMiddleware(mux)

	// Test 1: Request without correlation ID
	t.Run("NewRequest", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
		}

		responseID := rec.Header().Get("X-Correlation-ID")
		if responseID == "" {
			t.Error("No correlation ID in response")
		}
	})

	// Test 2: Request with existing correlation ID
	t.Run("ExistingID", func(t *testing.T) {
		existingID := NewCorrelationIDValue()

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("X-Correlation-ID", existingID.String())
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
		}

		responseID := rec.Header().Get("X-Correlation-ID")
		if responseID != existingID.String() {
			t.Errorf("Response ID = %q, want %q", responseID, existingID)
		}
	})
}

// Benchmarks

func BenchmarkWithCorrelationID(b *testing.B) {
	ctx := context.Background()
	id := NewCorrelationIDValue()

	for i := 0; i < b.N; i++ {
		WithCorrelationID(ctx, id)
	}
}

func BenchmarkCorrelationIDFromContext(b *testing.B) {
	id := NewCorrelationIDValue()
	ctx := WithCorrelationID(context.Background(), id)

	for i := 0; i < b.N; i++ {
		CorrelationIDFromContext(ctx)
	}
}

func BenchmarkCorrelationIDMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		CorrelationIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := CorrelationIDMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		middleware.ServeHTTP(rec, req)
	}
}
