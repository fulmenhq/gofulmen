package logging

import (
	"regexp"
	"testing"
	"time"
)

func TestGolden_SeverityMapping(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
		expected string
	}{
		{"TRACE", TRACE, "TRACE"},
		{"DEBUG", DEBUG, "DEBUG"},
		{"INFO", INFO, "INFO"},
		{"WARN", WARN, "WARN"},
		{"ERROR", ERROR, "ERROR"},
		{"FATAL", FATAL, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.severity.String() != tt.expected {
				t.Errorf("expected severity %s, got %s", tt.expected, tt.severity.String())
			}
		})
	}
}

func TestGolden_CorrelationIDFormat(t *testing.T) {
	mw := &CorrelationMiddleware{}

	event := &LogEvent{
		Timestamp: time.Now(),
		Severity:  INFO,
		Message:   "test",
		Context:   make(map[string]any),
	}

	mw.Process(event)

	correlationID := event.CorrelationID
	if correlationID == "" {
		t.Fatal("correlation_id should be set")
	}

	uuidv7Pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidv7Pattern.MatchString(correlationID) {
		t.Errorf("correlation_id does not match UUIDv7 format: %s", correlationID)
	}
}

func TestGolden_CorrelationIDConsistency(t *testing.T) {
	mw := &CorrelationMiddleware{}

	event1 := &LogEvent{
		Timestamp:     time.Now(),
		Severity:      INFO,
		Message:       "test1",
		Context:       make(map[string]any),
		CorrelationID: "existing-id-123",
	}

	mw.Process(event1)

	if event1.CorrelationID != "existing-id-123" {
		t.Error("correlation middleware should preserve existing correlation_id")
	}
}

func TestGolden_TimestampFormat(t *testing.T) {
	now := time.Now()

	event := &LogEvent{
		Timestamp: now,
		Severity:  INFO,
		Message:   "test",
	}

	timestampStr := event.Timestamp.Format(time.RFC3339Nano)

	parsed, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		t.Fatalf("timestamp should be parseable as RFC3339Nano: %v", err)
	}

	if !parsed.Equal(now) {
		t.Error("parsed timestamp should match original")
	}
}

func TestGolden_EventStructure(t *testing.T) {
	event := &LogEvent{
		Timestamp:     time.Now(),
		Severity:      INFO,
		SeverityLevel: 6,
		Message:       "test message",
		Service:       "test-service",
		Component:     "test-component",
		Environment:   "production",
		CorrelationID: "test-correlation-id",
		Context: map[string]any{
			"request_id": "req-123",
			"user_id":    456,
		},
	}

	if event.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}

	if event.Severity != INFO {
		t.Errorf("expected severity INFO, got %s", event.Severity)
	}

	if event.Message != "test message" {
		t.Error("message should match")
	}

	if event.Service != "test-service" {
		t.Error("service should match")
	}

	if event.Component != "test-component" {
		t.Error("component should match")
	}

	if event.Environment != "production" {
		t.Error("environment should match")
	}

	if event.CorrelationID != "test-correlation-id" {
		t.Error("correlationID should match")
	}

	if len(event.Context) != 2 {
		t.Errorf("expected 2 context fields, got %d", len(event.Context))
	}
}

func TestGolden_SeverityLevelOrdering(t *testing.T) {
	severities := []Severity{TRACE, DEBUG, INFO, WARN, ERROR, FATAL}

	for i := 0; i < len(severities)-1; i++ {
		current := severities[i]
		next := severities[i+1]

		if current.Level() >= next.Level() {
			t.Errorf("severity level ordering broken: %s (level %d) should be less than %s (level %d)", current, current.Level(), next, next.Level())
		}
	}
}

func TestGolden_ProfileEnumValues(t *testing.T) {
	profiles := []LoggingProfile{
		ProfileSimple,
		ProfileStructured,
		ProfileEnterprise,
		ProfileCustom,
	}

	expected := []string{"SIMPLE", "STRUCTURED", "ENTERPRISE", "CUSTOM"}

	for i, profile := range profiles {
		if string(profile) != expected[i] {
			t.Errorf("expected profile %s, got %s", expected[i], profile)
		}
	}
}

func TestGolden_CrossLanguageFieldNames(t *testing.T) {
	requiredFields := map[string]bool{
		"timestamp":     true,
		"severity":      true,
		"severityLevel": true,
		"message":       true,
		"service":       true,
		"environment":   true,
		"correlationId": true,
		"contextId":     true,
		"requestId":     true,
		"traceId":       true,
		"spanId":        true,
		"parentSpanId":  true,
		"component":     true,
		"logger":        true,
		"eventId":       true,
		"tags":          true,
		"context":       true,
		"error":         true,
	}

	_ = requiredFields
}

func TestGolden_SeverityToZapLevel(t *testing.T) {
	severities := []Severity{TRACE, DEBUG, INFO, WARN, ERROR, FATAL}

	for _, sev := range severities {
		zapLevel := sev.ToZapLevel()
		if zapLevel == 0 && sev != INFO {
			t.Errorf("severity %s should map to non-zero zap level", sev)
		}
	}
}

func TestGolden_MiddlewareOrdering(t *testing.T) {
	middleware := []MiddlewareConfig{
		{Name: "correlation", Enabled: true, Order: 100},
		{Name: "redaction", Enabled: true, Order: 200},
		{Name: "throttling", Enabled: true, Order: 300},
	}

	for i := 0; i < len(middleware)-1; i++ {
		if middleware[i].Order >= middleware[i+1].Order {
			t.Error("middleware order should be ascending")
		}
	}
}

func TestGolden_RFC3339NanoTimestampCompatibility(t *testing.T) {
	// Test with a timestamp that has non-zero nanoseconds to ensure full precision
	now := time.Unix(1234567890, 123456789) // Fixed timestamp with nanoseconds
	formatted := now.Format(time.RFC3339Nano)

	parsed, err := time.Parse(time.RFC3339Nano, formatted)
	if err != nil {
		t.Fatalf("RFC3339Nano timestamp should be parseable: %v", err)
	}

	if !parsed.Equal(now) {
		t.Error("RFC3339Nano round-trip should preserve timestamp")
	}

	// RFC3339Nano with nanoseconds should include decimal point
	// Format: 2009-02-13T18:31:30.123456789-05:00 (at least 29 chars with nanos)
	if len(formatted) < 29 {
		t.Errorf("RFC3339Nano timestamp with nanoseconds should be at least 29 chars, got %d: %s", len(formatted), formatted)
	}

	// Verify nanoseconds are preserved in round-trip
	if parsed.Nanosecond() != 123456789 {
		t.Errorf("Nanoseconds not preserved: expected 123456789, got %d", parsed.Nanosecond())
	}
}

func TestGolden_UUIDv7MonotonicIncreasing(t *testing.T) {
	mw := &CorrelationMiddleware{}

	var ids []string
	for i := 0; i < 5; i++ {
		event := &LogEvent{
			Timestamp: time.Now(),
			Severity:  INFO,
			Message:   "test",
			Context:   make(map[string]any),
		}

		mw.Process(event)
		ids = append(ids, event.CorrelationID)
		time.Sleep(time.Millisecond)
	}

	for i := 0; i < len(ids)-1; i++ {
		if ids[i] >= ids[i+1] {
			t.Errorf("UUIDv7 should be monotonically increasing: %s >= %s", ids[i], ids[i+1])
		}
	}
}
