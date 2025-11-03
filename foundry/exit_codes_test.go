package foundry

import (
	"testing"

	crucible "github.com/fulmenhq/crucible/foundry"
)

// TestExitCodeReexports verifies that our re-exported constants match Crucible's values.
func TestExitCodeReexports(t *testing.T) {
	tests := []struct {
		name          string
		gofulmenCode  ExitCode
		crucibleCode  int
		expectedValue int
	}{
		{"EXIT_SUCCESS", ExitSuccess, crucible.ExitSuccess, 0},
		{"EXIT_FAILURE", ExitFailure, crucible.ExitFailure, 1},
		{"EXIT_CONFIG_INVALID", ExitConfigInvalid, crucible.ExitConfigInvalid, 20},
		{"EXIT_PORT_IN_USE", ExitPortInUse, crucible.ExitPortInUse, 10},
		{"EXIT_CERTIFICATE_INVALID", ExitCertificateInvalid, crucible.ExitCertificateInvalid, 73},
		{"EXIT_SIGNAL_TERM", ExitSignalTerm, crucible.ExitSignalTerm, 143},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gofulmenCode != tt.expectedValue {
				t.Errorf("gofulmen constant %s = %d, want %d", tt.name, tt.gofulmenCode, tt.expectedValue)
			}
			if tt.crucibleCode != tt.expectedValue {
				t.Errorf("crucible constant %s = %d, want %d", tt.name, tt.crucibleCode, tt.expectedValue)
			}
			if tt.gofulmenCode != tt.crucibleCode {
				t.Errorf("mismatch: gofulmen=%d, crucible=%d", tt.gofulmenCode, tt.crucibleCode)
			}
		})
	}
}

// TestGetExitCodeInfo verifies metadata lookup by code number.
func TestGetExitCodeInfo(t *testing.T) {
	tests := []struct {
		code        ExitCode
		expectFound bool
		expectName  string
		expectCat   string
	}{
		{ExitSuccess, true, "EXIT_SUCCESS", "standard"},
		{ExitFailure, true, "EXIT_FAILURE", "standard"},
		{ExitConfigInvalid, true, "EXIT_CONFIG_INVALID", "configuration"},
		{ExitPortInUse, true, "EXIT_PORT_IN_USE", "networking"},
		{ExitSignalTerm, true, "EXIT_SIGNAL_TERM", "signals"},
		{999, false, "", ""}, // Non-existent code
	}

	for _, tt := range tests {
		t.Run(tt.expectName, func(t *testing.T) {
			info, found := GetExitCodeInfo(tt.code)
			if found != tt.expectFound {
				t.Errorf("GetExitCodeInfo(%d): found=%v, want %v", tt.code, found, tt.expectFound)
			}
			if !found {
				return
			}
			if info.Name != tt.expectName {
				t.Errorf("GetExitCodeInfo(%d): name=%q, want %q", tt.code, info.Name, tt.expectName)
			}
			if info.Category != tt.expectCat {
				t.Errorf("GetExitCodeInfo(%d): category=%q, want %q", tt.code, info.Category, tt.expectCat)
			}
			if info.Code != tt.code {
				t.Errorf("GetExitCodeInfo(%d): info.Code=%d, want %d", tt.code, info.Code, tt.code)
			}
			if info.Description == "" {
				t.Errorf("GetExitCodeInfo(%d): description is empty", tt.code)
			}
		})
	}
}

// TestLookupExitCode verifies metadata lookup by name.
func TestLookupExitCode(t *testing.T) {
	tests := []struct {
		name        string
		expectFound bool
		expectCode  int
		expectCat   string
	}{
		{"EXIT_SUCCESS", true, 0, "standard"},
		{"EXIT_FAILURE", true, 1, "standard"},
		{"EXIT_CONFIG_INVALID", true, 20, "configuration"},
		{"EXIT_PORT_IN_USE", true, 10, "networking"},
		{"EXIT_SIGNAL_TERM", true, 143, "signals"},
		{"EXIT_NONEXISTENT", false, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, found := LookupExitCode(tt.name)
			if found != tt.expectFound {
				t.Errorf("LookupExitCode(%q): found=%v, want %v", tt.name, found, tt.expectFound)
			}
			if !found {
				return
			}
			if info.Code != tt.expectCode {
				t.Errorf("LookupExitCode(%q): code=%d, want %d", tt.name, info.Code, tt.expectCode)
			}
			if info.Category != tt.expectCat {
				t.Errorf("LookupExitCode(%q): category=%q, want %q", tt.name, info.Category, tt.expectCat)
			}
			if info.Name != tt.name {
				t.Errorf("LookupExitCode(%q): name=%q, want %q", tt.name, info.Name, tt.name)
			}
		})
	}
}

// TestListExitCodes verifies listing all codes.
func TestListExitCodes(t *testing.T) {
	codes := ListExitCodes()

	if len(codes) == 0 {
		t.Fatal("ListExitCodes returned empty list")
	}

	// Verify sorting by code number
	for i := 1; i < len(codes); i++ {
		if codes[i].Code <= codes[i-1].Code {
			t.Errorf("codes not sorted: codes[%d]=%d, codes[%d]=%d",
				i-1, codes[i-1].Code, i, codes[i].Code)
		}
	}

	// Verify all entries have required fields
	for _, info := range codes {
		if info.Code < 0 {
			t.Errorf("invalid code: %d", info.Code)
		}
		if info.Name == "" {
			t.Errorf("code %d has empty name", info.Code)
		}
		if info.Category == "" {
			t.Errorf("code %d (%s) has empty category", info.Code, info.Name)
		}
		if info.Description == "" {
			t.Errorf("code %d (%s) has empty description", info.Code, info.Name)
		}
	}

	// Verify expected codes are present
	expectedCodes := []int{0, 1, 10, 20, 73, 143}
	foundCodes := make(map[int]bool)
	for _, info := range codes {
		foundCodes[info.Code] = true
	}
	for _, code := range expectedCodes {
		if !foundCodes[code] {
			t.Errorf("expected code %d not found in list", code)
		}
	}
}

// TestListExitCodesWithoutSignals verifies filtering signal codes on Windows.
func TestListExitCodesWithoutSignals(t *testing.T) {
	allCodes := ListExitCodes()
	filteredCodes := ListExitCodes(WithoutSignalCodes())

	// Filtered list should be smaller (unless no signal codes exist)
	if len(filteredCodes) > len(allCodes) {
		t.Errorf("filtered list larger than full list: %d > %d", len(filteredCodes), len(allCodes))
	}

	// Verify no signal codes (128-165) in filtered list
	for _, info := range filteredCodes {
		if info.Code >= 128 && info.Code <= 165 {
			t.Errorf("signal code %d (%s) found in filtered list", info.Code, info.Name)
		}
	}

	// Verify all filtered codes exist in full list
	allCodesMap := make(map[int]bool)
	for _, info := range allCodes {
		allCodesMap[info.Code] = true
	}
	for _, info := range filteredCodes {
		if !allCodesMap[info.Code] {
			t.Errorf("filtered code %d not in full list", info.Code)
		}
	}
}

// TestExitCodeInfoFields verifies specific metadata fields.
func TestExitCodeInfoFields(t *testing.T) {
	// Test EXIT_CONFIG_INVALID has retry hint
	info, found := GetExitCodeInfo(ExitConfigInvalid)
	if !found {
		t.Fatal("EXIT_CONFIG_INVALID not found")
	}
	if info.RetryHint == "" {
		t.Error("EXIT_CONFIG_INVALID missing retry hint")
	}

	// Test EXIT_PORT_IN_USE has BSD equivalent
	info, found = GetExitCodeInfo(ExitPortInUse)
	if !found {
		t.Fatal("EXIT_PORT_IN_USE not found")
	}
	// BSD equivalent may or may not be present - just verify field exists
	_ = info.BSDEquivalent

	// Test simplified mode mappings structure
	// (Actual presence depends on catalog content - will test mapping logic in simplified_modes_test.go)
	info, found = GetExitCodeInfo(ExitFailure)
	if !found {
		t.Fatal("EXIT_FAILURE not found")
	}
	// Simplified modes are optional - just verify the fields exist (may be nil)
	_ = info.SimplifiedBasic
	_ = info.SimplifiedSeverity
}

// TestCatalogLoadingFailure verifies error handling is present.
// This test documents expected behavior if catalog loading fails.
func TestCatalogLoadingBehavior(t *testing.T) {
	// The catalog loads in init() and panics on failure.
	// If we reach this test, the catalog loaded successfully.
	// This test verifies that behavior is consistent.

	if catalog == nil {
		t.Fatal("catalog is nil - should have been loaded in init()")
	}

	if loadErr != nil {
		t.Fatalf("catalog load error should be nil, got: %v", loadErr)
	}

	if len(codeInfoMap) == 0 {
		t.Error("codeInfoMap is empty after successful load")
	}

	if len(nameInfoMap) == 0 {
		t.Error("nameInfoMap is empty after successful load")
	}

	// Verify catalog has expected structure
	if catalog.Version == "" {
		t.Error("catalog version is empty")
	}
	if len(catalog.Categories) == 0 {
		t.Error("catalog has no categories")
	}
	if len(catalog.SimplifiedModes) == 0 {
		t.Error("catalog has no simplified modes")
	}
}

// BenchmarkGetExitCodeInfo measures metadata lookup performance.
func BenchmarkGetExitCodeInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetExitCodeInfo(ExitConfigInvalid)
	}
}

// BenchmarkLookupExitCode measures name lookup performance.
func BenchmarkLookupExitCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = LookupExitCode("EXIT_CONFIG_INVALID")
	}
}
