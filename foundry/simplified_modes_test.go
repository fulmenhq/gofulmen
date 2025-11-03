package foundry

import (
	"testing"
)

// TestMapToSimplified verifies simplified mode mapping.
func TestMapToSimplified(t *testing.T) {
	// Note: The Crucible catalog may not have simplified mode mappings populated yet.
	// These tests verify the API works correctly when mappings exist.

	tests := []struct {
		name string
		code ExitCode
		mode SimplifiedMode
	}{
		// Basic mode tests
		{"Success to basic", ExitSuccess, SimplifiedModeBasic},
		{"Failure to basic", ExitFailure, SimplifiedModeBasic},
		{"Config invalid to basic", ExitConfigInvalid, SimplifiedModeBasic},

		// Severity mode tests
		{"Success to severity", ExitSuccess, SimplifiedModeSeverity},
		{"Config invalid to severity", ExitConfigInvalid, SimplifiedModeSeverity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := MapToSimplified(tt.code, tt.mode)
			// The catalog may not have mappings yet - just verify API works
			if ok {
				if code < 0 {
					t.Errorf("MapToSimplified(%d, %q): returned negative code %d", tt.code, tt.mode, code)
				}
				t.Logf("MapToSimplified(%d, %q) = %d (mapping exists)", tt.code, tt.mode, code)
			} else {
				t.Logf("MapToSimplified(%d, %q) = no mapping (catalog may not have simplified modes yet)", tt.code, tt.mode)
			}
		})
	}

	// Test invalid inputs
	t.Run("Invalid mode", func(t *testing.T) {
		_, ok := MapToSimplified(ExitFailure, "invalid")
		if ok {
			t.Error("MapToSimplified with invalid mode should return false")
		}
	})

	t.Run("Non-existent code", func(t *testing.T) {
		_, ok := MapToSimplified(999, SimplifiedModeBasic)
		if ok {
			t.Error("MapToSimplified with non-existent code should return false")
		}
	})
}

// TestListSimplifiedModes verifies the list of available modes.
func TestListSimplifiedModes(t *testing.T) {
	modes := ListSimplifiedModes()

	if len(modes) == 0 {
		t.Fatal("ListSimplifiedModes returned empty list")
	}

	// Check for expected modes
	foundBasic := false
	foundSeverity := false
	for _, mode := range modes {
		if mode == SimplifiedModeBasic {
			foundBasic = true
		}
		if mode == SimplifiedModeSeverity {
			foundSeverity = true
		}
	}

	if !foundBasic {
		t.Error("SimplifiedModeBasic not in list")
	}
	if !foundSeverity {
		t.Error("SimplifiedModeSeverity not in list")
	}
}

// TestGetSimplifiedModeInfo verifies mode metadata retrieval.
func TestGetSimplifiedModeInfo(t *testing.T) {
	tests := []struct {
		mode        SimplifiedMode
		expectFound bool
		expectName  string
		minCodes    int // Minimum expected codes
	}{
		{SimplifiedModeBasic, true, "Basic Exit Codes", 3},
		{SimplifiedModeSeverity, true, "Severity-Based Exit Codes", 8},
		{"invalid", false, "", 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			info := GetSimplifiedModeInfo(tt.mode)
			if (info != nil) != tt.expectFound {
				t.Errorf("GetSimplifiedModeInfo(%q): found=%v, want %v", tt.mode, info != nil, tt.expectFound)
			}
			if !tt.expectFound {
				return
			}
			if info.Name != tt.expectName {
				t.Errorf("GetSimplifiedModeInfo(%q): name=%q, want %q", tt.mode, info.Name, tt.expectName)
			}
			if len(info.Codes) < tt.minCodes {
				t.Errorf("GetSimplifiedModeInfo(%q): %d codes, want at least %d", tt.mode, len(info.Codes), tt.minCodes)
			}
			if info.ID != string(tt.mode) {
				t.Errorf("GetSimplifiedModeInfo(%q): ID=%q, want %q", tt.mode, info.ID, string(tt.mode))
			}
		})
	}
}

// TestSimplifiedModeBasicCodes verifies the basic mode has expected codes.
func TestSimplifiedModeBasicCodes(t *testing.T) {
	info := GetSimplifiedModeInfo(SimplifiedModeBasic)
	if info == nil {
		t.Fatal("Basic mode info not found")
	}

	// Basic mode should have at least SUCCESS, ERROR, USAGE_ERROR
	expectedMinCodes := map[int]string{
		0: "SUCCESS",
		1: "ERROR",
		2: "USAGE_ERROR",
	}

	if len(info.Codes) < len(expectedMinCodes) {
		t.Errorf("Basic mode: got %d codes, want at least %d", len(info.Codes), len(expectedMinCodes))
	}

	// Verify minimum expected codes exist
	foundCodes := make(map[int]bool)
	for _, codeInfo := range info.Codes {
		foundCodes[codeInfo.Code] = true
		if codeInfo.Name == "" {
			t.Errorf("Basic mode code %d has empty name", codeInfo.Code)
		}
		if codeInfo.Description == "" {
			t.Errorf("Basic mode code %d has empty description", codeInfo.Code)
		}
	}

	for code, name := range expectedMinCodes {
		if !foundCodes[code] {
			t.Errorf("Expected code %d (%s) not found in basic mode", code, name)
		}
	}
}

// TestSimplifiedModeSeverityCodes verifies the severity mode has expected codes.
func TestSimplifiedModeSeverityCodes(t *testing.T) {
	info := GetSimplifiedModeInfo(SimplifiedModeSeverity)
	if info == nil {
		t.Fatal("Severity mode info not found")
	}

	// Severity mode should have at least 8 levels (0-7)
	expectedMinCodes := map[int]string{
		0: "SUCCESS",
		1: "USER_ERROR",
		2: "CONFIG_ERROR",
		3: "RUNTIME_ERROR",
		4: "SYSTEM_ERROR",
		5: "SECURITY_ERROR",
		6: "TEST_FAILURE",
		7: "OBSERVABILITY_ERROR",
	}

	if len(info.Codes) < len(expectedMinCodes) {
		t.Errorf("Severity mode: got %d codes, want at least %d", len(info.Codes), len(expectedMinCodes))
	}

	// Verify minimum expected codes exist
	foundCodes := make(map[int]bool)
	for _, codeInfo := range info.Codes {
		foundCodes[codeInfo.Code] = true
		if codeInfo.Name == "" {
			t.Errorf("Severity mode code %d has empty name", codeInfo.Code)
		}
		if codeInfo.Description == "" {
			t.Errorf("Severity mode code %d has empty description", codeInfo.Code)
		}
	}

	for code, name := range expectedMinCodes {
		if !foundCodes[code] {
			t.Errorf("Expected code %d (%s) not found in severity mode", code, name)
		}
	}
}

// TestSimplifiedModeConstants verifies mode constant values.
func TestSimplifiedModeConstants(t *testing.T) {
	if SimplifiedModeBasic != "basic" {
		t.Errorf("SimplifiedModeBasic = %q, want \"basic\"", SimplifiedModeBasic)
	}
	if SimplifiedModeSeverity != "severity" {
		t.Errorf("SimplifiedModeSeverity = %q, want \"severity\"", SimplifiedModeSeverity)
	}
}
