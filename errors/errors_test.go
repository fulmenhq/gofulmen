package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewErrorEnvelope(t *testing.T) {
	envelope := NewErrorEnvelope("TEST_ERROR", "This is a test error")

	assert.Equal(t, "TEST_ERROR", envelope.Code)
	assert.Equal(t, "This is a test error", envelope.Message)
	assert.NotEmpty(t, envelope.Timestamp)

	// Parse timestamp to ensure it's valid RFC3339
	_, err := time.Parse(time.RFC3339, envelope.Timestamp)
	assert.NoError(t, err)
}

func TestErrorEnvelopeWithSeverity(t *testing.T) {
	envelope := NewErrorEnvelope("TEST", "test")
	envelope, err := envelope.WithSeverity(SeverityHigh)
	require.NoError(t, err)

	assert.Equal(t, SeverityHigh, envelope.Severity)
	assert.Equal(t, 3, envelope.SeverityLevel)
}

func TestErrorEnvelopeWithCorrelationID(t *testing.T) {
	id := "test-correlation-id"
	envelope := NewErrorEnvelope("TEST", "test").
		WithCorrelationID(id)

	assert.Equal(t, id, envelope.CorrelationID)
}

func TestErrorEnvelopeWithTraceID(t *testing.T) {
	id := "test-trace-id"
	envelope := NewErrorEnvelope("TEST", "test").
		WithTraceID(id)

	assert.Equal(t, id, envelope.TraceID)
}

func TestErrorEnvelopeWithExitCode(t *testing.T) {
	code := 42
	envelope := NewErrorEnvelope("TEST", "test").
		WithExitCode(code)

	assert.NotNil(t, envelope.ExitCode)
	assert.Equal(t, code, *envelope.ExitCode)
}

func TestErrorEnvelopeWithContext(t *testing.T) {
	context := map[string]interface{}{
		"component": "test-component",
		"user_id":   123,
	}
	envelope := NewErrorEnvelope("TEST", "test")
	envelope, err := envelope.WithContext(context)
	require.NoError(t, err)

	assert.Equal(t, context, envelope.Context)
}

func TestErrorEnvelopeWithOriginal(t *testing.T) {
	original := assert.AnError
	envelope := NewErrorEnvelope("TEST", "test").
		WithOriginal(original)

	assert.Equal(t, original.Error(), envelope.Original)
}

func TestErrorEnvelopeWithDetails(t *testing.T) {
	details := map[string]interface{}{
		"field":      "username",
		"constraint": "required",
	}
	envelope := NewErrorEnvelope("TEST", "test").
		WithDetails(details)

	assert.Equal(t, details, envelope.Details)
}

func TestErrorEnvelopeWithPath(t *testing.T) {
	path := "/test/path"
	envelope := NewErrorEnvelope("TEST", "test").
		WithPath(path)

	assert.Equal(t, path, envelope.Path)
}

func TestErrorEnvelopeError(t *testing.T) {
	envelope := NewErrorEnvelope("TEST_ERROR", "test message")
	envelope, err := envelope.WithSeverity(SeverityCritical)
	require.NoError(t, err)

	expected := "[TEST_ERROR] critical: test message"
	assert.Equal(t, expected, envelope.Error())
}

func TestErrorEnvelopeJSONSerialization(t *testing.T) {
	envelope := NewErrorEnvelope("TEST_ERROR", "test message")
	envelope, err := envelope.WithSeverity(SeverityHigh)
	require.NoError(t, err)
	envelope = envelope.WithCorrelationID("test-id")
	envelope, err = envelope.WithContext(map[string]interface{}{"key": "value"})
	require.NoError(t, err)

	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	var unmarshaled ErrorEnvelope
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, envelope.Code, unmarshaled.Code)
	assert.Equal(t, envelope.Message, unmarshaled.Message)
	assert.Equal(t, envelope.Severity, unmarshaled.Severity)
	assert.Equal(t, envelope.SeverityLevel, unmarshaled.SeverityLevel)
	assert.Equal(t, envelope.CorrelationID, unmarshaled.CorrelationID)
	assert.Equal(t, envelope.Context, unmarshaled.Context)
}

func TestGenerateCorrelationID(t *testing.T) {
	id1 := GenerateCorrelationID()
	id2 := GenerateCorrelationID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)

	// Should be valid UUID format
	assert.Len(t, id1, 36) // UUID v4 string length
}

func TestSeverityLevelMapping(t *testing.T) {
	tests := []struct {
		severity Severity
		level    int
	}{
		{SeverityInfo, 0},
		{SeverityLow, 1},
		{SeverityMedium, 2},
		{SeverityHigh, 3},
		{SeverityCritical, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.severity), func(t *testing.T) {
			assert.Equal(t, tt.level, SeverityLevel[tt.severity])
		})
	}
}

func TestWithSeverityValidation(t *testing.T) {
	tests := []struct {
		name             string
		inputSeverity    Severity
		expectError      bool
		expectedSeverity Severity
		expectedLevel    int
	}{
		{
			name:             "valid severity - info",
			inputSeverity:    SeverityInfo,
			expectError:      false,
			expectedSeverity: SeverityInfo,
			expectedLevel:    0,
		},
		{
			name:             "valid severity - critical",
			inputSeverity:    SeverityCritical,
			expectError:      false,
			expectedSeverity: SeverityCritical,
			expectedLevel:    4,
		},
		{
			name:             "invalid severity - defaults to info",
			inputSeverity:    Severity("invalid"),
			expectError:      true,
			expectedSeverity: SeverityInfo,
			expectedLevel:    0,
		},
		{
			name:             "empty severity - defaults to info",
			inputSeverity:    Severity(""),
			expectError:      true,
			expectedSeverity: SeverityInfo,
			expectedLevel:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := NewErrorEnvelope("TEST", "test")
			result, err := envelope.WithSeverity(tt.inputSeverity)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSeverity, result.Severity)
			assert.Equal(t, tt.expectedLevel, result.SeverityLevel)
		})
	}
}

func TestWithContextValidation(t *testing.T) {
	tests := []struct {
		name         string
		inputContext map[string]interface{}
		expectError  bool
		expectedKeys []string
	}{
		{
			name: "valid context - strings and numbers",
			inputContext: map[string]interface{}{
				"component": "test",
				"user_id":   123,
				"active":    true,
			},
			expectError:  false,
			expectedKeys: []string{"component", "user_id", "active"},
		},
		{
			name: "valid context - string array",
			inputContext: map[string]interface{}{
				"tags": []string{"error", "validation"},
			},
			expectError:  false,
			expectedKeys: []string{"tags"},
		},
		{
			name: "valid context - mixed interface array of strings",
			inputContext: map[string]interface{}{
				"tags": []interface{}{"error", "validation"},
			},
			expectError:  false,
			expectedKeys: []string{"tags"},
		},
		{
			name: "invalid context - nested object",
			inputContext: map[string]interface{}{
				"component": "test",
				"nested":    map[string]string{"key": "value"},
			},
			expectError:  true,
			expectedKeys: []string{"component"}, // nested should be filtered out
		},
		{
			name: "invalid context - non-string array",
			inputContext: map[string]interface{}{
				"component": "test",
				"numbers":   []int{1, 2, 3},
			},
			expectError:  true,
			expectedKeys: []string{"component"}, // numbers should be filtered out
		},
		{
			name:         "nil context",
			inputContext: nil,
			expectError:  false,
			expectedKeys: nil,
		},
		{
			name:         "empty context",
			inputContext: map[string]interface{}{},
			expectError:  false,
			expectedKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envelope := NewErrorEnvelope("TEST", "test")
			result, err := envelope.WithContext(tt.inputContext)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedKeys == nil {
				assert.Nil(t, result.Context)
			} else {
				assert.NotNil(t, result.Context)
				assert.Equal(t, len(tt.expectedKeys), len(result.Context))
				for _, key := range tt.expectedKeys {
					assert.Contains(t, result.Context, key)
				}
			}
		})
	}
}

func TestValidateContextValue(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectError bool
	}{
		{"string", "test", false},
		{"int", 123, false},
		{"float64", 123.45, false},
		{"bool", true, false},
		{"string array", []string{"a", "b"}, false},
		{"interface string array", []interface{}{"a", "b"}, false},
		{"empty string array", []string{}, false},
		{"nested object", map[string]string{"key": "value"}, true},
		{"int array", []int{1, 2, 3}, true},
		{"mixed array", []interface{}{"a", 123}, true},
		{"nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContextValue(tt.value)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrorEnvelopeErrorWithNoSeverity(t *testing.T) {
	// Test that Error() works when severity is not set (defaults to info)
	envelope := NewErrorEnvelope("TEST_ERROR", "test message")
	expected := "[TEST_ERROR] info: test message"
	assert.Equal(t, expected, envelope.Error())
}

// TestBackwardCompatibility ensures existing error patterns remain unchanged
func TestBackwardCompatibility(t *testing.T) {
	// Test that standard library errors.New still works
	stdErr := errors.New("standard error")
	assert.Equal(t, "standard error", stdErr.Error())

	// Test that fmt.Errorf still works
	fmtErr := fmt.Errorf("formatted error: %s", "test")
	assert.Equal(t, "formatted error: test", fmtErr.Error())

	// Test that wrapped errors work
	wrappedErr := fmt.Errorf("wrapped: %w", stdErr)
	assert.ErrorIs(t, wrappedErr, stdErr)
}
