package foundry

import (
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
			name:     "Valid UUIDv4",
			input:    uuid.New().String(),
			expected: true,
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

func BenchmarkGenerateCorrelationID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateCorrelationID()
	}
}

func BenchmarkParseCorrelationID(b *testing.B) {
	id := GenerateCorrelationID()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseCorrelationID(id)
	}
}

func BenchmarkIsValidCorrelationID(b *testing.B) {
	id := GenerateCorrelationID()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		IsValidCorrelationID(id)
	}
}
