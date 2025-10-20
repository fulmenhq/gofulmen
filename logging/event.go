package logging

import (
	"encoding/json"
	"time"

	"github.com/fulmenhq/gofulmen/foundry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogEvent represents a structured log event matching Crucible schema.
//
// This envelope format provides full 20+ field Crucible observability standard
// for cross-language compatibility with pyfulmen and tsfulmen.
type LogEvent struct {
	Timestamp     time.Time      `json:"timestamp"`
	Severity      Severity       `json:"severity"`
	SeverityLevel int            `json:"severityLevel"`
	Message       string         `json:"message"`
	Logger        string         `json:"logger,omitempty"`
	Service       string         `json:"service"`
	Component     string         `json:"component,omitempty"`
	Environment   string         `json:"environment,omitempty"`
	Context       map[string]any `json:"context,omitempty"`
	Error         *LogError      `json:"error,omitempty"`
	TraceID       string         `json:"traceId,omitempty"`
	SpanID        string         `json:"spanId,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	EventID       string         `json:"eventId,omitempty"`

	// Extended Crucible envelope fields for middleware processing
	ContextID     string   `json:"contextId,omitempty"`
	RequestID     string   `json:"requestId,omitempty"`
	CorrelationID string   `json:"correlationId,omitempty"`
	ParentSpanID  string   `json:"parentSpanId,omitempty"`
	Operation     string   `json:"operation,omitempty"`
	DurationMs    *float64 `json:"durationMs,omitempty"`
	UserID        string   `json:"userId,omitempty"`

	// Middleware processing metadata (not serialized)
	ThrottleBucket  string   `json:"-"`
	RedactionFlags  []string `json:"-"`
	DroppedByPolicy bool     `json:"-"`
}

// LogError represents error information in log events
type LogError struct {
	Message string         `json:"message"`
	Type    string         `json:"type"`
	Stack   string         `json:"stack,omitempty"`
	Code    string         `json:"code,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// NewLogEvent creates a LogEvent from zapcore.Entry and fields.
//
// Bridges zap logging infrastructure with Crucible event envelope, extracting
// structured fields from zap.Field array and populating the full 20+ field schema.
//
// Auto-generates correlation ID using foundry.GenerateCorrelationID() if not present
// in fields.
func NewLogEvent(entry zapcore.Entry, fields []zapcore.Field, config *LoggerConfig) *LogEvent {
	severity := FromZapLevel(entry.Level)

	event := &LogEvent{
		Timestamp:     entry.Time,
		Severity:      severity,
		SeverityLevel: severity.Level(),
		Message:       entry.Message,
		Logger:        entry.LoggerName,
		Service:       config.Service,
		Component:     config.Component,
		Environment:   config.Environment,
		Context:       make(map[string]any),
	}

	extractFieldsToEvent(event, fields)

	if event.CorrelationID == "" {
		event.CorrelationID = foundry.GenerateCorrelationID()
	}

	return event
}

// ToJSON serializes the LogEvent to JSON bytes.
//
// Produces Crucible-compliant JSON output for cross-language validation
// and log aggregation systems.
func (e *LogEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// extractFieldsToEvent extracts zap.Field array into LogEvent fields.
//
// Handles all standard zap field types and maps them to appropriate
// LogEvent envelope fields (traceId, spanId, correlationId, etc.).
func extractFieldsToEvent(event *LogEvent, fields []zapcore.Field) {
	encoder := zapcore.NewMapObjectEncoder()

	for i := range fields {
		fields[i].AddTo(encoder)
	}

	for key, value := range encoder.Fields {
		switch key {
		case "traceId", "trace_id":
			if s, ok := value.(string); ok {
				event.TraceID = s
			}
		case "spanId", "span_id":
			if s, ok := value.(string); ok {
				event.SpanID = s
			}
		case "parentSpanId", "parent_span_id":
			if s, ok := value.(string); ok {
				event.ParentSpanID = s
			}
		case "correlationId", "correlation_id":
			if s, ok := value.(string); ok {
				event.CorrelationID = s
			}
		case "requestId", "request_id":
			if s, ok := value.(string); ok {
				event.RequestID = s
			}
		case "contextId", "context_id":
			if s, ok := value.(string); ok {
				event.ContextID = s
			}
		case "userId", "user_id":
			if s, ok := value.(string); ok {
				event.UserID = s
			}
		case "operation":
			if s, ok := value.(string); ok {
				event.Operation = s
			}
		case "durationMs", "duration_ms":
			if f, ok := value.(float64); ok {
				event.DurationMs = &f
			}
		case "eventId", "event_id":
			if s, ok := value.(string); ok {
				event.EventID = s
			}
		case "tags":
			switch v := value.(type) {
			case []interface{}:
				tags := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						tags = append(tags, s)
					}
				}
				event.Tags = tags
			case []string:
				event.Tags = v
			}
		case "error":
			if m, ok := value.(map[string]interface{}); ok {
				event.Error = extractLogError(m)
			}
		default:
			event.Context[key] = value
		}
	}
}

// extractLogError converts map[string]interface{} to LogError struct.
func extractLogError(m map[string]interface{}) *LogError {
	logErr := &LogError{}

	if msg, ok := m["message"].(string); ok {
		logErr.Message = msg
	}
	if typ, ok := m["type"].(string); ok {
		logErr.Type = typ
	}
	if stack, ok := m["stack"].(string); ok {
		logErr.Stack = stack
	}
	if code, ok := m["code"].(string); ok {
		logErr.Code = code
	}
	if details, ok := m["details"].(map[string]interface{}); ok {
		logErr.Details = details
	}

	if logErr.Message != "" {
		return logErr
	}
	return nil
}

// FromZapLevel converts zapcore.Level to Severity enum.
func FromZapLevel(level zapcore.Level) Severity {
	switch level {
	case zapcore.DebugLevel:
		return DEBUG
	case zapcore.InfoLevel:
		return INFO
	case zapcore.WarnLevel:
		return WARN
	case zapcore.ErrorLevel:
		return ERROR
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return FATAL
	default:
		return INFO
	}
}

// With returns a copy of the LogEvent with additional context fields.
//
// Useful for middleware enrichment without mutating original events.
// Deep copies pointer fields to ensure true immutability.
func (e *LogEvent) With(key string, value any) *LogEvent {
	copied := *e

	if e.DurationMs != nil {
		durationCopy := *e.DurationMs
		copied.DurationMs = &durationCopy
	}

	if e.Error != nil {
		errorCopy := *e.Error
		if e.Error.Details != nil {
			errorCopy.Details = make(map[string]any, len(e.Error.Details))
			for k, v := range e.Error.Details {
				errorCopy.Details[k] = v
			}
		}
		copied.Error = &errorCopy
	}

	if e.Tags != nil {
		copied.Tags = make([]string, len(e.Tags))
		copy(copied.Tags, e.Tags)
	}

	if e.RedactionFlags != nil {
		copied.RedactionFlags = make([]string, len(e.RedactionFlags))
		copy(copied.RedactionFlags, e.RedactionFlags)
	}

	if copied.Context == nil {
		copied.Context = make(map[string]any)
	}
	copiedContext := make(map[string]any, len(e.Context)+1)
	for k, v := range e.Context {
		copiedContext[k] = v
	}
	copiedContext[key] = value
	copied.Context = copiedContext

	return &copied
}

// ToZapFields converts LogEvent back to zap.Field array.
//
// Used when middleware pipeline modifies event and needs to write
// through underlying zap core.
func (e *LogEvent) ToZapFields() []zap.Field {
	fields := make([]zap.Field, 0, 20)

	if e.Service != "" {
		fields = append(fields, zap.String("service", e.Service))
	}
	if e.Component != "" {
		fields = append(fields, zap.String("component", e.Component))
	}
	if e.Environment != "" {
		fields = append(fields, zap.String("environment", e.Environment))
	}
	if e.TraceID != "" {
		fields = append(fields, zap.String("traceId", e.TraceID))
	}
	if e.SpanID != "" {
		fields = append(fields, zap.String("spanId", e.SpanID))
	}
	if e.ParentSpanID != "" {
		fields = append(fields, zap.String("parentSpanId", e.ParentSpanID))
	}
	if e.CorrelationID != "" {
		fields = append(fields, zap.String("correlationId", e.CorrelationID))
	}
	if e.RequestID != "" {
		fields = append(fields, zap.String("requestId", e.RequestID))
	}
	if e.ContextID != "" {
		fields = append(fields, zap.String("contextId", e.ContextID))
	}
	if e.UserID != "" {
		fields = append(fields, zap.String("userId", e.UserID))
	}
	if e.Operation != "" {
		fields = append(fields, zap.String("operation", e.Operation))
	}
	if e.DurationMs != nil {
		fields = append(fields, zap.Float64("durationMs", *e.DurationMs))
	}
	if e.EventID != "" {
		fields = append(fields, zap.String("eventId", e.EventID))
	}
	if len(e.Tags) > 0 {
		fields = append(fields, zap.Strings("tags", e.Tags))
	}
	if e.Error != nil {
		fields = append(fields, zap.Any("error", e.Error))
	}

	for k, v := range e.Context {
		fields = append(fields, zap.Any(k, v))
	}

	return fields
}
