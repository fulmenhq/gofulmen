package foundry

import (
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

// IsValidCorrelationID checks if a string is a valid UUID (any version).
//
// Note: This validates UUID format but doesn't check if it's specifically UUIDv7.
// Use ParseCorrelationID and check the version field if you need strict v7 validation.
//
// Example:
//
//	if IsValidCorrelationID("018b2c5e-8f4a-7890-b123-456789abcdef") {
//	    // Valid UUID format
//	}
func IsValidCorrelationID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
