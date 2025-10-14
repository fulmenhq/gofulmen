package foundry

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestDetectMimeTypeFromReader_JSON tests JSON detection from streaming data
func TestDetectMimeTypeFromReader_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Object", `{"key": "value"}`, "json"},
		{"Array", `["item1", "item2"]`, "json"},
		{"WithWhitespace", `  {"key": "value"}`, "json"},
		{"WithBOM", "\xEF\xBB\xBF{\"key\": \"value\"}", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))

			mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
			if err != nil {
				t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
			}

			if mimeType == nil {
				t.Fatal("Expected non-nil MIME type for JSON")
			}

			if mimeType.ID != tt.expected {
				t.Errorf("Expected MIME type %q, got %q", tt.expected, mimeType.ID)
			}

			// Verify reader preservation: read all data back
			data, err := io.ReadAll(newReader)
			if err != nil {
				t.Fatalf("Failed to read from new reader: %v", err)
			}

			if string(data) != tt.input {
				t.Errorf("Reader data mismatch:\nGot:  %q\nWant: %q", string(data), tt.input)
			}
		})
	}
}

// TestDetectMimeTypeFromReader_XML tests XML detection
func TestDetectMimeTypeFromReader_XML(t *testing.T) {
	input := `<?xml version="1.0"?><root><item>data</item></root>`
	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type for XML")
	}

	if mimeType.ID != "xml" {
		t.Errorf("Expected MIME type 'xml', got %q", mimeType.ID)
	}

	// Verify reader preservation
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_YAML tests YAML detection
func TestDetectMimeTypeFromReader_YAML(t *testing.T) {
	input := "name: John Doe\nage: 30\ncity: New York\n"
	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type for YAML")
	}

	if mimeType.ID != "yaml" {
		t.Errorf("Expected MIME type 'yaml', got %q", mimeType.ID)
	}

	// Verify reader preservation
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_CSV tests CSV detection
func TestDetectMimeTypeFromReader_CSV(t *testing.T) {
	input := "name,age,city\nJohn,30,NYC\nJane,25,LA\n"
	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type for CSV")
	}

	if mimeType.ID != "csv" {
		t.Errorf("Expected MIME type 'csv', got %q", mimeType.ID)
	}

	// Verify reader preservation
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_PlainText tests plain text detection
func TestDetectMimeTypeFromReader_PlainText(t *testing.T) {
	input := "This is plain text content without any special formatting."
	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type for plain text")
	}

	if mimeType.ID != "plain-text" {
		t.Errorf("Expected MIME type 'plain-text', got %q", mimeType.ID)
	}

	// Verify reader preservation
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_EmptyReader tests detection with empty input
func TestDetectMimeTypeFromReader_EmptyReader(t *testing.T) {
	reader := bytes.NewReader([]byte{})

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType != nil {
		t.Errorf("Expected nil MIME type for empty reader, got %q", mimeType.ID)
	}

	// Verify reader can still be read (should be empty)
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty reader, got %d bytes", len(data))
	}
}

// TestDetectMimeTypeFromReader_ShortReader tests with less than maxBytes available
func TestDetectMimeTypeFromReader_ShortReader(t *testing.T) {
	input := `{"short":"json"}` // Less than 512 bytes
	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json', got %q", mimeType.ID)
	}

	// Verify all data preserved
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_DefaultMaxBytes tests default maxBytes behavior
func TestDetectMimeTypeFromReader_DefaultMaxBytes(t *testing.T) {
	input := `{"key": "value"}`
	reader := bytes.NewReader([]byte(input))

	// Pass 0 to test default maxBytes (512)
	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 0)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json', got %q", mimeType.ID)
	}

	// Verify data preserved
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch")
	}
}

// TestDetectMimeTypeFromReader_LargeReader tests with reader larger than maxBytes
func TestDetectMimeTypeFromReader_LargeReader(t *testing.T) {
	// Create input larger than maxBytes
	header := `{"large": "json", "data": [`
	largeData := make([]byte, 1000) // Much larger than detection buffer
	for i := range largeData {
		largeData[i] = 'x'
	}
	footer := `]}`
	input := header + string(largeData) + footer

	reader := bytes.NewReader([]byte(input))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 100) // Small maxBytes
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json', got %q", mimeType.ID)
	}

	// Verify ALL data preserved (including data beyond maxBytes)
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != input {
		t.Errorf("Reader data mismatch: expected %d bytes, got %d bytes", len(input), len(data))
	}
}

// TestDetectMimeTypeFromReader_HTTPBody simulates HTTP request body detection
func TestDetectMimeTypeFromReader_HTTPBody(t *testing.T) {
	// Simulate HTTP request body with JSON
	requestBody := `{"username": "alice", "action": "login"}`
	reader := io.NopCloser(bytes.NewReader([]byte(requestBody)))

	mimeType, newReader, err := DetectMimeTypeFromReader(reader, 512)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromReader() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json', got %q", mimeType.ID)
	}

	// Simulate processing the request body after detection
	data, err := io.ReadAll(newReader)
	if err != nil {
		t.Fatalf("Failed to read from new reader: %v", err)
	}

	if string(data) != requestBody {
		t.Errorf("Request body mismatch")
	}
}

// TestDetectMimeTypeFromFile tests file-based MIME detection
func TestDetectMimeTypeFromFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"JSON", `{"key": "value"}`, "json"},
		{"XML", `<?xml version="1.0"?><root/>`, "xml"},
		{"YAML", "name: test\nvalue: 123\n", "yaml"},
		{"CSV", "col1,col2,col3\nval1,val2,val3\n", "csv"},
		{"PlainText", "This is plain text.", "plain-text"},
	}

	// Create temp directory for test files
	tempDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			filename := filepath.Join(tempDir, "test_"+tt.name)
			err := os.WriteFile(filename, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Detect MIME type
			mimeType, err := DetectMimeTypeFromFile(filename)
			if err != nil {
				t.Fatalf("DetectMimeTypeFromFile() error: %v", err)
			}

			if mimeType == nil {
				t.Fatal("Expected non-nil MIME type")
			}

			if mimeType.ID != tt.expected {
				t.Errorf("Expected MIME type %q, got %q", tt.expected, mimeType.ID)
			}
		})
	}
}

// TestDetectMimeTypeFromFile_NotFound tests detection with non-existent file
func TestDetectMimeTypeFromFile_NotFound(t *testing.T) {
	_, err := DetectMimeTypeFromFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestDetectMimeTypeFromFile_EmptyFile tests detection with empty file
func TestDetectMimeTypeFromFile_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "empty.txt")

	// Create empty file
	err := os.WriteFile(filename, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	mimeType, err := DetectMimeTypeFromFile(filename)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromFile() error: %v", err)
	}

	if mimeType != nil {
		t.Errorf("Expected nil MIME type for empty file, got %q", mimeType.ID)
	}
}

// TestDetectMimeTypeFromFile_LargeFile tests detection with large file
func TestDetectMimeTypeFromFile_LargeFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "large.json")

	// Create large JSON file (only beginning matters for detection)
	header := `{"large": "file", "data": [`
	largeData := make([]byte, 10000) // 10KB
	for i := range largeData {
		largeData[i] = 'x'
	}
	footer := `]}`
	content := header + string(largeData) + footer

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	mimeType, err := DetectMimeTypeFromFile(filename)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromFile() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json', got %q", mimeType.ID)
	}
}

// TestDetectMimeTypeFromFile_BOM tests BOM handling in files
func TestDetectMimeTypeFromFile_BOM(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "bom.json")

	// Create file with UTF-8 BOM
	content := "\xEF\xBB\xBF{\"key\": \"value\"}"
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create BOM file: %v", err)
	}

	mimeType, err := DetectMimeTypeFromFile(filename)
	if err != nil {
		t.Fatalf("DetectMimeTypeFromFile() error: %v", err)
	}

	if mimeType == nil {
		t.Fatal("Expected non-nil MIME type")
	}

	if mimeType.ID != "json" {
		t.Errorf("Expected MIME type 'json' (BOM stripped), got %q", mimeType.ID)
	}
}

// Benchmarks

func BenchmarkDetectMimeTypeFromReader_JSON(b *testing.B) {
	input := []byte(`{"key": "value", "array": [1, 2, 3]}`)

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(input)
		DetectMimeTypeFromReader(reader, 512)
	}
}

func BenchmarkDetectMimeTypeFromReader_LargeJSON(b *testing.B) {
	header := `{"large": "json", "data": [`
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = 'x'
	}
	footer := `]}`
	input := []byte(header + string(largeData) + footer)

	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(input)
		DetectMimeTypeFromReader(reader, 512)
	}
}

func BenchmarkDetectMimeTypeFromFile(b *testing.B) {
	tempDir := b.TempDir()
	filename := filepath.Join(tempDir, "bench.json")

	content := []byte(`{"key": "value"}`)
	err := os.WriteFile(filename, content, 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectMimeTypeFromFile(filename)
	}
}
