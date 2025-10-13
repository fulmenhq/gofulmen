package foundry

import (
	"strings"
	"testing"
	"time"
)

func TestUTCNowRFC3339Nano(t *testing.T) {
	timestamp := UTCNowRFC3339Nano()

	// Check format matches RFC3339Nano with Z suffix
	if !strings.HasSuffix(timestamp, "Z") {
		t.Errorf("Expected timestamp to end with 'Z', got: %s", timestamp)
	}

	// Verify it's parseable
	parsed, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		t.Errorf("Failed to parse timestamp: %v", err)
	}

	// Verify it's approximately current time (within 1 second)
	now := time.Now().UTC()
	diff := now.Sub(parsed)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("Timestamp differs from current time by %v (expected < 1s)", diff)
	}
}

func TestFormatRFC3339Nano(t *testing.T) {
	testTime := time.Date(2025, 10, 13, 14, 32, 15, 123456789, time.UTC)
	formatted := FormatRFC3339Nano(testTime)

	// Verify format
	if !strings.HasPrefix(formatted, "2025-10-13T14:32:15.") {
		t.Errorf("Unexpected formatted time: %s", formatted)
	}

	if !strings.HasSuffix(formatted, "Z") {
		t.Errorf("Expected formatted time to end with 'Z', got: %s", formatted)
	}

	// Verify it's parseable
	parsed, err := time.Parse(time.RFC3339Nano, formatted)
	if err != nil {
		t.Errorf("Failed to parse formatted timestamp: %v", err)
	}

	// Verify parsed time matches original (within nanosecond precision)
	if !parsed.Equal(testTime) {
		t.Errorf("Parsed time %v doesn't match original %v", parsed, testTime)
	}
}

func TestFormatRFC3339Nano_NonUTC(t *testing.T) {
	// Test with non-UTC timezone - should convert to UTC
	loc, _ := time.LoadLocation("America/New_York")
	testTime := time.Date(2025, 10, 13, 10, 32, 15, 123456789, loc)
	formatted := FormatRFC3339Nano(testTime)

	// Should be converted to UTC (14:32:15 UTC = 10:32:15 EDT)
	if !strings.Contains(formatted, "14:32:15") {
		t.Errorf("Expected time to be converted to UTC, got: %s", formatted)
	}

	if !strings.HasSuffix(formatted, "Z") {
		t.Errorf("Expected formatted time to end with 'Z', got: %s", formatted)
	}
}

func TestParseRFC3339Nano(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "Valid nanosecond precision",
			input:     "2025-10-13T14:32:15.123456789Z",
			expectErr: false,
		},
		{
			name:      "Valid microsecond precision",
			input:     "2025-10-13T14:32:15.123456Z",
			expectErr: false,
		},
		{
			name:      "Valid millisecond precision",
			input:     "2025-10-13T14:32:15.123Z",
			expectErr: false,
		},
		{
			name:      "Valid no fractional seconds",
			input:     "2025-10-13T14:32:15Z",
			expectErr: false,
		},
		{
			name:      "Invalid format",
			input:     "2025-10-13 14:32:15",
			expectErr: true,
		},
		{
			name:      "Invalid format missing T",
			input:     "2025-10-13 14:32:15Z",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseRFC3339Nano(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error parsing %q, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error parsing %q: %v", tt.input, err)
				}

				// Verify parsed time is in UTC
				if parsed.Location() != time.UTC {
					t.Errorf("Expected UTC location, got: %v", parsed.Location())
				}
			}
		})
	}
}

func TestRFC3339Nano_RoundTrip(t *testing.T) {
	// Test that we can format and parse back to the same value
	original := time.Date(2025, 10, 13, 14, 32, 15, 123456789, time.UTC)
	formatted := FormatRFC3339Nano(original)
	parsed, err := ParseRFC3339Nano(formatted)

	if err != nil {
		t.Fatalf("Failed to parse formatted timestamp: %v", err)
	}

	if !parsed.Equal(original) {
		t.Errorf("Round-trip failed: original=%v, parsed=%v", original, parsed)
	}
}
