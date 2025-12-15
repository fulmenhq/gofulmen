package pathfinder

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectCIBoundaryHint_GitHubWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	start := filepath.Join(tempDir, "a", "b", "c")
	if err := os.MkdirAll(start, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKSPACE", tempDir)

	hint, ok := DetectCIBoundaryHint(start)
	if !ok {
		t.Fatal("expected hint")
	}
	if hint.Provider != "github" {
		t.Fatalf("expected provider github, got %q", hint.Provider)
	}
	if hint.Source != "GITHUB_WORKSPACE" {
		t.Fatalf("expected source GITHUB_WORKSPACE, got %q", hint.Source)
	}
	if filepath.Clean(hint.Boundary) != filepath.Clean(tempDir) {
		t.Fatalf("expected boundary %s, got %s", tempDir, hint.Boundary)
	}
}

func TestDetectCIBoundaryHint_PrefersFulmenWorkspaceRoot(t *testing.T) {
	tempDir := t.TempDir()
	start := filepath.Join(tempDir, "a", "b")
	if err := os.MkdirAll(start, 0755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKSPACE", "/should/not/win")
	t.Setenv("FULMEN_WORKSPACE_ROOT", tempDir)

	hint, ok := DetectCIBoundaryHint(start)
	if !ok {
		t.Fatal("expected hint")
	}
	if hint.Source != "FULMEN_WORKSPACE_ROOT" {
		t.Fatalf("expected source FULMEN_WORKSPACE_ROOT, got %q", hint.Source)
	}
}

func TestDetectCIBoundaryHint_RejectsNonAncestorWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	start := filepath.Join(tempDir, "a")
	if err := os.MkdirAll(start, 0755); err != nil {
		t.Fatal(err)
	}

	other := t.TempDir()

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKSPACE", other)

	_, ok := DetectCIBoundaryHint(start)
	if ok {
		t.Fatal("expected no hint")
	}
}

func TestDetectCIBoundaryHint_FileStartPath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "a", "b", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePath, []byte("ok"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKSPACE", tempDir)

	_, ok := DetectCIBoundaryHint(filePath)
	if !ok {
		t.Fatal("expected hint")
	}
}

func TestDetectCIBoundaryHint_RejectsFilesystemRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("filesystem root differs on windows")
	}

	start := t.TempDir()

	t.Setenv("CI", "true")
	t.Setenv("FULMEN_WORKSPACE_ROOT", "/")

	_, ok := DetectCIBoundaryHint(start)
	if ok {
		t.Fatal("expected no hint")
	}
}
