package foundry

import (
	"fmt"

	"github.com/google/uuid"
)

// GenerateCorrelationID generates a time-sortable UUIDv7 for correlation tracking.
//
// UUIDv7 embeds a timestamp in the UUID (first 48 bits), making it naturally time-sortable.
// This is beneficial for log aggregation systems (Splunk, Datadog, etc.) and ensures
// consistent cross-service correlation tracking.
//
// Benefits of UUIDv7:
//   - Time-sortable: UUIDs generated at different times sort chronologically
//   - Monotonic: IDs generally increase over time (with some entropy for uniqueness)
//   - Globally unique: Random bits ensure uniqueness across distributed systems
//   - Database-friendly: Better index performance than random UUIDv4
//
// Returns a UUIDv7 string in standard format (8-4-4-4-12 hyphenated).
//
// Example output: "018b2c5e-8f4a-7890-b123-456789abcdef"
//
// Note: All *fulmen libraries (gofulmen, tsfulmen, pyfulmen) use UUIDv7
// for consistency across languages and services.
func GenerateCorrelationID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// MustGenerateCorrelationID is an alias for GenerateCorrelationID.
//
// This function exists for consistency with pyfulmen's API where generate_correlation_id
// is the primary function name. In Go, we provide both names for flexibility.
//
// Deprecated: Use GenerateCorrelationID instead for clearer semantics.
func MustGenerateCorrelationID() string {
	return GenerateCorrelationID()
}

// ParseCorrelationID parses a UUIDv7 string and returns the uuid.UUID object.
//
// This can be useful for extracting timestamp information or validating
// correlation IDs.
//
// Example:
//
//	id := GenerateCorrelationID()
//	parsed, err := ParseCorrelationID(id)
//	if err != nil {
//	    // Handle invalid UUID
//	}
//	// Use parsed UUID
func ParseCorrelationID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// IsValidCorrelationID checks if a string is a valid UUIDv7.
//
// This validates both UUID format and ensures the version is specifically v7
// for time-sortable guarantees required by correlation tracking.
//
// Example:
//
//	if IsValidCorrelationID("018b2c5e-8f4a-7890-b123-456789abcdef") {
//	    // Valid UUIDv7 format
//	}
func IsValidCorrelationID(s string) bool {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return false
	}
	return parsed.Version() == 7
}

// CorrelationID is a validated UUIDv7 correlation ID newtype for distributed tracing.
//
// This type enforces UUIDv7 format and provides type safety for correlation IDs
// across service boundaries. UUIDv7 embeds a timestamp (first 48 bits), making
// IDs naturally time-sortable and beneficial for log aggregation.
//
// Use this type when:
//   - Tracking requests across distributed services
//   - Correlating logs, traces, and metrics
//   - Ensuring type safety for correlation IDs in APIs
//   - Validating correlation ID format at service boundaries
//
// Example:
//
//	type Request struct {
//	    CorrelationID CorrelationID     `json:"correlation_id" header:"X-Correlation-ID"`
//	    Data          map[string]string `json:"data"`
//	}
//
//	req := Request{
//	    CorrelationID: foundry.NewCorrelationIDValue(),
//	    Data:          map[string]string{"key": "value"},
//	}
type CorrelationID string

// NewCorrelationIDValue generates a new time-sortable UUIDv7 correlation ID.
//
// Returns a correlation ID suitable for distributed tracing. The ID is globally
// unique and time-ordered for efficient log aggregation and indexing.
//
// Example:
//
//	id := foundry.NewCorrelationIDValue()
//	fmt.Println(id) // 018b2c5e-8f4a-7890-b123-456789abcdef
func NewCorrelationIDValue() CorrelationID {
	return CorrelationID(GenerateCorrelationID())
}

// ParseCorrelationIDValue parses and validates a correlation ID string.
//
// Returns an error if the string is not a valid UUIDv7 format.
// Enforces UUIDv7 to ensure time-sortable guarantees.
//
// Example:
//
//	id, err := foundry.ParseCorrelationIDValue("018b2c5e-8f4a-7890-b123-456789abcdef")
//	if err != nil {
//	    log.Fatal("Invalid correlation ID:", err)
//	}
func ParseCorrelationIDValue(s string) (CorrelationID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return "", fmt.Errorf("invalid correlation ID format: %w", err)
	}

	// Enforce UUIDv7 for time-sortable guarantees
	if parsed.Version() != 7 {
		return "", fmt.Errorf("correlation ID must be UUIDv7, got UUID version %d", parsed.Version())
	}

	return CorrelationID(s), nil
}

// String returns the correlation ID as a string.
func (c CorrelationID) String() string {
	return string(c)
}

// Validate checks if the correlation ID is valid UUIDv7.
//
// Enforces UUIDv7 format to ensure time-sortable guarantees.
func (c CorrelationID) Validate() error {
	if c == "" {
		return fmt.Errorf("correlation ID is empty")
	}

	parsed, err := uuid.Parse(string(c))
	if err != nil {
		return fmt.Errorf("invalid correlation ID format: %w", err)
	}

	// Enforce UUIDv7 for time-sortable guarantees
	if parsed.Version() != 7 {
		return fmt.Errorf("correlation ID must be UUIDv7, got UUID version %d", parsed.Version())
	}

	return nil
}

// IsValid returns true if the correlation ID is valid.
func (c CorrelationID) IsValid() bool {
	return c.Validate() == nil
}

// MarshalText implements encoding.TextMarshaler for JSON, YAML, TOML support.
func (c CorrelationID) MarshalText() ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return []byte(c), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON, YAML, TOML support.
func (c *CorrelationID) UnmarshalText(text []byte) error {
	parsed, err := ParseCorrelationIDValue(string(text))
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}
