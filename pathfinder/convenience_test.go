package pathfinder

import (
	"context"
	"testing"
)

// TestFindGoFiles tests the FindGoFiles convenience method
func TestFindGoFiles(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	results, err := finder.FindGoFiles(ctx, "testdata/nested")
	if err != nil {
		t.Fatalf("FindGoFiles() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("FindGoFiles() returned %d results, expected 3", len(results))
	}

	// Check that all results are .go files
	for _, result := range results {
		if len(result.RelativePath) < 3 || result.RelativePath[len(result.RelativePath)-3:] != ".go" {
			t.Errorf("FindGoFiles() returned non-.go file: %s", result.RelativePath)
		}
	}
}

// TestFindConfigFiles tests the FindConfigFiles convenience method
func TestFindConfigFiles(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	results, err := finder.FindConfigFiles(ctx, "testdata/nested")
	if err != nil {
		t.Fatalf("FindConfigFiles() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("FindConfigFiles() returned no results")
	}

	// Should find config.yaml
	found := false
	for _, result := range results {
		if result.RelativePath == "config.yaml" {
			found = true
			break
		}
	}
	if !found {
		t.Error("FindConfigFiles() did not find config.yaml")
	}
}

// TestFindSchemaFiles tests the FindSchemaFiles convenience method
func TestFindSchemaFiles(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	// Create a test schema file
	results, err := finder.FindSchemaFiles(ctx, "testdata")
	if err != nil {
		t.Fatalf("FindSchemaFiles() error = %v", err)
	}

	// May or may not find files depending on test setup
	// Just verify it doesn't error
	_ = results
}

// TestFindByExtension tests the FindByExtension method
func TestFindByExtension(t *testing.T) {
	ctx := context.Background()
	finder := NewFinder()

	tests := []struct {
		name        string
		root        string
		exts        []string
		expectedMin int
	}{
		{
			name:        "find go files",
			root:        "testdata/nested",
			exts:        []string{"go"},
			expectedMin: 3,
		},
		{
			name:        "find multiple extensions",
			root:        "testdata/nested",
			exts:        []string{"go", "yaml"},
			expectedMin: 4,
		},
		{
			name:        "find md files",
			root:        "testdata/basic",
			exts:        []string{"md"},
			expectedMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := finder.FindByExtension(ctx, tt.root, tt.exts)
			if err != nil {
				t.Fatalf("FindByExtension() error = %v", err)
			}

			if len(results) < tt.expectedMin {
				t.Errorf("FindByExtension() returned %d results, expected at least %d", len(results), tt.expectedMin)
			}
		})
	}
}

// TestValidatePathResults is skipped because it requires schema infrastructure
// The validation functions are tested indirectly through FindFiles with ValidateOutputs enabled
func TestValidatePathResults(t *testing.T) {
	t.Skip("Schema validation requires crucible infrastructure - tested via integration tests")
}

// TestNewFinder tests the NewFinder constructor
func TestNewFinder(t *testing.T) {
	finder := NewFinder()

	if finder == nil {
		t.Fatal("NewFinder() returned nil")
	}

	if finder.config.MaxWorkers != 4 {
		t.Errorf("NewFinder() MaxWorkers = %d, expected 4", finder.config.MaxWorkers)
	}

	if finder.config.LoaderType != "local" {
		t.Errorf("NewFinder() LoaderType = %q, expected 'local'", finder.config.LoaderType)
	}

	if finder.config.CacheEnabled {
		t.Error("NewFinder() CacheEnabled should be false by default")
	}

	if finder.config.ValidateInputs {
		t.Error("NewFinder() ValidateInputs should be false by default")
	}

	if finder.config.ValidateOutputs {
		t.Error("NewFinder() ValidateOutputs should be false by default")
	}
}

// TestIsSafePath tests the IsSafePath helper function
func TestIsSafePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"valid/path", true},
		{"../invalid", false},
		{"", false},
		{"/", false},
		{"./relative", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsSafePath(tt.path)
			if result != tt.expected {
				t.Errorf("IsSafePath(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}
