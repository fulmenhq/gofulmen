package pathfinder

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindFiles_RecursiveGlob tests recursive glob pattern matching with **
func TestFindFiles_RecursiveGlob(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name     string
		query    FindQuery
		expected int
		contains []string
	}{
		{
			name: "recursive go files",
			query: FindQuery{
				Root:    "testdata/nested",
				Include: []string{"**/*.go"},
			},
			expected: 3,
			contains: []string{"top.go", "level1/mid.go", "level1/level2/deep.go"},
		},
		{
			name: "single level glob",
			query: FindQuery{
				Root:    "testdata/basic",
				Include: []string{"*.go"},
			},
			expected: 1,
			contains: []string{"file1.go"},
		},
		{
			name: "multiple patterns",
			query: FindQuery{
				Root:    "testdata/basic",
				Include: []string{"*.go", "*.txt"},
			},
			expected: 2,
			contains: []string{"file1.go", "file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := finder.FindFiles(ctx, tt.query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("FindFiles() returned %d results, expected %d", len(results), tt.expected)
			}

			for _, expected := range tt.contains {
				found := false
				for _, result := range results {
					if result.RelativePath == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected result %q not found in results", expected)
				}
			}
		})
	}
}

// TestFindFiles_Exclude tests exclude pattern filtering
func TestFindFiles_Exclude(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name     string
		query    FindQuery
		expected int
		excludes []string
	}{
		{
			name: "exclude test files",
			query: FindQuery{
				Root:    "testdata/mixed",
				Include: []string{"**/*.go"},
				Exclude: []string{"**/*_test.go"},
			},
			expected: 1,
			excludes: []string{"src/main_test.go"},
		},
		{
			name: "exclude hidden directories",
			query: FindQuery{
				Root:    "testdata/mixed",
				Include: []string{"**/*"},
				Exclude: []string{".tmp/**"},
			},
			expected: 3,
			excludes: []string{".tmp/temp.txt"},
		},
		{
			name: "multiple exclude patterns",
			query: FindQuery{
				Root:    "testdata/mixed",
				Include: []string{"**/*"},
				Exclude: []string{"**/*_test.go", ".tmp/**"},
			},
			expected: 2,
			excludes: []string{"src/main_test.go", ".tmp/temp.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := finder.FindFiles(ctx, tt.query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("FindFiles() returned %d results, expected %d", len(results), tt.expected)
				for _, r := range results {
					t.Logf("  - %s", r.RelativePath)
				}
			}

			for _, excluded := range tt.excludes {
				for _, result := range results {
					if result.RelativePath == excluded {
						t.Errorf("Excluded file %q found in results", excluded)
					}
				}
			}
		})
	}
}

// TestFindFiles_MaxDepth tests depth limiting
func TestFindFiles_MaxDepth(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name     string
		maxDepth int
		expected int
		contains []string
		excludes []string
	}{
		{
			name:     "max depth 1",
			maxDepth: 1,
			expected: 1,
			contains: []string{"top.go"},
			excludes: []string{"level1/mid.go", "level1/level2/deep.go"},
		},
		{
			name:     "max depth 2",
			maxDepth: 2,
			expected: 2,
			contains: []string{"top.go", "level1/mid.go"},
			excludes: []string{"level1/level2/deep.go"},
		},
		{
			name:     "max depth 0 (unlimited)",
			maxDepth: 0,
			expected: 3,
			contains: []string{"top.go", "level1/mid.go", "level1/level2/deep.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := FindQuery{
				Root:     "testdata/nested",
				Include:  []string{"**/*.go"},
				MaxDepth: tt.maxDepth,
			}

			results, err := finder.FindFiles(ctx, query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("FindFiles() returned %d results, expected %d", len(results), tt.expected)
				for _, r := range results {
					t.Logf("  - %s", r.RelativePath)
				}
			}

			for _, expected := range tt.contains {
				found := false
				for _, result := range results {
					if result.RelativePath == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected result %q not found", expected)
				}
			}

			for _, excluded := range tt.excludes {
				for _, result := range results {
					if result.RelativePath == excluded {
						t.Errorf("File %q should be excluded by MaxDepth=%d", excluded, tt.maxDepth)
					}
				}
			}
		})
	}
}

// TestFindFiles_FollowSymlinks tests symlink handling
func TestFindFiles_FollowSymlinks(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name           string
		followSymlinks bool
		expectedMin    int
		shouldContain  string
	}{
		{
			name:           "follow symlinks",
			followSymlinks: true,
			expectedMin:    2, // real.txt and link.txt
			shouldContain:  "link.txt",
		},
		{
			name:           "skip symlinks",
			followSymlinks: false,
			expectedMin:    1, // only real.txt
			shouldContain:  "real.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := FindQuery{
				Root:           "testdata/symlinks",
				Include:        []string{"*.txt"},
				FollowSymlinks: tt.followSymlinks,
			}

			results, err := finder.FindFiles(ctx, query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) < tt.expectedMin {
				t.Errorf("FindFiles() returned %d results, expected at least %d", len(results), tt.expectedMin)
			}

			found := false
			for _, result := range results {
				if result.RelativePath == tt.shouldContain {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected result %q not found", tt.shouldContain)
			}
		})
	}
}

// TestFindFiles_IncludeHidden tests hidden file handling
func TestFindFiles_IncludeHidden(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name          string
		includeHidden bool
		expected      int
		shouldContain []string
	}{
		{
			name:          "include hidden files",
			includeHidden: true,
			expected:      2,
			shouldContain: []string{".hidden.txt", "visible.txt"},
		},
		{
			name:          "exclude hidden files",
			includeHidden: false,
			expected:      1,
			shouldContain: []string{"visible.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := FindQuery{
				Root:          "testdata/hidden",
				Include:       []string{"*.txt"},
				IncludeHidden: tt.includeHidden,
			}

			results, err := finder.FindFiles(ctx, query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) != tt.expected {
				t.Errorf("FindFiles() returned %d results, expected %d", len(results), tt.expected)
				for _, r := range results {
					t.Logf("  - %s", r.RelativePath)
				}
			}

			for _, expected := range tt.shouldContain {
				found := false
				for _, result := range results {
					if result.RelativePath == expected {
						found = true
						break
					}
				}
				if !found && (tt.includeHidden || expected != ".hidden.txt") {
					t.Errorf("Expected result %q not found", expected)
				}
			}
		})
	}
}

// TestFindFiles_ErrorHandler tests error handler callback
func TestFindFiles_ErrorHandler(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	errorCount := 0
	errorHandler := func(path string, err error) error {
		errorCount++
		return nil // Continue processing
	}

	query := FindQuery{
		Root:         "testdata/basic",
		Include:      []string{"*.go", "[invalid"},
		ErrorHandler: errorHandler,
	}

	_, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if errorCount == 0 {
		t.Error("ErrorHandler was not called for invalid pattern")
	}
}

// TestFindFiles_ProgressCallback tests progress callback
func TestFindFiles_ProgressCallback(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	callbackCount := 0
	progressCallback := func(processed int, total int, currentPath string) {
		callbackCount++
	}

	query := FindQuery{
		Root:             "testdata/basic",
		Include:          []string{"*"},
		ProgressCallback: progressCallback,
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if callbackCount != len(results) {
		t.Errorf("ProgressCallback called %d times, expected %d", callbackCount, len(results))
	}
}

// TestFindFiles_ContextCancellation tests context cancellation
func TestFindFiles_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	finder := NewFinder()
	query := FindQuery{
		Root:    "testdata/nested",
		Include: []string{"**/*"},
	}

	_, err := finder.FindFiles(ctx, query)
	if err != context.Canceled {
		t.Errorf("FindFiles() error = %v, expected context.Canceled", err)
	}
}

// TestFindFiles_EmptyResults tests queries that match nothing
func TestFindFiles_EmptyResults(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	query := FindQuery{
		Root:    "testdata/basic",
		Include: []string{"*.nonexistent"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("FindFiles() returned %d results, expected 0", len(results))
	}
}

// TestFindFiles_PathResult tests PathResult structure
func TestFindFiles_PathResult(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	query := FindQuery{
		Root:    "testdata/basic",
		Include: []string{"*.go"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("FindFiles() returned no results")
	}

	result := results[0]

	// Check RelativePath
	if result.RelativePath == "" {
		t.Error("PathResult.RelativePath is empty")
	}

	// Check SourcePath is absolute
	if !filepath.IsAbs(result.SourcePath) {
		t.Errorf("PathResult.SourcePath %q is not absolute", result.SourcePath)
	}

	// Check LogicalPath matches RelativePath
	if result.LogicalPath != result.RelativePath {
		t.Errorf("PathResult.LogicalPath %q != RelativePath %q", result.LogicalPath, result.RelativePath)
	}

	// Check LoaderType
	if result.LoaderType != "local" {
		t.Errorf("PathResult.LoaderType = %q, expected 'local'", result.LoaderType)
	}

	// Check Metadata is initialized
	if result.Metadata == nil {
		t.Error("PathResult.Metadata is nil")
	}
}

// TestFindFiles_Checksums tests checksum calculation functionality
func TestFindFiles_Checksums(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		query           FindQuery
		expectChecksum  bool
		expectAlgorithm bool
		expectError     bool
	}{
		{
			name: "checksums disabled",
			query: FindQuery{
				Root:               "testdata/basic",
				Include:            []string{"*.go"},
				CalculateChecksums: false,
			},
			expectChecksum:  false,
			expectAlgorithm: false,
			expectError:     false,
		},
		{
			name: "checksums enabled xxh3-128",
			query: FindQuery{
				Root:               "testdata/basic",
				Include:            []string{"*.go"},
				CalculateChecksums: true,
				ChecksumAlgorithm:  "xxh3-128",
			},
			expectChecksum:  true,
			expectAlgorithm: true,
			expectError:     false,
		},
		{
			name: "checksums enabled sha256",
			query: FindQuery{
				Root:               "testdata/basic",
				Include:            []string{"*.go"},
				CalculateChecksums: true,
				ChecksumAlgorithm:  "sha256",
			},
			expectChecksum:  true,
			expectAlgorithm: true,
			expectError:     false,
		},
		{
			name: "checksums enabled default algorithm",
			query: FindQuery{
				Root:               "testdata/basic",
				Include:            []string{"*.go"},
				CalculateChecksums: true,
				// ChecksumAlgorithm empty - should default to xxh3-128
			},
			expectChecksum:  true,
			expectAlgorithm: true,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder()

			results, err := finder.FindFiles(ctx, tt.query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			if len(results) == 0 {
				t.Fatal("FindFiles() returned no results")
			}

			for _, result := range results {
				// Check metadata exists
				if result.Metadata == nil {
					t.Error("PathResult.Metadata is nil")
					continue
				}

				// Check checksum field
				checksum, hasChecksum := result.Metadata["checksum"]
				if tt.expectChecksum && !hasChecksum {
					t.Errorf("Expected checksum field but not found")
				}
				if !tt.expectChecksum && hasChecksum {
					t.Errorf("Unexpected checksum field: %v", checksum)
				}

				// Check checksumAlgorithm field
				algorithm, hasAlgorithm := result.Metadata["checksumAlgorithm"]
				if tt.expectAlgorithm && !hasAlgorithm {
					t.Errorf("Expected checksumAlgorithm field but not found")
				}
				if !tt.expectAlgorithm && hasAlgorithm {
					t.Errorf("Unexpected checksumAlgorithm field: %v", algorithm)
				}

				// Check checksumError field
				checksumError, hasError := result.Metadata["checksumError"]
				if tt.expectError && !hasError {
					t.Errorf("Expected checksumError field but not found")
				}
				if !tt.expectError && hasError {
					t.Errorf("Unexpected checksumError field: %v", checksumError)
				}

				// Validate checksum format if present
				if hasChecksum {
					checksumStr, ok := checksum.(string)
					if !ok {
						t.Errorf("Checksum is not a string: %T", checksum)
						continue
					}

					// Should match pattern: algorithm:hex
					if !strings.Contains(checksumStr, ":") {
						t.Errorf("Checksum format invalid, expected 'algorithm:hex', got: %s", checksumStr)
					}
				}

				// Validate algorithm value if present
				if hasAlgorithm {
					algStr, ok := algorithm.(string)
					if !ok {
						t.Errorf("ChecksumAlgorithm is not a string: %T", algorithm)
						continue
					}

					if algStr != "xxh3-128" && algStr != "sha256" {
						t.Errorf("Invalid checksumAlgorithm: %s", algStr)
					}
				}
			}
		})
	}
}
