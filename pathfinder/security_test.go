package pathfinder

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindFiles_PathTraversalPrevention tests that path traversal attacks are blocked
func TestFindFiles_PathTraversalPrevention(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	// Create test structure in temp directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "safe")
	outsideDir := filepath.Join(tmpDir, "outside")

	// Create directories
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outsideDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files
	safeFile := filepath.Join(testDir, "safe.txt")
	if err := os.WriteFile(safeFile, []byte("safe"), 0644); err != nil {
		t.Fatal(err)
	}

	outsideFile := filepath.Join(outsideDir, "outside.txt")
	if err := os.WriteFile(outsideFile, []byte("outside"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		root            string
		include         []string
		shouldFindSafe  bool
		shouldNotEscape bool
	}{
		{
			name:            "normal pattern stays in root",
			root:            testDir,
			include:         []string{"*.txt"},
			shouldFindSafe:  true,
			shouldNotEscape: true,
		},
		{
			name:            "recursive pattern stays in root",
			root:            testDir,
			include:         []string{"**/*.txt"},
			shouldFindSafe:  true,
			shouldNotEscape: true,
		},
		{
			name:            "parent traversal blocked",
			root:            testDir,
			include:         []string{"../*.txt"},
			shouldFindSafe:  false,
			shouldNotEscape: true,
		},
		{
			name:            "deep parent traversal blocked",
			root:            testDir,
			include:         []string{"../../**/*.txt"},
			shouldFindSafe:  false, // Pattern attempts traversal but should be blocked
			shouldNotEscape: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := FindQuery{
				Root:    tt.root,
				Include: tt.include,
			}

			results, err := finder.FindFiles(ctx, query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			// Check if safe file was found
			foundSafe := false
			for _, result := range results {
				if strings.Contains(result.SourcePath, "safe.txt") {
					foundSafe = true
				}

				// CRITICAL: Ensure no results escape the root directory
				if tt.shouldNotEscape {
					absRoot, _ := filepath.Abs(tt.root)
					if err := ValidatePathWithinRoot(result.SourcePath, absRoot); err != nil {
						t.Errorf("Result escaped root: %s (root: %s)", result.SourcePath, absRoot)
					}

					// Double-check: outside.txt should NEVER appear in results
					if strings.Contains(result.SourcePath, "outside.txt") {
						t.Errorf("SECURITY VIOLATION: outside.txt found in results from root %s", tt.root)
					}
				}
			}

			if foundSafe != tt.shouldFindSafe {
				t.Errorf("foundSafe = %v, want %v", foundSafe, tt.shouldFindSafe)
			}
		})
	}
}

// TestFindFiles_HiddenDirectories tests that files under hidden directories are filtered
func TestFindFiles_HiddenDirectories(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	// Create test structure
	tmpDir := t.TempDir()

	// Create visible and hidden directory structures
	visibleDir := filepath.Join(tmpDir, "visible")
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	nestedHiddenDir := filepath.Join(visibleDir, ".secrets")

	if err := os.MkdirAll(visibleDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nestedHiddenDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create files
	visibleFile := filepath.Join(visibleDir, "visible.txt")
	hiddenFile := filepath.Join(hiddenDir, "hidden.txt")
	nestedHiddenFile := filepath.Join(nestedHiddenDir, "secret.txt")
	hiddenInVisible := filepath.Join(visibleDir, ".hidden.txt")

	for _, f := range []string{visibleFile, hiddenFile, nestedHiddenFile, hiddenInVisible} {
		if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name          string
		includeHidden bool
		shouldFind    []string
		shouldNotFind []string
	}{
		{
			name:          "exclude hidden - filters all hidden files and directories",
			includeHidden: false,
			shouldFind:    []string{"visible/visible.txt"},
			shouldNotFind: []string{".hidden/hidden.txt", "visible/.secrets/secret.txt", "visible/.hidden.txt"},
		},
		{
			name:          "include hidden - finds all files",
			includeHidden: true,
			shouldFind:    []string{"visible/visible.txt", ".hidden/hidden.txt", "visible/.secrets/secret.txt", "visible/.hidden.txt"},
			shouldNotFind: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := FindQuery{
				Root:          tmpDir,
				Include:       []string{"**/*.txt"},
				IncludeHidden: tt.includeHidden,
			}

			results, err := finder.FindFiles(ctx, query)
			if err != nil {
				t.Fatalf("FindFiles() error = %v", err)
			}

			// Build map of found relative paths
			found := make(map[string]bool)
			for _, result := range results {
				found[result.RelativePath] = true
			}

			// Check files that should be found
			for _, expected := range tt.shouldFind {
				if !found[expected] {
					t.Errorf("Expected to find %q but it was not in results", expected)
				}
			}

			// Check files that should NOT be found
			for _, notExpected := range tt.shouldNotFind {
				if found[notExpected] {
					t.Errorf("Should NOT find %q but it was in results", notExpected)
				}
			}
		})
	}
}

// TestFindFiles_MetadataPopulation tests that metadata is correctly populated
func TestFindFiles_MetadataPopulation(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	// Create test file with known content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("Hello, World!")

	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	query := FindQuery{
		Root:    tmpDir,
		Include: []string{"*.txt"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	// Verify metadata is not nil
	if result.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	// Check size metadata
	size, ok := result.Metadata["size"]
	if !ok {
		t.Error("Metadata missing 'size' field")
	} else {
		sizeInt, ok := size.(int64)
		if !ok {
			t.Errorf("size is not int64, got %T", size)
		} else if sizeInt != int64(len(testContent)) {
			t.Errorf("size = %d, want %d", sizeInt, len(testContent))
		}
	}

	// Check mtime metadata
	mtime, ok := result.Metadata["mtime"]
	if !ok {
		t.Error("Metadata missing 'mtime' field")
	} else {
		mtimeStr, ok := mtime.(string)
		if !ok {
			t.Errorf("mtime is not string, got %T", mtime)
		} else {
			// Verify it's a valid RFC3339Nano timestamp format
			if len(mtimeStr) < 20 {
				t.Errorf("mtime appears invalid: %s", mtimeStr)
			}
			if !strings.Contains(mtimeStr, "T") {
				t.Errorf("mtime not in RFC3339 format: %s", mtimeStr)
			}
		}
	}
}

// TestIgnoreMatcher_Basic tests basic .fulmenignore functionality
func TestIgnoreMatcher_Basic(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		ignored  bool
	}{
		{
			name:     "simple pattern matches",
			patterns: []string{"*.log"},
			path:     "test.log",
			ignored:  true,
		},
		{
			name:     "simple pattern no match",
			patterns: []string{"*.log"},
			path:     "test.txt",
			ignored:  false,
		},
		{
			name:     "directory pattern matches",
			patterns: []string{"node_modules/"},
			path:     "node_modules/package/file.js",
			ignored:  true,
		},
		{
			name:     "glob pattern matches nested",
			patterns: []string{"**/*.tmp"},
			path:     "deep/nested/file.tmp",
			ignored:  true,
		},
		{
			name:     "multiple patterns - first matches",
			patterns: []string{"*.log", "*.tmp"},
			path:     "test.log",
			ignored:  true,
		},
		{
			name:     "multiple patterns - second matches",
			patterns: []string{"*.log", "*.tmp"},
			path:     "test.tmp",
			ignored:  true,
		},
		{
			name:     "multiple patterns - none match",
			patterns: []string{"*.log", "*.tmp"},
			path:     "test.txt",
			ignored:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &IgnoreMatcher{
				root:     "/test",
				patterns: tt.patterns,
			}

			result := matcher.IsIgnored(tt.path)
			if result != tt.ignored {
				t.Errorf("IsIgnored(%q) = %v, want %v", tt.path, result, tt.ignored)
			}
		})
	}
}

// TestFindFiles_FulmenignoreIntegration tests .fulmenignore file integration
func TestFindFiles_FulmenignoreIntegration(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	// Create test structure
	tmpDir := t.TempDir()

	// Create .fulmenignore file
	ignoreContent := `# Test ignore file
*.log
*.tmp
node_modules/
build/
`
	ignoreFile := filepath.Join(tmpDir, ".fulmenignore")
	if err := os.WriteFile(ignoreFile, []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test files
	files := map[string]bool{
		"keep.txt":            true,  // Should be found
		"test.log":            false, // Should be ignored
		"data.tmp":            false, // Should be ignored
		"node_modules/pkg.js": false, // Should be ignored
		"build/output.txt":    false, // Should be ignored
		"src/main.txt":        true,  // Should be found
		"src/debug.log":       false, // Should be ignored
	}

	for path := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	query := FindQuery{
		Root:    tmpDir,
		Include: []string{"**/*"},
	}

	results, err := finder.FindFiles(ctx, query)
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	// Build map of found files
	found := make(map[string]bool)
	for _, result := range results {
		found[result.RelativePath] = true
	}

	// Verify expectations
	for path, shouldFind := range files {
		if shouldFind && !found[path] {
			t.Errorf("Expected to find %q but it was ignored", path)
		}
		if !shouldFind && found[path] {
			t.Errorf("Expected %q to be ignored but it was found", path)
		}
	}
}
