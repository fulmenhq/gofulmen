package logging

import (
	"github.com/fulmenhq/gofulmen/foundry"
)

// CorrelationMiddleware auto-generates or preserves correlation IDs for distributed tracing.
//
// This middleware ensures every log event has a UUIDv7 correlation ID for cross-service
// request tracking. If the event already has a correlation ID, it is preserved unless
// validateExisting is enabled and the ID is invalid.
type CorrelationMiddleware struct {
	order            int
	validateExisting bool
}

// NewCorrelationMiddleware creates a new correlation middleware instance.
//
// The middleware auto-registers with DefaultRegistry during package initialization.
// Default execution order is 5 (early in pipeline, before redaction/throttling).
//
// Config options:
//   - order (int): Execution order in pipeline (default: 5)
//   - validateExisting (bool): Validate existing IDs are UUIDv7 (default: false)
func NewCorrelationMiddleware(config map[string]any) (Middleware, error) {
	order := 5
	if configOrder, ok := config["order"].(int); ok {
		order = configOrder
	} else if configOrder, ok := config["order"].(float64); ok {
		order = int(configOrder)
	}

	validateExisting := false
	if configValidate, ok := config["validateExisting"].(bool); ok {
		validateExisting = configValidate
	}

	return &CorrelationMiddleware{
		order:            order,
		validateExisting: validateExisting,
	}, nil
}

// Process generates correlation ID if missing, preserves existing ID.
//
// Uses foundry.GenerateCorrelationID() to create time-sortable UUIDv7
// for optimal log aggregation and cross-service tracing.
//
// If validateExisting is enabled, invalid existing IDs are replaced with
// new UUIDv7 to ensure strict UUIDv7 compliance.
func (m *CorrelationMiddleware) Process(event *LogEvent) *LogEvent {
	if event == nil {
		return nil
	}

	if event.CorrelationID == "" {
		event.CorrelationID = foundry.GenerateCorrelationID()
	} else if m.validateExisting && !foundry.IsValidCorrelationID(event.CorrelationID) {
		event.CorrelationID = foundry.GenerateCorrelationID()
	}

	return event
}

// Order returns the execution order within the middleware pipeline.
//
// Default is 5 (early processing). Lower values execute first.
func (m *CorrelationMiddleware) Order() int {
	return m.order
}

// Name returns the middleware identifier for registry lookup.
func (m *CorrelationMiddleware) Name() string {
	return "correlation"
}

func init() {
	DefaultRegistry().Register("correlation", NewCorrelationMiddleware)
}
