package foundry

import (
	"testing"
)

func TestNewCatalog(t *testing.T) {
	catalog := NewCatalog()
	if catalog == nil {
		t.Fatal("Expected non-nil catalog")
	}
}

func TestGetDefaultCatalog(t *testing.T) {
	catalog := GetDefaultCatalog()
	if catalog == nil {
		t.Fatal("Expected non-nil default catalog")
	}

	// Calling again should return same instance
	catalog2 := GetDefaultCatalog()
	if catalog != catalog2 {
		t.Error("Expected GetDefaultCatalog to return same instance")
	}
}

func TestCatalog_GetPattern(t *testing.T) {
	catalog := GetDefaultCatalog()

	// Test loading a known pattern from Crucible
	pattern, err := catalog.GetPattern("slug")
	if err != nil {
		t.Fatalf("Failed to get pattern: %v", err)
	}

	if pattern == nil {
		t.Fatal("Expected non-nil pattern for 'slug'")
	}

	if pattern.ID != "slug" {
		t.Errorf("Expected pattern ID 'slug', got %q", pattern.ID)
	}

	if pattern.Kind != PatternKindRegex {
		t.Errorf("Expected pattern kind 'regex', got %q", pattern.Kind)
	}

	// Test matching
	if matched, _ := pattern.Match("valid-slug"); !matched {
		t.Error("Expected 'valid-slug' to match slug pattern")
	}

	if matched, _ := pattern.Match("Invalid Slug"); matched {
		t.Error("Expected 'Invalid Slug' to not match slug pattern")
	}
}

func TestCatalog_GetPattern_NotFound(t *testing.T) {
	catalog := GetDefaultCatalog()

	pattern, err := catalog.GetPattern("non-existent-pattern")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if pattern != nil {
		t.Error("Expected nil pattern for non-existent ID")
	}
}

func TestCatalog_GetAllPatterns(t *testing.T) {
	catalog := GetDefaultCatalog()

	patterns, err := catalog.GetAllPatterns()
	if err != nil {
		t.Fatalf("Failed to get all patterns: %v", err)
	}

	if len(patterns) == 0 {
		t.Error("Expected at least one pattern")
	}

	// Verify some known patterns exist
	knownPatterns := []string{"slug", "ansi-email", "identifier"}
	for _, id := range knownPatterns {
		if _, exists := patterns[id]; !exists {
			t.Errorf("Expected pattern %q to exist", id)
		}
	}
}

func TestCatalog_GetMimeType(t *testing.T) {
	catalog := GetDefaultCatalog()

	// Test loading JSON MIME type
	mimeType, err := catalog.GetMimeType("json")
	if err != nil {
		t.Fatalf("Failed to get MIME type: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type for 'json'")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type ID 'json', got %q", mimeType.ID)
	}

	if mimeType.Mime != "application/json" {
		t.Errorf("Expected MIME 'application/json', got %q", mimeType.Mime)
	}

	// Verify extensions
	if !mimeType.MatchesExtension("json") {
		t.Error("Expected JSON MIME type to match .json extension")
	}
}

func TestCatalog_GetMimeType_NotFound(t *testing.T) {
	catalog := GetDefaultCatalog()

	mimeType, err := catalog.GetMimeType("non-existent-mime")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mimeType != nil {
		t.Error("Expected nil MIME type for non-existent ID")
	}
}

func TestCatalog_GetMimeTypeByExtension(t *testing.T) {
	catalog := GetDefaultCatalog()

	tests := []struct {
		extension    string
		expectedMime string
	}{
		{"json", "application/json"},
		{".json", "application/json"},
		{"yaml", "application/yaml"},
		{".yaml", "application/yaml"},
		{"yml", "application/yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			mimeType, err := catalog.GetMimeTypeByExtension(tt.extension)
			if err != nil {
				t.Fatalf("Failed to get MIME type: %v", err)
			}

			if mimeType == nil {
				t.Fatalf("Expected non-nil MIME type for extension %q", tt.extension)
			}

			if mimeType.Mime != tt.expectedMime {
				t.Errorf("Expected MIME %q, got %q", tt.expectedMime, mimeType.Mime)
			}
		})
	}
}

// TestCatalog_GetMimeTypeByExtension_NotFound tests non-existent extension
func TestCatalog_GetMimeTypeByExtension_NotFound(t *testing.T) {
	catalog := GetDefaultCatalog()

	mimeType, err := catalog.GetMimeTypeByExtension("nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mimeType != nil {
		t.Error("Expected nil MIME type for non-existent extension")
	}
}

func TestCatalog_GetHTTPStatusGroup(t *testing.T) {
	catalog := GetDefaultCatalog()

	group, err := catalog.GetHTTPStatusGroup("success")
	if err != nil {
		t.Fatalf("Failed to get HTTP status group: %v", err)
	}

	if group == nil {
		t.Fatal("Expected non-nil group for 'success'")
	}

	if group.ID != "success" {
		t.Errorf("Expected group ID 'success', got %q", group.ID)
	}

	// Verify it contains known success codes
	if !group.Contains(200) {
		t.Error("Expected success group to contain 200")
	}

	if !group.Contains(201) {
		t.Error("Expected success group to contain 201")
	}

	if group.Contains(404) {
		t.Error("Expected success group to not contain 404")
	}
}

// TestCatalog_GetHTTPStatusGroup_NotFound tests non-existent group
func TestCatalog_GetHTTPStatusGroup_NotFound(t *testing.T) {
	catalog := GetDefaultCatalog()

	group, err := catalog.GetHTTPStatusGroup("non-existent-group")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if group != nil {
		t.Error("Expected nil group for non-existent ID")
	}
}

func TestCatalog_GetHTTPStatusGroupForCode(t *testing.T) {
	catalog := GetDefaultCatalog()

	tests := []struct {
		code         int
		expectedID   string
		expectedName string
	}{
		{200, "success", "Successful Responses"},
		{404, "client-error", "Client Error Responses"},
		{500, "server-error", "Server Error Responses"},
		{301, "redirect", "Redirection Responses"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			group, err := catalog.GetHTTPStatusGroupForCode(tt.code)
			if err != nil {
				t.Fatalf("Failed to get HTTP status group: %v", err)
			}

			if group == nil {
				t.Fatalf("Expected non-nil group for code %d", tt.code)
			}

			if group.ID != tt.expectedID {
				t.Errorf("Expected group ID %q, got %q", tt.expectedID, group.ID)
			}

			if group.Name != tt.expectedName {
				t.Errorf("Expected group name %q, got %q", tt.expectedName, group.Name)
			}
		})
	}
}

// TestCatalog_GetHTTPStatusGroupForCode_NotFound tests invalid status code
func TestCatalog_GetHTTPStatusGroupForCode_NotFound(t *testing.T) {
	catalog := GetDefaultCatalog()

	group, err := catalog.GetHTTPStatusGroupForCode(999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if group != nil {
		t.Error("Expected nil group for invalid status code 999")
	}
}

func TestCatalog_GetHTTPStatusHelper(t *testing.T) {
	catalog := GetDefaultCatalog()

	helper, err := catalog.GetHTTPStatusHelper()
	if err != nil {
		t.Fatalf("Failed to get HTTP status helper: %v", err)
	}

	if helper == nil {
		t.Fatal("Expected non-nil helper")
	}

	// Test helper methods
	if !helper.IsSuccess(200) {
		t.Error("Expected 200 to be success")
	}

	if !helper.IsClientError(404) {
		t.Error("Expected 404 to be client error")
	}

	if !helper.IsServerError(500) {
		t.Error("Expected 500 to be server error")
	}

	if helper.IsSuccess(404) {
		t.Error("Expected 404 to not be success")
	}
}

func TestCatalog_GetAllMimeTypes(t *testing.T) {
	catalog := GetDefaultCatalog()

	mimeTypes, err := catalog.GetAllMimeTypes()
	if err != nil {
		t.Fatalf("Failed to get all MIME types: %v", err)
	}

	if len(mimeTypes) == 0 {
		t.Error("Expected at least one MIME type")
	}

	// Verify some known MIME types exist
	knownTypes := []string{"json", "yaml", "xml", "csv", "plain-text"}
	for _, id := range knownTypes {
		if _, exists := mimeTypes[id]; !exists {
			t.Errorf("Expected MIME type %q to exist", id)
		}
	}
}

func TestCatalog_GetAllHTTPStatusGroups(t *testing.T) {
	catalog := GetDefaultCatalog()

	groups, err := catalog.GetAllHTTPStatusGroups()
	if err != nil {
		t.Fatalf("Failed to get all HTTP status groups: %v", err)
	}

	if len(groups) == 0 {
		t.Error("Expected at least one HTTP status group")
	}

	// Verify known groups exist
	knownGroups := map[string]bool{
		"informational": false,
		"success":       false,
		"redirect":      false,
		"client-error":  false,
		"server-error":  false,
	}

	for _, group := range groups {
		if _, exists := knownGroups[group.ID]; exists {
			knownGroups[group.ID] = true
		}
	}

	for groupID, found := range knownGroups {
		if !found {
			t.Errorf("Expected HTTP status group %q to exist", groupID)
		}
	}
}

func TestHTTPStatusGroup_GetReason(t *testing.T) {
	group := &HTTPStatusGroup{
		ID:   "success",
		Name: "Successful Responses",
		Codes: []HTTPStatusCode{
			{Value: 200, Reason: "OK"},
			{Value: 201, Reason: "Created"},
		},
	}

	tests := []struct {
		code           int
		expectedReason string
	}{
		{200, "OK"},
		{201, "Created"},
		{404, ""}, // Not in this group
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			reason := group.GetReason(tt.code)
			if reason != tt.expectedReason {
				t.Errorf("Expected reason %q for code %d, got %q", tt.expectedReason, tt.code, reason)
			}
		})
	}
}

func TestHTTPStatusHelper_IsInformational(t *testing.T) {
	catalog := GetDefaultCatalog()
	helper, err := catalog.GetHTTPStatusHelper()
	if err != nil {
		t.Fatalf("Failed to get HTTP status helper: %v", err)
	}

	tests := []struct {
		code     int
		expected bool
	}{
		{100, true},  // Continue
		{101, true},  // Switching Protocols
		{200, false}, // Success, not informational
		{404, false}, // Client error, not informational
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := helper.IsInformational(tt.code)
			if result != tt.expected {
				t.Errorf("IsInformational(%d) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestHTTPStatusHelper_IsRedirect(t *testing.T) {
	catalog := GetDefaultCatalog()
	helper, err := catalog.GetHTTPStatusHelper()
	if err != nil {
		t.Fatalf("Failed to get HTTP status helper: %v", err)
	}

	tests := []struct {
		code     int
		expected bool
	}{
		{301, true},  // Moved Permanently
		{302, true},  // Found
		{304, true},  // Not Modified
		{200, false}, // Success, not redirect
		{404, false}, // Client error, not redirect
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			result := helper.IsRedirect(tt.code)
			if result != tt.expected {
				t.Errorf("IsRedirect(%d) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestHTTPStatusHelper_GetReasonPhrase(t *testing.T) {
	catalog := GetDefaultCatalog()
	helper, err := catalog.GetHTTPStatusHelper()
	if err != nil {
		t.Fatalf("Failed to get HTTP status helper: %v", err)
	}

	tests := []struct {
		code           int
		expectedReason string
	}{
		{200, "OK"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{999, ""}, // Unknown code
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			reason := helper.GetReasonPhrase(tt.code)
			if reason != tt.expectedReason {
				t.Errorf("GetReasonPhrase(%d) = %q, want %q", tt.code, reason, tt.expectedReason)
			}
		})
	}
}

func TestHTTPStatusHelper_GetGroup(t *testing.T) {
	catalog := GetDefaultCatalog()
	helper, err := catalog.GetHTTPStatusHelper()
	if err != nil {
		t.Fatalf("Failed to get HTTP status helper: %v", err)
	}

	tests := []struct {
		code       int
		expectedID string
	}{
		{200, "success"},
		{404, "client-error"},
		{500, "server-error"},
		{301, "redirect"},
		{100, "informational"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.code)), func(t *testing.T) {
			group := helper.GetGroup(tt.code)
			if group == nil {
				t.Fatalf("Expected non-nil group for code %d", tt.code)
			}
			if group.ID != tt.expectedID {
				t.Errorf("GetGroup(%d).ID = %q, want %q", tt.code, group.ID, tt.expectedID)
			}
		})
	}

	// Test unknown code
	group := helper.GetGroup(999)
	if group != nil {
		t.Error("Expected nil group for unknown code 999")
	}
}

func BenchmarkCatalog_GetPattern(b *testing.B) {
	catalog := GetDefaultCatalog()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		catalog.GetPattern("slug")
	}
}

func BenchmarkCatalog_GetMimeType(b *testing.B) {
	catalog := GetDefaultCatalog()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		catalog.GetMimeType("json")
	}
}
