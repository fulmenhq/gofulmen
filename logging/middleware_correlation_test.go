package logging

import (
	"testing"

	"github.com/fulmenhq/gofulmen/foundry"
)

func TestCorrelationMiddleware_GeneratesID(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test-service",
		Message: "test message",
	}

	result := middleware.Process(event)

	if result.CorrelationID == "" {
		t.Error("Expected correlation ID to be generated, got empty string")
	}

	if !foundry.IsValidCorrelationID(result.CorrelationID) {
		t.Errorf("Expected valid UUIDv7 correlation ID, got %q", result.CorrelationID)
	}
}

func TestCorrelationMiddleware_PreservesExistingID(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	existingID := foundry.GenerateCorrelationID()
	event := &LogEvent{
		Service:       "test-service",
		Message:       "test message",
		CorrelationID: existingID,
	}

	result := middleware.Process(event)

	if result.CorrelationID != existingID {
		t.Errorf("Expected correlation ID to be preserved as %q, got %q", existingID, result.CorrelationID)
	}
}

func TestCorrelationMiddleware_NilEvent(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	result := middleware.Process(nil)

	if result != nil {
		t.Error("Expected nil event to be passed through, got non-nil")
	}
}

func TestCorrelationMiddleware_Order(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]any
		expectedOrder int
	}{
		{
			name:          "default order",
			config:        map[string]any{},
			expectedOrder: 5,
		},
		{
			name:          "custom order int",
			config:        map[string]any{"order": 10},
			expectedOrder: 10,
		},
		{
			name:          "custom order float64",
			config:        map[string]any{"order": 15.0},
			expectedOrder: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := NewCorrelationMiddleware(tt.config)
			if err != nil {
				t.Fatalf("NewCorrelationMiddleware failed: %v", err)
			}

			if middleware.Order() != tt.expectedOrder {
				t.Errorf("Expected order %d, got %d", tt.expectedOrder, middleware.Order())
			}
		})
	}
}

func TestCorrelationMiddleware_Name(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	if middleware.Name() != "correlation" {
		t.Errorf("Expected name 'correlation', got %q", middleware.Name())
	}
}

func TestCorrelationMiddleware_Registration(t *testing.T) {
	factory := DefaultRegistry().factories["correlation"]
	if factory == nil {
		t.Fatal("Expected correlation middleware to be registered in DefaultRegistry")
	}

	middleware, err := factory(map[string]any{})
	if err != nil {
		t.Fatalf("Factory failed to create middleware: %v", err)
	}

	if middleware.Name() != "correlation" {
		t.Errorf("Expected factory to create correlation middleware, got %q", middleware.Name())
	}
}

func TestCorrelationMiddleware_IDUniqueness(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	event1 := &LogEvent{Service: "test", Message: "msg1"}
	event2 := &LogEvent{Service: "test", Message: "msg2"}

	result1 := middleware.Process(event1)
	result2 := middleware.Process(event2)

	if result1.CorrelationID == result2.CorrelationID {
		t.Error("Expected unique correlation IDs for different events")
	}
}

func TestCorrelationMiddleware_MutationSafety(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test-service",
		Message: "test message",
	}

	result := middleware.Process(event)

	if result != event {
		t.Error("Expected middleware to modify event in-place (same pointer)")
	}

	if event.CorrelationID == "" {
		t.Error("Expected original event to be modified with correlation ID")
	}
}

func TestCorrelationMiddleware_ValidationMode(t *testing.T) {
	tests := []struct {
		name            string
		config          map[string]any
		existingID      string
		expectGenerated bool
		expectPreserved bool
	}{
		{
			name:            "validation disabled, invalid ID preserved",
			config:          map[string]any{"validateExisting": false},
			existingID:      "not-a-valid-uuid",
			expectPreserved: true,
		},
		{
			name:            "validation enabled, invalid ID replaced",
			config:          map[string]any{"validateExisting": true},
			existingID:      "not-a-valid-uuid",
			expectGenerated: true,
		},
		{
			name:            "validation enabled, valid UUIDv7 preserved",
			config:          map[string]any{"validateExisting": true},
			existingID:      foundry.GenerateCorrelationID(),
			expectPreserved: true,
		},
		{
			name:            "validation disabled by default",
			config:          map[string]any{},
			existingID:      "invalid-but-preserved",
			expectPreserved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware, err := NewCorrelationMiddleware(tt.config)
			if err != nil {
				t.Fatalf("NewCorrelationMiddleware failed: %v", err)
			}

			event := &LogEvent{
				Service:       "test-service",
				Message:       "test",
				CorrelationID: tt.existingID,
			}

			result := middleware.Process(event)

			if tt.expectPreserved {
				if result.CorrelationID != tt.existingID {
					t.Errorf("Expected ID to be preserved as %q, got %q", tt.existingID, result.CorrelationID)
				}
			}

			if tt.expectGenerated {
				if result.CorrelationID == tt.existingID {
					t.Error("Expected invalid ID to be replaced with generated UUIDv7")
				}
				if !foundry.IsValidCorrelationID(result.CorrelationID) {
					t.Errorf("Expected valid UUIDv7 after replacement, got %q", result.CorrelationID)
				}
			}
		})
	}
}

func TestCorrelationMiddleware_ValidateExistingConfig(t *testing.T) {
	middleware, err := NewCorrelationMiddleware(map[string]any{
		"validateExisting": true,
		"order":            10,
	})
	if err != nil {
		t.Fatalf("NewCorrelationMiddleware failed: %v", err)
	}

	cm := middleware.(*CorrelationMiddleware)
	if !cm.validateExisting {
		t.Error("Expected validateExisting to be true")
	}

	if cm.order != 10 {
		t.Errorf("Expected order 10, got %d", cm.order)
	}
}
