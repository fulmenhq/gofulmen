package foundry

import (
	"testing"
)

// TestMapToBSD verifies mapping Fulmen codes to BSD.
func TestMapToBSD(t *testing.T) {
	// Note: The Crucible catalog may not have BSD mappings populated yet.
	// These tests verify the API works correctly when mappings exist.

	tests := []struct {
		name string
		code ExitCode
	}{
		{"Success", ExitSuccess},
		{"Config invalid", ExitConfigInvalid},
		{"Usage", ExitUsage},
		{"Permission denied", ExitPermissionDenied},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bsdCode, ok := MapToBSD(tt.code)
			if ok {
				if bsdCode < 0 {
					t.Errorf("MapToBSD(%d): returned negative code %d", tt.code, bsdCode)
				}
				// Verify it's a valid BSD code
				info := GetBSDCodeInfo(bsdCode)
				if info == nil {
					t.Errorf("MapToBSD(%d) returned %d which is not a valid BSD code", tt.code, bsdCode)
				} else {
					t.Logf("MapToBSD(%d) = %d (%s: %s)", tt.code, bsdCode, info.Name, info.Description)
				}
			} else {
				t.Logf("MapToBSD(%d) = no mapping (catalog may not have BSD mappings yet)", tt.code)
			}
		})
	}

	// Test invalid code
	t.Run("Non-existent code", func(t *testing.T) {
		_, ok := MapToBSD(999)
		if ok {
			t.Error("MapToBSD with non-existent code should return false")
		}
	})
}

// TestMapFromBSD verifies mapping BSD codes to Fulmen.
func TestMapFromBSD(t *testing.T) {
	tests := []struct {
		name    string
		bsdCode int
		bsdName string
	}{
		{"EX_OK", 0, "EX_OK"},
		{"EX_USAGE", 64, "EX_USAGE"},
		{"EX_CONFIG", 78, "EX_CONFIG"},
		{"EX_NOPERM", 77, "EX_NOPERM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fulmenCode, ok := MapFromBSD(tt.bsdCode)
			if ok {
				if fulmenCode < 0 {
					t.Errorf("MapFromBSD(%d): returned negative code %d", tt.bsdCode, fulmenCode)
				}
				// Verify round-trip
				bsdCodeBack, okBack := MapToBSD(fulmenCode)
				if okBack && bsdCodeBack != tt.bsdCode {
					t.Errorf("Round-trip failed: BSD %d -> Fulmen %d -> BSD %d", tt.bsdCode, fulmenCode, bsdCodeBack)
				}
				t.Logf("MapFromBSD(%d/%s) = %d", tt.bsdCode, tt.bsdName, fulmenCode)
			} else {
				t.Logf("MapFromBSD(%d/%s) = no mapping (catalog may not have BSD mappings yet)", tt.bsdCode, tt.bsdName)
			}
		})
	}

	// Test invalid BSD code
	t.Run("Invalid BSD code", func(t *testing.T) {
		_, ok := MapFromBSD(999)
		if ok {
			t.Error("MapFromBSD with invalid BSD code should return false")
		}
	})
}

// TestGetBSDCodeInfo verifies BSD code metadata retrieval.
func TestGetBSDCodeInfo(t *testing.T) {
	tests := []struct {
		bsdCode     int
		expectFound bool
		expectName  string
	}{
		{0, true, "EX_OK"},
		{64, true, "EX_USAGE"},
		{78, true, "EX_CONFIG"},
		{77, true, "EX_NOPERM"},
		{999, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expectName, func(t *testing.T) {
			info := GetBSDCodeInfo(tt.bsdCode)
			if (info != nil) != tt.expectFound {
				t.Errorf("GetBSDCodeInfo(%d): found=%v, want %v", tt.bsdCode, info != nil, tt.expectFound)
			}
			if !tt.expectFound {
				return
			}
			if info.Name != tt.expectName {
				t.Errorf("GetBSDCodeInfo(%d): name=%q, want %q", tt.bsdCode, info.Name, tt.expectName)
			}
			if info.Code != tt.bsdCode {
				t.Errorf("GetBSDCodeInfo(%d): code=%d, want %d", tt.bsdCode, info.Code, tt.bsdCode)
			}
			if info.Description == "" {
				t.Errorf("GetBSDCodeInfo(%d): description is empty", tt.bsdCode)
			}
		})
	}
}

// TestBSDNameToCodeMapping verifies the bsdNameToCode table.
func TestBSDNameToCodeMapping(t *testing.T) {
	// Verify critical BSD codes
	expectedCodes := map[string]int{
		"EX_OK":          0,
		"EX_USAGE":       64,
		"EX_DATAERR":     65,
		"EX_CONFIG":      78,
		"EX_NOPERM":      77,
		"EX_UNAVAILABLE": 69,
	}

	for name, expectedCode := range expectedCodes {
		if code, ok := bsdNameToCode[name]; !ok {
			t.Errorf("BSD code %s missing from bsdNameToCode", name)
		} else if code != expectedCode {
			t.Errorf("BSD code %s: got %d, want %d", name, code, expectedCode)
		}
	}
}

// TestBSDDescriptions verifies all BSD codes have descriptions.
func TestBSDDescriptions(t *testing.T) {
	for name := range bsdNameToCode {
		if desc, ok := bsdDescriptions[name]; !ok {
			t.Errorf("BSD code %s missing description", name)
		} else if desc == "" {
			t.Errorf("BSD code %s has empty description", name)
		}
	}
}
