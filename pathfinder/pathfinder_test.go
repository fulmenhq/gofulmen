package pathfinder

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"valid/path", true},
		{"../invalid", false},
		{"", false},
		{"/", false},
	}

	for _, test := range tests {
		err := ValidatePath(test.path)
		valid := err == nil
		if valid != test.expected {
			t.Errorf("ValidatePath(%q) = %v, expected %v", test.path, valid, test.expected)
		}
	}
}
