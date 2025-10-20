package logging

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fulmenhq/gofulmen/foundry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogEvent(t *testing.T) {
	config := &LoggerConfig{
		Service:     "test-service",
		Environment: "test",
	}

	entry := zapcore.Entry{
		Level:      zapcore.InfoLevel,
		Time:       time.Now(),
		LoggerName: "test-logger",
		Message:    "test message",
	}

	fields := []zapcore.Field{
		zap.String("traceId", "trace-123"),
		zap.String("spanId", "span-456"),
		zap.String("requestId", "req-789"),
		zap.String("contextId", "ctx-abc"),
		zap.String("userId", "user-def"),
		zap.String("operation", "test.operation"),
		zap.Float64("durationMs", 123.45),
		zap.Strings("tags", []string{"tag1", "tag2"}),
		zap.String("custom", "value"),
	}

	event := NewLogEvent(entry, fields, config)

	if event.Severity != INFO {
		t.Errorf("expected severity INFO, got %v", event.Severity)
	}

	if event.SeverityLevel != 20 {
		t.Errorf("expected severityLevel 20, got %d", event.SeverityLevel)
	}

	if event.Message != "test message" {
		t.Errorf("expected message 'test message', got %q", event.Message)
	}

	if event.Service != "test-service" {
		t.Errorf("expected service 'test-service', got %q", event.Service)
	}

	if event.TraceID != "trace-123" {
		t.Errorf("expected traceId 'trace-123', got %q", event.TraceID)
	}

	if event.SpanID != "span-456" {
		t.Errorf("expected spanId 'span-456', got %q", event.SpanID)
	}

	if event.RequestID != "req-789" {
		t.Errorf("expected requestId 'req-789', got %q", event.RequestID)
	}

	if event.ContextID != "ctx-abc" {
		t.Errorf("expected contextId 'ctx-abc', got %q", event.ContextID)
	}

	if event.UserID != "user-def" {
		t.Errorf("expected userId 'user-def', got %q", event.UserID)
	}

	if event.Operation != "test.operation" {
		t.Errorf("expected operation 'test.operation', got %q", event.Operation)
	}

	if event.DurationMs == nil || *event.DurationMs != 123.45 {
		t.Errorf("expected durationMs 123.45, got %v", event.DurationMs)
	}

	if len(event.Tags) != 2 || event.Tags[0] != "tag1" || event.Tags[1] != "tag2" {
		t.Errorf("expected tags [tag1, tag2], got %v", event.Tags)
	}

	if event.Context["custom"] != "value" {
		t.Errorf("expected context.custom 'value', got %v", event.Context["custom"])
	}

	if event.CorrelationID == "" {
		t.Error("expected auto-generated correlationId")
	}

	if !foundry.IsValidCorrelationID(event.CorrelationID) {
		t.Errorf("expected valid UUIDv7 correlationId, got %q", event.CorrelationID)
	}
}

func TestNewLogEvent_PreservesCorrelationID(t *testing.T) {
	config := &LoggerConfig{
		Service: "test-service",
	}

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	existingCorrID := foundry.GenerateCorrelationID()
	fields := []zapcore.Field{
		zap.String("correlationId", existingCorrID),
	}

	event := NewLogEvent(entry, fields, config)

	if event.CorrelationID != existingCorrID {
		t.Errorf("expected correlation ID preserved as %q, got %q", existingCorrID, event.CorrelationID)
	}
}

func TestLogEventToJSON(t *testing.T) {
	now := time.Now()
	durationMs := 123.45

	event := &LogEvent{
		Timestamp:     now,
		Severity:      INFO,
		SeverityLevel: 20,
		Message:       "test message",
		Service:       "test-service",
		Environment:   "production",
		TraceID:       "trace-123",
		SpanID:        "span-456",
		CorrelationID: foundry.GenerateCorrelationID(),
		RequestID:     "req-789",
		Operation:     "test.operation",
		DurationMs:    &durationMs,
		Context: map[string]any{
			"key": "value",
		},
		Tags: []string{"tag1", "tag2"},
	}

	jsonBytes, err := event.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if decoded["severity"] != "INFO" {
		t.Errorf("expected severity INFO, got %v", decoded["severity"])
	}

	if decoded["severityLevel"].(float64) != 20 {
		t.Errorf("expected severityLevel 20, got %v", decoded["severityLevel"])
	}

	if decoded["message"] != "test message" {
		t.Errorf("expected message 'test message', got %v", decoded["message"])
	}

	if decoded["service"] != "test-service" {
		t.Errorf("expected service 'test-service', got %v", decoded["service"])
	}

	if decoded["traceId"] != "trace-123" {
		t.Errorf("expected traceId 'trace-123', got %v", decoded["traceId"])
	}
}

func TestLogEventWith(t *testing.T) {
	durationMs := 123.45
	original := &LogEvent{
		Service:    "test-service",
		Message:    "original",
		DurationMs: &durationMs,
		Tags:       []string{"tag1", "tag2"},
		Error: &LogError{
			Message: "error",
			Details: map[string]any{"key": "value"},
		},
		RedactionFlags: []string{"secret"},
		Context: map[string]any{
			"existing": "value",
		},
	}

	enriched := original.With("newKey", "newValue")

	if enriched == original {
		t.Error("With() should return a new instance, not mutate original")
	}

	if original.Context["newKey"] != nil {
		t.Error("With() mutated original event context")
	}

	if enriched.Context["newKey"] != "newValue" {
		t.Errorf("expected newKey in enriched context, got %v", enriched.Context)
	}

	if enriched.Context["existing"] != "value" {
		t.Error("With() did not preserve existing context")
	}

	if enriched.Service != original.Service {
		t.Error("With() did not copy other fields")
	}

	if enriched.DurationMs == original.DurationMs {
		t.Error("DurationMs pointer should be deep copied, not shared")
	}

	if enriched.Error == original.Error {
		t.Error("Error pointer should be deep copied, not shared")
	}

	if len(enriched.Tags) != len(original.Tags) {
		t.Error("Tags should be copied")
	}

	enriched.Tags[0] = "modified"
	if original.Tags[0] == "modified" {
		t.Error("Tags mutation leaked to original")
	}

	*enriched.DurationMs = 999.99
	if *original.DurationMs == 999.99 {
		t.Error("DurationMs mutation leaked to original")
	}

	enriched.Error.Details["new"] = "value"
	if original.Error.Details["new"] != nil {
		t.Error("Error.Details mutation leaked to original")
	}
}

func TestLogEventToZapFields(t *testing.T) {
	durationMs := 123.45

	event := &LogEvent{
		Service:       "test-service",
		Environment:   "production",
		TraceID:       "trace-123",
		SpanID:        "span-456",
		CorrelationID: foundry.GenerateCorrelationID(),
		RequestID:     "req-789",
		ContextID:     "ctx-abc",
		UserID:        "user-def",
		Operation:     "test.operation",
		DurationMs:    &durationMs,
		EventID:       "evt-123",
		Tags:          []string{"tag1", "tag2"},
		Context: map[string]any{
			"custom": "value",
		},
		Error: &LogError{
			Message: "error message",
			Type:    "TestError",
		},
	}

	fields := event.ToZapFields()

	if len(fields) == 0 {
		t.Fatal("ToZapFields returned empty array")
	}

	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(encoder)
	}
	fieldMap := encoder.Fields

	if fieldMap["service"] != "test-service" {
		t.Errorf("expected service 'test-service', got %v", fieldMap["service"])
	}

	if fieldMap["traceId"] != "trace-123" {
		t.Errorf("expected traceId 'trace-123', got %v", fieldMap["traceId"])
	}

	if fieldMap["correlationId"] != event.CorrelationID {
		t.Errorf("expected correlationId %q, got %v", event.CorrelationID, fieldMap["correlationId"])
	}

	if fieldMap["durationMs"] != 123.45 {
		t.Errorf("expected durationMs 123.45, got %v", fieldMap["durationMs"])
	}

	if fieldMap["custom"] != "value" {
		t.Errorf("expected custom context field, got %v", fieldMap["custom"])
	}
}

func TestExtractFieldsToEvent_ErrorHandling(t *testing.T) {
	config := &LoggerConfig{
		Service: "test-service",
	}

	entry := zapcore.Entry{
		Level:   zapcore.ErrorLevel,
		Time:    time.Now(),
		Message: "error occurred",
	}

	fields := []zapcore.Field{
		zap.Any("error", map[string]interface{}{
			"message": "test error",
			"type":    "TestError",
			"stack":   "stack trace",
			"code":    "ERR001",
			"details": map[string]interface{}{
				"key": "value",
			},
		}),
	}

	event := NewLogEvent(entry, fields, config)

	if event.Error == nil {
		t.Fatal("expected error field to be populated")
	}

	if event.Error.Message != "test error" {
		t.Errorf("expected error message 'test error', got %q", event.Error.Message)
	}

	if event.Error.Type != "TestError" {
		t.Errorf("expected error type 'TestError', got %q", event.Error.Type)
	}

	if event.Error.Stack != "stack trace" {
		t.Errorf("expected error stack 'stack trace', got %q", event.Error.Stack)
	}

	if event.Error.Code != "ERR001" {
		t.Errorf("expected error code 'ERR001', got %q", event.Error.Code)
	}

	if event.Error.Details["key"] != "value" {
		t.Errorf("expected error details key 'value', got %v", event.Error.Details)
	}
}

func TestFromZapLevel(t *testing.T) {
	tests := []struct {
		zapLevel zapcore.Level
		expected Severity
	}{
		{zapcore.DebugLevel, DEBUG},
		{zapcore.InfoLevel, INFO},
		{zapcore.WarnLevel, WARN},
		{zapcore.ErrorLevel, ERROR},
		{zapcore.DPanicLevel, FATAL},
		{zapcore.PanicLevel, FATAL},
		{zapcore.FatalLevel, FATAL},
		{zapcore.InvalidLevel, INFO},
	}

	for _, tt := range tests {
		t.Run(tt.zapLevel.String(), func(t *testing.T) {
			result := FromZapLevel(tt.zapLevel)
			if result != tt.expected {
				t.Errorf("FromZapLevel(%v) = %v, want %v", tt.zapLevel, result, tt.expected)
			}
		})
	}
}

func TestExtractFieldsToEvent_SnakeCaseFields(t *testing.T) {
	config := &LoggerConfig{
		Service: "test-service",
	}

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	fields := []zapcore.Field{
		zap.String("trace_id", "trace-snake"),
		zap.String("span_id", "span-snake"),
		zap.String("correlation_id", "corr-snake"),
		zap.String("request_id", "req-snake"),
		zap.String("context_id", "ctx-snake"),
		zap.String("user_id", "user-snake"),
		zap.String("event_id", "evt-snake"),
		zap.Float64("duration_ms", 99.99),
	}

	event := NewLogEvent(entry, fields, config)

	if event.TraceID != "trace-snake" {
		t.Errorf("expected trace_id mapped to TraceID, got %q", event.TraceID)
	}

	if event.SpanID != "span-snake" {
		t.Errorf("expected span_id mapped to SpanID, got %q", event.SpanID)
	}

	if event.CorrelationID != "corr-snake" {
		t.Errorf("expected correlation_id mapped to CorrelationID, got %q", event.CorrelationID)
	}

	if event.RequestID != "req-snake" {
		t.Errorf("expected request_id mapped to RequestID, got %q", event.RequestID)
	}

	if event.ContextID != "ctx-snake" {
		t.Errorf("expected context_id mapped to ContextID, got %q", event.ContextID)
	}

	if event.UserID != "user-snake" {
		t.Errorf("expected user_id mapped to UserID, got %q", event.UserID)
	}

	if event.EventID != "evt-snake" {
		t.Errorf("expected event_id mapped to EventID, got %q", event.EventID)
	}

	if event.DurationMs == nil || *event.DurationMs != 99.99 {
		t.Errorf("expected duration_ms mapped to DurationMs, got %v", event.DurationMs)
	}
}

func TestExtractFieldsToEvent_ParentSpanID(t *testing.T) {
	config := &LoggerConfig{
		Service: "test-service",
	}

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	fields := []zapcore.Field{
		zap.String("parentSpanId", "parent-camel"),
	}

	event := NewLogEvent(entry, fields, config)

	if event.ParentSpanID != "parent-camel" {
		t.Errorf("expected parentSpanId mapped to ParentSpanID, got %q", event.ParentSpanID)
	}

	fields2 := []zapcore.Field{
		zap.String("parent_span_id", "parent-snake"),
	}

	event2 := NewLogEvent(entry, fields2, config)

	if event2.ParentSpanID != "parent-snake" {
		t.Errorf("expected parent_span_id mapped to ParentSpanID, got %q", event2.ParentSpanID)
	}
}

func TestNewLogEvent_ComponentField(t *testing.T) {
	config := &LoggerConfig{
		Service:   "test-service",
		Component: "test-component",
	}

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	event := NewLogEvent(entry, []zapcore.Field{}, config)

	if event.Component != "test-component" {
		t.Errorf("expected component 'test-component', got %q", event.Component)
	}
}

func TestExtractFieldsToEvent_TagsStringArray(t *testing.T) {
	config := &LoggerConfig{
		Service: "test-service",
	}

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Now(),
		Message: "test",
	}

	fields := []zapcore.Field{
		zap.Strings("tags", []string{"tag1", "tag2", "tag3"}),
	}

	event := NewLogEvent(entry, fields, config)

	if len(event.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(event.Tags))
	}

	if event.Tags[0] != "tag1" || event.Tags[1] != "tag2" || event.Tags[2] != "tag3" {
		t.Errorf("expected tags [tag1, tag2, tag3], got %v", event.Tags)
	}
}
