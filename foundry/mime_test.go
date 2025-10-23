package foundry

import (
	"testing"
)

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		expectedMime string
		expectNil    bool
	}{
		{
			name:         "JSON object",
			input:        []byte(`{"key": "value"}`),
			expectedMime: "application/json",
		},
		{
			name:         "JSON array",
			input:        []byte(`["item1", "item2"]`),
			expectedMime: "application/json",
		},
		{
			name:         "XML with declaration",
			input:        []byte(`<?xml version="1.0"?><root></root>`),
			expectedMime: "application/xml",
		},
		{
			name:         "YAML document",
			input:        []byte("key: value\nanother: data\n"),
			expectedMime: "application/yaml",
		},
		{
			name:         "CSV data",
			input:        []byte("name,age,city\nJohn,30,NYC\nJane,25,LA\n"),
			expectedMime: "text/csv",
		},
		{
			name:         "Plain text",
			input:        []byte("This is just plain text without special formatting."),
			expectedMime: "text/plain",
		},
		{
			name:      "Empty input",
			input:     []byte{},
			expectNil: true,
		},
		{
			name:      "Binary data",
			input:     []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE},
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := DetectMimeType(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectNil {
				if mimeType != nil {
					t.Errorf("Expected nil MIME type, got %v", mimeType)
				}
				return
			}

			if mimeType == nil {
				t.Fatal("Expected non-nil MIME type")
			}

			if mimeType.Mime != tt.expectedMime {
				t.Errorf("Expected MIME %q, got %q", tt.expectedMime, mimeType.Mime)
			}
		})
	}
}

func TestIsSupportedMimeType(t *testing.T) {
	tests := []struct {
		mime      string
		supported bool
	}{
		{"application/json", true},
		{"application/yaml", true},
		{"text/csv", true},
		{"text/plain", true},
		{"application/xml", true},
		{"application/x-ndjson", true},
		{"application/x-protobuf", true},
		{"application/octet-stream", false},
		{"image/png", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			result := IsSupportedMimeType(tt.mime)
			if result != tt.supported {
				t.Errorf("IsSupportedMimeType(%q) = %v, want %v", tt.mime, result, tt.supported)
			}
		})
	}
}

func TestGetMimeTypeByExtension(t *testing.T) {
	tests := []struct {
		extension    string
		expectedMime string
		expectNil    bool
	}{
		{"json", "application/json", false},
		{".json", "application/json", false},
		{"yaml", "application/yaml", false},
		{".yaml", "application/yaml", false},
		{"yml", "application/yaml", false},
		{"csv", "text/csv", false},
		{"xml", "application/xml", false},
		{"txt", "text/plain", false},
		{"unknown", "", true},
		{".unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			mimeType, err := GetMimeTypeByExtension(tt.extension)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectNil {
				if mimeType != nil {
					t.Errorf("Expected nil MIME type, got %v", mimeType)
				}
				return
			}

			if mimeType == nil {
				t.Fatal("Expected non-nil MIME type")
			}

			if mimeType.Mime != tt.expectedMime {
				t.Errorf("Expected MIME %q, got %q", tt.expectedMime, mimeType.Mime)
			}
		})
	}
}

func TestListMimeTypes(t *testing.T) {
	mimeTypes, err := ListMimeTypes()
	if err != nil {
		t.Fatalf("Failed to list MIME types: %v", err)
	}

	if len(mimeTypes) == 0 {
		t.Fatal("Expected at least one MIME type")
	}

	// Verify expected MIME types are in the list
	expectedMimes := map[string]bool{
		"application/json": false,
		"application/yaml": false,
		"text/csv":         false,
		"text/plain":       false,
		"application/xml":  false,
	}

	for _, mimeType := range mimeTypes {
		if _, exists := expectedMimes[mimeType.Mime]; exists {
			expectedMimes[mimeType.Mime] = true
		}
	}

	for mime, found := range expectedMimes {
		if !found {
			t.Errorf("Expected MIME type %q to be in the list", mime)
		}
	}
}

func TestMimeType_MatchesExtension(t *testing.T) {
	mimeType := &MimeType{
		ID:         "json",
		Mime:       "application/json",
		Extensions: []string{"json", "map"},
	}

	tests := []struct {
		extension string
		matches   bool
	}{
		{"json", true},
		{".json", true},
		{"JSON", true},
		{".JSON", true},
		{"map", true},
		{".map", true},
		{"yaml", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.extension, func(t *testing.T) {
			result := mimeType.MatchesExtension(tt.extension)
			if result != tt.matches {
				t.Errorf("MatchesExtension(%q) = %v, want %v", tt.extension, result, tt.matches)
			}
		})
	}
}

func TestMimeType_MatchesFilename(t *testing.T) {
	mimeType := &MimeType{
		ID:         "json",
		Mime:       "application/json",
		Extensions: []string{"json"},
	}

	tests := []struct {
		filename string
		matches  bool
	}{
		{"config.json", true},
		{"data.JSON", true},
		{"/path/to/file.json", true},
		{"config.yaml", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := mimeType.MatchesFilename(tt.filename)
			if result != tt.matches {
				t.Errorf("MatchesFilename(%q) = %v, want %v", tt.filename, result, tt.matches)
			}
		})
	}
}

func TestMimeType_GetPrimaryExtension(t *testing.T) {
	tests := []struct {
		name        string
		mimeType    *MimeType
		expectedExt string
	}{
		{
			name: "Single extension",
			mimeType: &MimeType{
				Extensions: []string{"json"},
			},
			expectedExt: "json",
		},
		{
			name: "Multiple extensions",
			mimeType: &MimeType{
				Extensions: []string{"yaml", "yml"},
			},
			expectedExt: "yaml",
		},
		{
			name: "No extensions",
			mimeType: &MimeType{
				Extensions: []string{},
			},
			expectedExt: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mimeType.GetPrimaryExtension()
			if result != tt.expectedExt {
				t.Errorf("GetPrimaryExtension() = %q, want %q", result, tt.expectedExt)
			}
		})
	}
}

func BenchmarkDetectMimeType_JSON(b *testing.B) {
	data := []byte(`{"key": "value"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DetectMimeType(data) //nolint:errcheck
	}
}

func BenchmarkDetectMimeType_YAML(b *testing.B) {
	data := []byte("key: value\nanother: data\n")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DetectMimeType(data) //nolint:errcheck
	}
}

func BenchmarkIsSupportedMimeType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsSupportedMimeType("application/json")
	}
}

func BenchmarkGetMimeTypeByExtension(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GetMimeTypeByExtension("json") //nolint:errcheck
	}
}
