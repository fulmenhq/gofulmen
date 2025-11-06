package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// Severity represents error severity levels aligned with assessment schema
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// SeverityLevel maps severity names to numeric levels
var SeverityLevel = map[Severity]int{
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

// ErrorEnvelope extends the pathfinder error response with telemetry fields
type ErrorEnvelope struct {
	// Base pathfinder fields
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Timestamp string                 `json:"timestamp"`

	// Extended telemetry fields
	Severity      Severity               `json:"severity,omitempty"`
	SeverityLevel int                    `json:"severity_level,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	ExitCode      *int                   `json:"exit_code,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Original      interface{}            `json:"original,omitempty"`
}

// NewErrorEnvelope creates a new error envelope with required fields
func NewErrorEnvelope(code, message string) *ErrorEnvelope {
	start := time.Now()
	defer func() {
		telemetry.EmitCounter(metrics.ErrorHandlingWrapsTotal, 1, map[string]string{metrics.TagOperation: "new_envelope"})
		telemetry.EmitHistogram(metrics.ErrorHandlingWrapMs, time.Since(start), map[string]string{metrics.TagOperation: "new_envelope"})
	}()

	return &ErrorEnvelope{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// WithSeverity adds severity classification, validating against the schema enum.
// If an invalid severity is provided, it defaults to "info" and returns an error.
func (e *ErrorEnvelope) WithSeverity(severity Severity) (*ErrorEnvelope, error) {
	// Validate severity against the allowed enum values
	switch severity {
	case SeverityInfo, SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		e.Severity = severity
		e.SeverityLevel = SeverityLevel[severity]
		return e, nil
	default:
		// Default to info severity for invalid values
		e.Severity = SeverityInfo
		e.SeverityLevel = SeverityLevel[SeverityInfo]
		return e, fmt.Errorf("invalid severity %q, must be one of: info, low, medium, high, critical", severity)
	}
}

// WithCorrelationID adds correlation identifier
func (e *ErrorEnvelope) WithCorrelationID(id string) *ErrorEnvelope {
	e.CorrelationID = id
	return e
}

// WithTraceID adds tracing identifier
func (e *ErrorEnvelope) WithTraceID(id string) *ErrorEnvelope {
	e.TraceID = id
	return e
}

// WithExitCode adds process exit code
func (e *ErrorEnvelope) WithExitCode(code int) *ErrorEnvelope {
	e.ExitCode = &code
	return e
}

// WithContext adds structured context, validating entries against schema constraints.
// Only allows: string, number, boolean, or array of strings.
// Invalid entries are filtered out and an error is returned.
func (e *ErrorEnvelope) WithContext(context map[string]interface{}) (*ErrorEnvelope, error) {
	if context == nil {
		e.Context = nil
		return e, nil
	}

	validatedContext := make(map[string]interface{})
	var validationErrors []string

	for key, value := range context {
		if err := validateContextValue(value); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("key %q: %s", key, err))
			continue // Skip invalid entries
		}
		validatedContext[key] = value
	}

	e.Context = validatedContext

	if len(validationErrors) > 0 {
		return e, fmt.Errorf("context validation failed: %s", strings.Join(validationErrors, "; "))
	}
	return e, nil
}

// validateContextValue validates a single context value against schema constraints
func validateContextValue(value interface{}) error {
	switch v := value.(type) {
	case string, float64, int, bool:
		return nil
	case []interface{}:
		// Check that all array elements are strings
		for i, elem := range v {
			if _, ok := elem.(string); !ok {
				return fmt.Errorf("array element at index %d is not a string (got %T)", i, elem)
			}
		}
		return nil
	case []string:
		// Already validated as string array
		return nil
	default:
		return fmt.Errorf("invalid type %T, must be string, number, boolean, or string array", value)
	}
}

// WithOriginal adds the original error
func (e *ErrorEnvelope) WithOriginal(original error) *ErrorEnvelope {
	if original != nil {
		e.Original = original.Error()
	}
	return e
}

// WithDetails adds error details
func (e *ErrorEnvelope) WithDetails(details map[string]interface{}) *ErrorEnvelope {
	e.Details = details
	return e
}

// WithPath adds path information
func (e *ErrorEnvelope) WithPath(path string) *ErrorEnvelope {
	e.Path = path
	return e
}

// Error implements the error interface
func (e *ErrorEnvelope) Error() string {
	severity := e.Severity
	if severity == "" {
		severity = SeverityInfo // Default to info if not set
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, severity, e.Message)
}

// MarshalJSON ensures proper JSON serialization
func (e *ErrorEnvelope) MarshalJSON() ([]byte, error) {
	type Alias ErrorEnvelope
	return json.Marshal((*Alias)(e))
}

// GenerateCorrelationID creates a new UUID for correlation
func GenerateCorrelationID() string {
	return uuid.New().String()
}
