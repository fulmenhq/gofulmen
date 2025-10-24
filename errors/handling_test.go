package errors

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorHandlingStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy ErrorHandlingStrategy
		validate func(*testing.T, *ErrorEnvelope, *bytes.Buffer)
	}{
		{
			name:     "StrategyLogWarning",
			strategy: StrategyLogWarning,
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should log warning but return original envelope
				assert.Contains(t, buf.String(), "Warning: failed to set severity")
				assert.Equal(t, "test", envelope.Message) // Original message preserved
			},
		},
		{
			name:     "StrategyAppendToMessage",
			strategy: StrategyAppendToMessage,
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should append error to message
				assert.Contains(t, envelope.Message, "(severity error:")
				assert.Contains(t, envelope.Message, "invalid severity")
			},
		},
		{
			name:     "StrategyFailFast",
			strategy: StrategyFailFast,
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should log error and return original envelope
				assert.Contains(t, buf.String(), "Error: severity validation failed")
				assert.Equal(t, "test", envelope.Message) // Original message preserved
			},
		},
		{
			name:     "StrategySilent",
			strategy: StrategySilent,
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should not log anything and return original envelope
				assert.Empty(t, buf.String())
				assert.Equal(t, "test", envelope.Message) // Original message preserved
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &ErrorHandlingConfig{
				SeverityStrategy: tt.strategy,
				ContextStrategy:  StrategySilent, // Don't interfere with severity test
				Logger:           log.New(&buf, "", 0),
			}

			// Create a fresh envelope and try to apply an invalid severity
			envelope := NewErrorEnvelope("TEST", "test")

			// This will trigger the error handling because "invalid" is not a valid severity
			result := ApplySeverityWithHandling(envelope, Severity("invalid"), config)

			tt.validate(t, result, &buf)
		})
	}
}

func TestApplyContextWithHandling(t *testing.T) {
	tests := []struct {
		name     string
		strategy ErrorHandlingStrategy
		context  map[string]interface{}
		validate func(*testing.T, *ErrorEnvelope, *bytes.Buffer)
	}{
		{
			name:     "StrategyLogWarning with valid context",
			strategy: StrategyLogWarning,
			context:  map[string]interface{}{"component": "test"},
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should succeed without error
				assert.Equal(t, "test", envelope.Message)
				assert.Empty(t, buf.String())
			},
		},
		{
			name:     "StrategyLogWarning with invalid context",
			strategy: StrategyLogWarning,
			context:  map[string]interface{}{"nested": map[string]string{"key": "value"}},
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should log warning but return original envelope
				assert.Contains(t, buf.String(), "Warning: failed to set context")
				assert.Equal(t, "test", envelope.Message)
			},
		},
		{
			name:     "StrategyAppendToMessage with invalid context",
			strategy: StrategyAppendToMessage,
			context:  map[string]interface{}{"nested": map[string]string{"key": "value"}},
			validate: func(t *testing.T, envelope *ErrorEnvelope, buf *bytes.Buffer) {
				// Should append error to message
				assert.Contains(t, envelope.Message, "(context error:")
				assert.Contains(t, envelope.Message, "invalid type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := &ErrorHandlingConfig{
				SeverityStrategy: StrategySilent,
				ContextStrategy:  tt.strategy,
				Logger:           log.New(&buf, "", 0),
			}

			envelope := NewErrorEnvelope("TEST", "test")
			result := ApplyContextWithHandling(envelope, tt.context, config)

			tt.validate(t, result, &buf)
		})
	}
}

func TestSafeWithSeverity(t *testing.T) {
	envelope := NewErrorEnvelope("TEST", "test")
	result := SafeWithSeverity(envelope, SeverityHigh)

	assert.Equal(t, SeverityHigh, result.Severity)
	assert.Equal(t, 3, result.SeverityLevel)
}

func TestSafeWithContext(t *testing.T) {
	envelope := NewErrorEnvelope("TEST", "test")
	context := map[string]interface{}{"component": "test"}

	result := SafeWithContext(envelope, context)

	assert.Equal(t, context, result.Context)
}

func TestDefaultErrorHandlingConfig(t *testing.T) {
	config := DefaultErrorHandlingConfig()

	assert.Equal(t, StrategyLogWarning, config.SeverityStrategy)
	assert.Equal(t, StrategyAppendToMessage, config.ContextStrategy)
	assert.NotNil(t, config.Logger)
}

func TestErrorHandlingWithNilConfig(t *testing.T) {
	envelope := NewErrorEnvelope("TEST", "test")

	// Test with nil config (should use defaults)
	result := ApplySeverityWithHandling(envelope, SeverityHigh, nil)

	assert.Equal(t, SeverityHigh, result.Severity)
	assert.Equal(t, 3, result.SeverityLevel)
}

func TestErrorHandlingWithCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	customLogger := log.New(&buf, "[CUSTOM] ", 0)

	config := &ErrorHandlingConfig{
		SeverityStrategy: StrategyLogWarning,
		ContextStrategy:  StrategySilent,
		Logger:           customLogger,
	}

	envelope := NewErrorEnvelope("TEST", "test")

	// This will trigger the error handling because "invalid" is not a valid severity
	_ = ApplySeverityWithHandling(envelope, Severity("invalid"), config)

	output := buf.String()
	assert.Contains(t, output, "[CUSTOM]")
	assert.Contains(t, output, "Warning: failed to set severity")
}

// Test that error handling preserves envelope state
func TestErrorHandlingPreservesEnvelopeState(t *testing.T) {
	envelope := NewErrorEnvelope("TEST", "test message")
	envelope = envelope.WithCorrelationID("test-123")
	envelope = envelope.WithTraceID("trace-456")

	config := &ErrorHandlingConfig{
		SeverityStrategy: StrategyAppendToMessage,
		ContextStrategy:  StrategySilent,
		Logger:           log.New(&bytes.Buffer{}, "", 0),
	}

	// This will trigger the error handling because "invalid" is not a valid severity
	result := ApplySeverityWithHandling(envelope, Severity("invalid"), config)

	// Verify all other fields are preserved
	assert.Equal(t, "test message (severity error: invalid severity \"invalid\", must be one of: info, low, medium, high, critical)", result.Message)
	assert.Equal(t, "test-123", result.CorrelationID)
	assert.Equal(t, "trace-456", result.TraceID)
}
