package foundry

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateCorrelationID(t *testing.T) {
	id := GenerateCorrelationID()

	// Verify it's a valid UUID format
	if !IsValidCorrelationID(id) {
		t.Errorf("Generated ID is not valid UUID: %s", id)
	}

	// Verify it has the standard UUID format (8-4-4-4-12)
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Errorf("Expected 5 UUID parts, got %d: %s", len(parts), id)
	}

	// Verify length of each part
	expectedLengths := []int{8, 4, 4, 4, 12}
	for i, part := range parts {
		if len(part) != expectedLengths[i] {
			t.Errorf("Part %d has length %d, expected %d: %s", i, len(part), expectedLengths[i], id)
		}
	}
}

func TestGenerateCorrelationID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and verify they're unique
	ids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		id := GenerateCorrelationID()
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestGenerateCorrelationID_TimeSortable(t *testing.T) {
	// Generate IDs with time gaps and verify they're sortable
	id1 := GenerateCorrelationID()
	time.Sleep(2 * time.Millisecond)
	id2 := GenerateCorrelationID()
	time.Sleep(2 * time.Millisecond)
	id3 := GenerateCorrelationID()

	// UUIDv7 should be lexicographically sortable by time
	if id1 >= id2 {
		t.Errorf("ID1 should be less than ID2: %s >= %s", id1, id2)
	}
	if id2 >= id3 {
		t.Errorf("ID2 should be less than ID3: %s >= %s", id2, id3)
	}
}

func TestMustGenerateCorrelationID(t *testing.T) {
	// Verify alias function works
	id := MustGenerateCorrelationID()

	if !IsValidCorrelationID(id) {
		t.Errorf("Generated ID is not valid UUID: %s", id)
	}
}

func TestParseCorrelationID(t *testing.T) {
	id := GenerateCorrelationID()
	parsed, err := ParseCorrelationID(id)

	if err != nil {
		t.Errorf("Failed to parse correlation ID: %v", err)
	}

	// Verify parsed UUID string matches original
	if parsed.String() != id {
		t.Errorf("Parsed UUID %s doesn't match original %s", parsed.String(), id)
	}
}

func TestParseCorrelationID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty string", ""},
		{"Invalid format", "not-a-uuid"},
		{"Partial UUID", "123e4567-e89b-12d3"},
		{"Invalid characters", "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCorrelationID(tt.input)
			if err == nil {
				t.Errorf("Expected error parsing %q, but got none", tt.input)
			}
		})
	}
}

func TestIsValidCorrelationID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid UUIDv7",
			input:    GenerateCorrelationID(),
			expected: true,
		},
		{
			name:     "Invalid UUIDv4 (must be v7)",
			input:    uuid.New().String(), // UUIDv4 is rejected
			expected: false,
		},
		{
			name:     "Invalid format",
			input:    "not-a-uuid",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Partial UUID",
			input:    "123e4567-e89b-12d3",
			expected: false,
		},
		{
			name:     "Too short",
			input:    "123e4567",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidCorrelationID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidCorrelationID(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsValidCorrelationID_RejectsNonV7 verifies that only UUIDv7 is accepted
func TestIsValidCorrelationID_RejectsNonV7(t *testing.T) {
	// Generate UUIDv4 (random)
	uuidv4 := uuid.New() // This is v4
	if IsValidCorrelationID(uuidv4.String()) {
		t.Error("IsValidCorrelationID should reject UUIDv4")
	}

	// Generate UUIDv7
	uuidv7 := uuid.Must(uuid.NewV7())
	if !IsValidCorrelationID(uuidv7.String()) {
		t.Error("IsValidCorrelationID should accept UUIDv7")
	}
}

func BenchmarkGenerateCorrelationID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateCorrelationID()
	}
}

func BenchmarkParseCorrelationID(b *testing.B) {
	id := GenerateCorrelationID()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = ParseCorrelationID(id)
	}
}

func BenchmarkIsValidCorrelationID(b *testing.B) {
	id := GenerateCorrelationID()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsValidCorrelationID(id)
	}
}

// Tests for CorrelationID newtype

func TestNewCorrelationIDValue(t *testing.T) {
	id := NewCorrelationIDValue()

	if !id.IsValid() {
		t.Error("NewCorrelationIDValue() generated invalid ID")
	}

	if len(id.String()) != 36 {
		t.Errorf("NewCorrelationIDValue() length = %d, want 36", len(id.String()))
	}
}

func TestParseCorrelationIDValue(t *testing.T) {
	validID := "018b2c5e-8f4a-7890-b123-456789abcdef"
	id, err := ParseCorrelationIDValue(validID)
	if err != nil {
		t.Fatalf("ParseCorrelationIDValue() error: %v", err)
	}

	if id.String() != validID {
		t.Errorf("ParseCorrelationIDValue() = %q, want %q", id.String(), validID)
	}
}

func TestParseCorrelationIDValue_Invalid(t *testing.T) {
	_, err := ParseCorrelationIDValue("invalid")
	if err == nil {
		t.Error("ParseCorrelationIDValue() should return error for invalid input")
	}
}

// TestParseCorrelationIDValue_RejectsNonV7 verifies that non-v7 UUIDs are rejected
func TestParseCorrelationIDValue_RejectsNonV7(t *testing.T) {
	// UUIDv4 should be rejected
	uuidv4 := uuid.New()
	_, err := ParseCorrelationIDValue(uuidv4.String())
	if err == nil {
		t.Error("ParseCorrelationIDValue should reject UUIDv4")
	}
	if !strings.Contains(err.Error(), "UUIDv7") {
		t.Errorf("Error message should mention UUIDv7 requirement, got: %v", err)
	}

	// UUIDv7 should be accepted
	uuidv7 := uuid.Must(uuid.NewV7())
	_, err = ParseCorrelationIDValue(uuidv7.String())
	if err != nil {
		t.Errorf("ParseCorrelationIDValue should accept UUIDv7, got error: %v", err)
	}
}

func TestCorrelationIDValue_JSONRoundTrip(t *testing.T) {
	type Request struct {
		CorrelationID CorrelationID `json:"correlation_id"`
		Data          string        `json:"data"`
	}

	original := Request{
		CorrelationID: NewCorrelationIDValue(),
		Data:          "test",
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal
	var decoded Request
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Compare
	if decoded.CorrelationID.String() != original.CorrelationID.String() {
		t.Errorf("JSON round-trip failed: got %q, want %q", decoded.CorrelationID, original.CorrelationID)
	}
}

// TestCorrelationIDValue_MarshalText_Error tests MarshalText with invalid ID
func TestCorrelationIDValue_MarshalText_Error(t *testing.T) {
	// Create invalid correlation ID (empty)
	invalidID := CorrelationID("")

	_, err := invalidID.MarshalText()
	if err == nil {
		t.Error("MarshalText should return error for empty correlation ID")
	}

	// Create invalid correlation ID (non-UUID)
	invalidID2 := CorrelationID("not-a-uuid")

	_, err = invalidID2.MarshalText()
	if err == nil {
		t.Error("MarshalText should return error for invalid correlation ID format")
	}

	// Create invalid correlation ID (UUIDv4)
	invalidID3 := CorrelationID(uuid.New().String())

	_, err = invalidID3.MarshalText()
	if err == nil {
		t.Error("MarshalText should return error for UUIDv4 correlation ID")
	}
}

func TestCorrelationIDValue_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      CorrelationID
		wantErr bool
	}{
		{"Valid UUIDv7", CorrelationID(uuid.Must(uuid.NewV7()).String()), false},
		{"Empty", "", true},
		{"Invalid format", "not-a-uuid", true},
		{"UUIDv4 rejected", CorrelationID(uuid.New().String()), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCorrelationIDValue_Validate_RejectsNonV7 verifies Validate rejects non-v7 UUIDs
func TestCorrelationIDValue_Validate_RejectsNonV7(t *testing.T) {
	// UUIDv4 should be rejected
	uuidv4 := CorrelationID(uuid.New().String())
	err := uuidv4.Validate()
	if err == nil {
		t.Error("Validate should reject UUIDv4")
	}
	if !strings.Contains(err.Error(), "UUIDv7") {
		t.Errorf("Error message should mention UUIDv7 requirement, got: %v", err)
	}

	// UUIDv7 should be accepted
	uuidv7 := CorrelationID(uuid.Must(uuid.NewV7()).String())
	err = uuidv7.Validate()
	if err != nil {
		t.Errorf("Validate should accept UUIDv7, got error: %v", err)
	}
}

func BenchmarkNewCorrelationIDValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCorrelationIDValue()
	}
}

func BenchmarkCorrelationIDValue_Validate(b *testing.B) {
	id := NewCorrelationIDValue()
	for i := 0; i < b.N; i++ {
		_ = id.Validate()
	}
}
