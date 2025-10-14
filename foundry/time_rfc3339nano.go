package foundry

import (
	"fmt"
	"time"
)

// TimestampRFC3339Nano is a time.Time that always marshals with nanosecond precision.
//
// Standard time.Time marshals to RFC3339 (second precision), losing nanoseconds.
// This type ensures RFC3339Nano format for cross-language compatibility with
// pyfulmen and tsfulmen.
//
// Use this type when:
//   - Nanosecond precision is required in logs, traces, or events
//   - Cross-language timestamp consistency is needed
//   - Avoiding silent precision loss in JSON/YAML serialization
//
// Example:
//
//	type Event struct {
//	    ID        string                `json:"id"`
//	    Timestamp TimestampRFC3339Nano `json:"timestamp"`
//	    Data      map[string]interface{} `json:"data"`
//	}
//
//	event := Event{
//	    ID:        "evt-123",
//	    Timestamp: foundry.Now(),
//	    Data:      map[string]interface{}{"key": "value"},
//	}
//
//	// Always marshals with nanosecond precision:
//	// {"id":"evt-123","timestamp":"2025-10-14T14:32:15.123456789Z","data":{"key":"value"}}
type TimestampRFC3339Nano time.Time

// Now returns the current time as a TimestampRFC3339Nano.
//
// This is equivalent to NewTimestamp(time.Now()).
//
// Example:
//
//	ts := foundry.Now()
//	fmt.Println(ts) // 2025-10-14T14:32:15.123456789Z
func Now() TimestampRFC3339Nano {
	return TimestampRFC3339Nano(time.Now().UTC())
}

// NewTimestamp creates a TimestampRFC3339Nano from a time.Time.
//
// The time is converted to UTC for consistency with RFC3339Nano format.
//
// Example:
//
//	t := time.Date(2025, 10, 14, 14, 32, 15, 123456789, time.UTC)
//	ts := foundry.NewTimestamp(t)
func NewTimestamp(t time.Time) TimestampRFC3339Nano {
	return TimestampRFC3339Nano(t.UTC())
}

// Time returns the underlying time.Time value.
//
// This allows using TimestampRFC3339Nano with standard time package functions.
//
// Example:
//
//	ts := foundry.Now()
//	t := ts.Time()
//	fmt.Println(t.Year()) // 2025
func (t TimestampRFC3339Nano) Time() time.Time {
	return time.Time(t)
}

// String returns the timestamp formatted as RFC3339Nano.
//
// This implements the fmt.Stringer interface.
//
// Example:
//
//	ts := foundry.Now()
//	fmt.Println(ts.String()) // 2025-10-14T14:32:15.123456789Z
func (t TimestampRFC3339Nano) String() string {
	return time.Time(t).UTC().Format(time.RFC3339Nano)
}

// MarshalText implements encoding.TextMarshaler for JSON, YAML, TOML support.
//
// The timestamp is always marshaled in RFC3339Nano format with nanosecond precision.
//
// Example:
//
//	ts := foundry.Now()
//	data, _ := json.Marshal(ts)
//	// "2025-10-14T14:32:15.123456789Z"
func (t TimestampRFC3339Nano) MarshalText() ([]byte, error) {
	formatted := time.Time(t).UTC().Format(time.RFC3339Nano)
	return []byte(formatted), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON, YAML, TOML support.
//
// Accepts RFC3339Nano format and is lenient with fractional seconds precision.
// Supports both microsecond and nanosecond precision.
//
// Example:
//
//	var ts foundry.TimestampRFC3339Nano
//	json.Unmarshal([]byte(`"2025-10-14T14:32:15.123456789Z"`), &ts)
func (t *TimestampRFC3339Nano) UnmarshalText(text []byte) error {
	parsed, err := time.Parse(time.RFC3339Nano, string(text))
	if err != nil {
		return fmt.Errorf("invalid RFC3339Nano timestamp: %w", err)
	}

	*t = TimestampRFC3339Nano(parsed.UTC())
	return nil
}

// Before reports whether the timestamp t is before u.
//
// Example:
//
//	t1 := foundry.Now()
//	time.Sleep(10 * time.Millisecond)
//	t2 := foundry.Now()
//	fmt.Println(t1.Before(t2)) // true
func (t TimestampRFC3339Nano) Before(u TimestampRFC3339Nano) bool {
	return time.Time(t).Before(time.Time(u))
}

// After reports whether the timestamp t is after u.
//
// Example:
//
//	t1 := foundry.Now()
//	time.Sleep(10 * time.Millisecond)
//	t2 := foundry.Now()
//	fmt.Println(t2.After(t1)) // true
func (t TimestampRFC3339Nano) After(u TimestampRFC3339Nano) bool {
	return time.Time(t).After(time.Time(u))
}

// Equal reports whether t and u represent the same time instant.
//
// Example:
//
//	t1 := foundry.NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
//	t2 := foundry.NewTimestamp(time.Date(2025, 10, 14, 14, 32, 15, 0, time.UTC))
//	fmt.Println(t1.Equal(t2)) // true
func (t TimestampRFC3339Nano) Equal(u TimestampRFC3339Nano) bool {
	return time.Time(t).Equal(time.Time(u))
}

// IsZero reports whether t represents the zero time instant.
//
// Example:
//
//	var t foundry.TimestampRFC3339Nano
//	fmt.Println(t.IsZero()) // true
func (t TimestampRFC3339Nano) IsZero() bool {
	return time.Time(t).IsZero()
}

// Unix returns the number of seconds since January 1, 1970 UTC.
//
// Example:
//
//	ts := foundry.Now()
//	fmt.Println(ts.Unix()) // 1697293935
func (t TimestampRFC3339Nano) Unix() int64 {
	return time.Time(t).Unix()
}

// UnixNano returns the number of nanoseconds since January 1, 1970 UTC.
//
// Example:
//
//	ts := foundry.Now()
//	fmt.Println(ts.UnixNano()) // 1697293935123456789
func (t TimestampRFC3339Nano) UnixNano() int64 {
	return time.Time(t).UnixNano()
}

// Add returns the timestamp t+d.
//
// Example:
//
//	ts := foundry.Now()
//	future := ts.Add(1 * time.Hour)
func (t TimestampRFC3339Nano) Add(d time.Duration) TimestampRFC3339Nano {
	return TimestampRFC3339Nano(time.Time(t).Add(d))
}

// Sub returns the duration t-u.
//
// Example:
//
//	t1 := foundry.Now()
//	time.Sleep(100 * time.Millisecond)
//	t2 := foundry.Now()
//	duration := t2.Sub(t1)
//	fmt.Println(duration) // ~100ms
func (t TimestampRFC3339Nano) Sub(u TimestampRFC3339Nano) time.Duration {
	return time.Time(t).Sub(time.Time(u))
}
