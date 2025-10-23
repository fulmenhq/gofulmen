package foundry

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestNow tests the Now() constructor
func TestNow(t *testing.T) {
	before := time.Now()
	ts := Now()
	after := time.Now()

	// Verify timestamp is between before and after
	if ts.Time().Before(before) || ts.Time().After(after) {
		t.Error("Now() timestamp not within expected range")
	}

	// Verify timestamp is in UTC
	if ts.Time().Location() != time.UTC {
		t.Error("Now() should return UTC time")
	}

	// Verify not zero
	if ts.IsZero() {
		t.Error("Now() should not return zero time")
	}
}

// TestNewTimestamp tests the NewTimestamp() constructor
func TestNewTimestamp(t *testing.T) {
	// Create specific time with nanosecond precision
	original := time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC)
	ts := NewTimestamp(original)

	// Verify Time() returns the same value
	if !ts.Time().Equal(original) {
		t.Errorf("NewTimestamp() time mismatch: got %v, want %v", ts.Time(), original)
	}

	// Verify nanosecond precision preserved
	if ts.Time().Nanosecond() != 123456789 {
		t.Errorf("Nanosecond precision lost: got %d, want 123456789", ts.Time().Nanosecond())
	}

	// Verify UTC conversion
	if ts.Time().Location() != time.UTC {
		t.Error("NewTimestamp() should convert to UTC")
	}
}

// TestNewTimestamp_NonUTC tests conversion of non-UTC times
func TestNewTimestamp_NonUTC(t *testing.T) {
	// Create time in EST (UTC-5)
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("EST timezone not available")
	}

	local := time.Date(2025, 10, 14, 10, 0, 0, 0, est)
	ts := NewTimestamp(local)

	// Should be converted to UTC
	if ts.Time().Location() != time.UTC {
		t.Error("NewTimestamp() should convert to UTC")
	}

	// Verify time is equivalent (same instant, different zone)
	if !ts.Time().Equal(local) {
		t.Error("NewTimestamp() should preserve time instant")
	}
}

// TestTimestampRFC3339Nano_Time tests the Time() accessor
func TestTimestampRFC3339Nano_Time(t *testing.T) {
	original := time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC)
	ts := NewTimestamp(original)
	extracted := ts.Time()

	if !extracted.Equal(original) {
		t.Errorf("Time() returned different value: got %v, want %v", extracted, original)
	}

	if extracted.Nanosecond() != 123456789 {
		t.Errorf("Nanosecond precision lost in Time(): got %d", extracted.Nanosecond())
	}
}

// TestTimestampRFC3339Nano_String tests the String() method
func TestTimestampRFC3339Nano_String(t *testing.T) {
	ts := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC))
	str := ts.String()

	expected := "2025-10-14T14:32:15.123456789Z"
	if str != expected {
		t.Errorf("String() = %q, want %q", str, expected)
	}

	// Verify it contains nanoseconds
	if len(str) < 30 { // RFC3339Nano is at least 30 chars with full nanoseconds
		t.Errorf("String() appears to have lost nanosecond precision: %q", str)
	}
}

// TestTimestampRFC3339Nano_JSONMarshal tests JSON marshaling
func TestTimestampRFC3339Nano_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		contains string // expected substring in JSON output
	}{
		{
			name:     "WithNanoseconds",
			time:     time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC),
			contains: "2025-10-14T14:32:15.123456789Z",
		},
		{
			name:     "WithMicroseconds",
			time:     time.Date(2025, 10, 14, 14, 32, 15, 123456000, time.UTC),
			contains: "2025-10-14T14:32:15.123456",
		},
		{
			name:     "WithMilliseconds",
			time:     time.Date(2025, 10, 14, 14, 32, 15, 123000000, time.UTC),
			contains: "2025-10-14T14:32:15.123",
		},
		{
			name:     "NoFractionalSeconds",
			time:     time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC),
			contains: "2025-10-14T14:32:15Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTimestamp(tt.time)
			data, err := json.Marshal(ts)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}

			output := string(data)
			if !containsString(output, tt.contains) {
				t.Errorf("JSON output %q does not contain expected %q", output, tt.contains)
			}
		})
	}
}

// TestTimestampRFC3339Nano_JSONUnmarshal tests JSON unmarshaling
func TestTimestampRFC3339Nano_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNanos int
		wantErr   bool
	}{
		{
			name:      "FullNanoseconds",
			input:     `"2025-10-14T14:32:15.123456789Z"`,
			wantNanos: 123456789,
			wantErr:   false,
		},
		{
			name:      "Microseconds",
			input:     `"2025-10-14T14:32:15.123456Z"`,
			wantNanos: 123456000,
			wantErr:   false,
		},
		{
			name:      "Milliseconds",
			input:     `"2025-10-14T14:32:15.123Z"`,
			wantNanos: 123000000,
			wantErr:   false,
		},
		{
			name:      "NoFractional",
			input:     `"2025-10-14T14:32:15Z"`,
			wantNanos: 0,
			wantErr:   false,
		},
		{
			name:    "Invalid",
			input:   `"not-a-timestamp"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TimestampRFC3339Nano
			err := json.Unmarshal([]byte(tt.input), &ts)

			if (err != nil) != tt.wantErr {
				t.Fatalf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if ts.Time().Nanosecond() != tt.wantNanos {
					t.Errorf("Nanosecond precision: got %d, want %d", ts.Time().Nanosecond(), tt.wantNanos)
				}
			}
		})
	}
}

// TestTimestampRFC3339Nano_JSONRoundTrip tests JSON marshal/unmarshal round-trip
func TestTimestampRFC3339Nano_JSONRoundTrip(t *testing.T) {
	type Event struct {
		Timestamp TimestampRFC3339Nano `json:"timestamp"`
		Message   string               `json:"message"`
	}

	tests := []struct {
		name string
		time time.Time
	}{
		{
			name: "FullNanoseconds",
			time: time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC),
		},
		{
			name: "Microseconds",
			time: time.Date(2025, 10, 14, 14, 32, 15, 123456000, time.UTC),
		},
		{
			name: "Milliseconds",
			time: time.Date(2025, 10, 14, 14, 32, 15, 123000000, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Event{
				Timestamp: NewTimestamp(tt.time),
				Message:   "test event",
			}

			// Marshal
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("json.Marshal() error: %v", err)
			}

			// Unmarshal
			var decoded Event
			err = json.Unmarshal(data, &decoded)
			if err != nil {
				t.Fatalf("json.Unmarshal() error: %v", err)
			}

			// Compare (use Equal for time comparison due to precision)
			if !decoded.Timestamp.Equal(original.Timestamp) {
				t.Errorf("Round-trip failed: got %v, want %v", decoded.Timestamp, original.Timestamp)
			}

			// Verify nanosecond precision preserved
			origNanos := original.Timestamp.Time().Nanosecond()
			decodedNanos := decoded.Timestamp.Time().Nanosecond()
			if decodedNanos != origNanos {
				t.Errorf("Nanosecond precision lost: got %d, want %d", decodedNanos, origNanos)
			}
		})
	}
}

// TestTimestampRFC3339Nano_YAMLMarshal tests YAML marshaling
func TestTimestampRFC3339Nano_YAMLMarshal(t *testing.T) {
	type Config struct {
		CreatedAt TimestampRFC3339Nano `yaml:"created_at"`
	}

	ts := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC))
	config := Config{CreatedAt: ts}

	data, err := yaml.Marshal(config)
	if err != nil {
		t.Fatalf("yaml.Marshal() error: %v", err)
	}

	output := string(data)
	if !containsString(output, "2025-10-14T14:32:15") {
		t.Errorf("YAML output missing timestamp: %s", output)
	}
}

// TestTimestampRFC3339Nano_YAMLUnmarshal tests YAML unmarshaling
func TestTimestampRFC3339Nano_YAMLUnmarshal(t *testing.T) {
	type Config struct {
		CreatedAt TimestampRFC3339Nano `yaml:"created_at"`
	}

	input := `created_at: 2025-10-14T14:32:15.123456789Z`

	var config Config
	err := yaml.Unmarshal([]byte(input), &config)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() error: %v", err)
	}

	if config.CreatedAt.IsZero() {
		t.Error("Unmarshaled timestamp is zero")
	}

	if config.CreatedAt.Time().Nanosecond() != 123456789 {
		t.Errorf("Nanosecond precision lost: got %d", config.CreatedAt.Time().Nanosecond())
	}
}

// TestTimestampRFC3339Nano_Before tests the Before() method
func TestTimestampRFC3339Nano_Before(t *testing.T) {
	t1 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
	t2 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 16, 0, time.UTC))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}

	if t2.Before(t1) {
		t.Error("t2 should not be before t1")
	}

	if t1.Before(t1) {
		t.Error("t1 should not be before itself")
	}
}

// TestTimestampRFC3339Nano_After tests the After() method
func TestTimestampRFC3339Nano_After(t *testing.T) {
	t1 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
	t2 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 16, 0, time.UTC))

	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}

	if t1.After(t2) {
		t.Error("t1 should not be after t2")
	}

	if t1.After(t1) {
		t.Error("t1 should not be after itself")
	}
}

// TestTimestampRFC3339Nano_Equal tests the Equal() method
func TestTimestampRFC3339Nano_Equal(t *testing.T) {
	t1 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC))
	t2 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC))
	t3 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456790, time.UTC))

	if !t1.Equal(t2) {
		t.Error("t1 should equal t2")
	}

	if t1.Equal(t3) {
		t.Error("t1 should not equal t3 (different nanoseconds)")
	}
}

// TestTimestampRFC3339Nano_IsZero tests the IsZero() method
func TestTimestampRFC3339Nano_IsZero(t *testing.T) {
	var zero TimestampRFC3339Nano
	if !zero.IsZero() {
		t.Error("Default value should be zero")
	}

	now := Now()
	if now.IsZero() {
		t.Error("Now() should not be zero")
	}

	explicit := NewTimestamp(time.Time{})
	if !explicit.IsZero() {
		t.Error("Zero time should be zero")
	}
}

// TestTimestampRFC3339Nano_Unix tests the Unix() method
func TestTimestampRFC3339Nano_Unix(t *testing.T) {
	// January 1, 1970 00:00:00 UTC
	epoch := NewTimestamp(time.Unix(0, 0))
	if epoch.Unix() != 0 {
		t.Errorf("Epoch Unix() = %d, want 0", epoch.Unix())
	}

	// Known timestamp
	ts := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
	unix := ts.Unix()
	if unix <= 0 {
		t.Error("Unix() should return positive value for future date")
	}
}

// TestTimestampRFC3339Nano_UnixNano tests the UnixNano() method
func TestTimestampRFC3339Nano_UnixNano(t *testing.T) {
	ts := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC))
	nanos := ts.UnixNano()

	// Verify nanoseconds are included
	lastDigits := nanos % 1000000000
	if lastDigits != 123456789 {
		t.Errorf("UnixNano() lost nanosecond precision: got %d, want 123456789", lastDigits)
	}
}

// TestTimestampRFC3339Nano_Add tests the Add() method
func TestTimestampRFC3339Nano_Add(t *testing.T) {
	ts := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))

	// Add 1 hour
	future := ts.Add(1 * time.Hour)
	expected := NewTimestamp(time.Date(2025, 10, 14, 15, 32, 15, 0, time.UTC))

	if !future.Equal(expected) {
		t.Errorf("Add(1h) failed: got %v, want %v", future, expected)
	}

	// Add negative duration (subtract)
	past := ts.Add(-1 * time.Hour)
	expectedPast := NewTimestamp(time.Date(2025, 10, 14, 13, 32, 15, 0, time.UTC))

	if !past.Equal(expectedPast) {
		t.Errorf("Add(-1h) failed: got %v, want %v", past, expectedPast)
	}
}

// TestTimestampRFC3339Nano_Sub tests the Sub() method
func TestTimestampRFC3339Nano_Sub(t *testing.T) {
	t1 := NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
	t2 := NewTimestamp(time.Date(2025, 10, 14, 15, 32, 15, 0, time.UTC))

	duration := t2.Sub(t1)
	expected := 1 * time.Hour

	if duration != expected {
		t.Errorf("Sub() = %v, want %v", duration, expected)
	}

	// Reverse
	reverseDuration := t1.Sub(t2)
	expectedReverse := -1 * time.Hour

	if reverseDuration != expectedReverse {
		t.Errorf("Sub() reverse = %v, want %v", reverseDuration, expectedReverse)
	}
}

// TestTimestampRFC3339Nano_NanosecondPrecision verifies nanosecond precision is maintained
func TestTimestampRFC3339Nano_NanosecondPrecision(t *testing.T) {
	// Create time with specific nanoseconds
	original := time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC)
	ts := NewTimestamp(original)

	// Marshal to JSON
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal back
	var decoded TimestampRFC3339Nano
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify exact nanosecond match
	if decoded.Time().Nanosecond() != 123456789 {
		t.Errorf("Nanosecond precision lost: got %d, want 123456789", decoded.Time().Nanosecond())
	}

	// Verify full time equality
	if !decoded.Equal(ts) {
		t.Error("Round-trip produced different timestamp")
	}
}

// TestTimestampRFC3339Nano_StandardTimeComparison compares with standard time.Time marshaling
func TestTimestampRFC3339Nano_StandardTimeComparison(t *testing.T) {
	original := time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC)

	// Standard time.Time JSON marshaling
	standardData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Standard time.Time marshal error: %v", err)
	}

	// TimestampRFC3339Nano marshaling
	ts := NewTimestamp(original)
	nanoData, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("TimestampRFC3339Nano marshal error: %v", err)
	}

	// Compare lengths - nano should have more digits
	if len(nanoData) <= len(standardData) {
		t.Logf("Standard: %s", string(standardData))
		t.Logf("Nano:     %s", string(nanoData))
		// Note: This may not always fail if standard time happens to include nanos,
		// but it demonstrates the difference in formats
	}

	// Verify nano format contains nanoseconds
	nanoStr := string(nanoData)
	if !containsString(nanoStr, "123456789") {
		t.Errorf("TimestampRFC3339Nano should preserve nanoseconds: %s", nanoStr)
	}
}

// Benchmarks

func BenchmarkNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Now()
	}
}

func BenchmarkNewTimestamp(b *testing.B) {
	t := time.Now()
	for i := 0; i < b.N; i++ {
		NewTimestamp(t)
	}
}

func BenchmarkTimestampRFC3339Nano_MarshalJSON(b *testing.B) {
	ts := Now()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(ts)
	}
}

func BenchmarkTimestampRFC3339Nano_UnmarshalJSON(b *testing.B) {
	data := []byte(`"2025-10-14T14:32:15.123456789Z"`)
	for i := 0; i < b.N; i++ {
		var ts TimestampRFC3339Nano
		_ = json.Unmarshal(data, &ts)
	}
}

func BenchmarkTimestampRFC3339Nano_String(b *testing.B) {
	ts := Now()
	for i := 0; i < b.N; i++ {
		_ = ts.String()
	}
}
