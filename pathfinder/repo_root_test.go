package pathfinder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepositoryRoot_GitMarker(t *testing.T) {
	// This test runs from pathfinder/ directory
	// We expect to find .git at the repo root
	root, err := FindRepositoryRoot(".", GitMarkers)
	if err != nil {
		t.Fatalf("FindRepositoryRoot failed: %v", err)
	}

	// Verify we found a directory
	if root == "" {
		t.Fatal("FindRepositoryRoot returned empty string")
	}

	// Verify .git exists in the found directory
	gitPath := filepath.Join(root, ".git")
	if _, err := os.Stat(gitPath); err != nil {
		t.Fatalf("Expected .git directory at %s, but got error: %v", gitPath, err)
	}

	t.Logf("Found repository root at: %s", root)
}

func TestFindRepositoryRoot_GoModMarker(t *testing.T) {
	// This test should find go.mod at the repo root
	root, err := FindRepositoryRoot(".", GoModMarkers)
	if err != nil {
		t.Fatalf("FindRepositoryRoot failed: %v", err)
	}

	// Verify go.mod exists
	goModPath := filepath.Join(root, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		t.Fatalf("Expected go.mod at %s, but got error: %v", goModPath, err)
	}

	t.Logf("Found Go module root at: %s", root)
}

func TestFindRepositoryRoot_InvalidStartPath(t *testing.T) {
	_, err := FindRepositoryRoot("/nonexistent/path/that/does/not/exist", GitMarkers)
	if err == nil {
		t.Fatal("Expected error for nonexistent start path, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

func TestFindRepositoryRoot_EmptyStartPath(t *testing.T) {
	_, err := FindRepositoryRoot("", GitMarkers)
	if err == nil {
		t.Fatal("Expected error for empty start path, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

func TestFindRepositoryRoot_EmptyMarkers(t *testing.T) {
	_, err := FindRepositoryRoot(".", []string{})
	if err == nil {
		t.Fatal("Expected error for empty markers, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

func TestFindRepositoryRoot_MaxDepth(t *testing.T) {
	// Try to find .git but limit depth to 1
	// This should fail if we're running from a subdirectory
	_, err := FindRepositoryRoot(".", GitMarkers, WithMaxDepth(1))
	// We might find it if we're at the root, or we might not
	// Just verify the function runs without panic
	t.Logf("Result with maxDepth=1: err=%v", err)
}

func TestFindRepositoryRoot_CustomBoundary(t *testing.T) {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Set boundary to current directory
	// This should fail to find .git since we can't traverse upward
	_, err = FindRepositoryRoot(".", GitMarkers, WithBoundary(cwd))
	if err == nil {
		// We might be AT the repo root, which is fine
		t.Log("Found marker at boundary (we're at repo root)")
	} else {
		t.Logf("Got expected error with custom boundary: %v", err)
	}
}

func TestFindRepositoryRoot_FromNestedDirectory(t *testing.T) {
	// Create a temporary nested directory structure
	tempDir := t.TempDir()

	// Create nested dirs
	nestedPath := filepath.Join(tempDir, "a", "b", "c", "d")
	if err := os.MkdirAll(nestedPath, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a marker file at the root
	markerPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(markerPath, []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Find from nested directory with custom boundary (tempDir parent)
	// This ensures we can traverse up to tempDir
	tempParent := filepath.Dir(tempDir)
	root, err := FindRepositoryRoot(nestedPath, GoModMarkers, WithBoundary(tempParent))
	if err != nil {
		t.Fatalf("FindRepositoryRoot failed: %v", err)
	}

	// Clean paths for comparison
	expectedRoot := filepath.Clean(tempDir)
	actualRoot := filepath.Clean(root)

	if actualRoot != expectedRoot {
		t.Fatalf("Expected root %s, got %s", expectedRoot, actualRoot)
	}

	t.Logf("Successfully found marker from 4 levels deep")
}

func TestFindRepositoryRoot_NoMarkerFound(t *testing.T) {
	// Create a temporary directory with no markers
	tempDir := t.TempDir()

	// Set boundary to tempDir so we don't traverse upward to real repo
	_, err := FindRepositoryRoot(tempDir, GitMarkers, WithBoundary(tempDir))
	if err == nil {
		t.Fatal("Expected error when no marker found, got nil")
	}

	t.Logf("Got expected error: %v", err)
}
