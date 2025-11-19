package crucible

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/ascii"
	"github.com/fulmenhq/gofulmen/pathfinder"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

// TestCrucibleVersionMatchesMetadata ensures go.mod and synced metadata agree.
//
// This test prevents the critical failure mode where:
// - make sync updates .crucible/metadata/metadata.yaml to a new Crucible version
// - But go.mod still points to the old version
// - Result: Embedded assets (docs/schemas/configs) are from v0.2.19 but runtime imports get v0.2.18
//
// After running 'make sync', you MUST also run:
//
//	go get github.com/fulmenhq/crucible@<new-version>
//	go mod tidy
//
// This test runs as part of 'make check-all' to catch this mistake before release.
func TestCrucibleVersionMatchesMetadata(t *testing.T) {
	// Find repository root using pathfinder (dogfooding our own library)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	root, err := pathfinder.FindRepositoryRoot(cwd, pathfinder.GoModMarkers)
	if err != nil {
		t.Fatalf("Failed to find repository root: %v", err)
	}

	// Read Crucible version from go.mod
	goModPath := filepath.Join(root, "go.mod")
	goModVersion, err := readCrucibleVersionFromGoMod(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod: %v", err)
	}

	// Read Crucible version from metadata
	metadataPath := filepath.Join(root, ".crucible", "metadata", "metadata.yaml")
	metadataVersion, err := readCrucibleVersionFromMetadata(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	// Normalize versions for comparison (metadata has no 'v' prefix, go.mod does)
	normalizedGoMod := strings.TrimPrefix(goModVersion, "v")
	normalizedMetadata := strings.TrimPrefix(metadataVersion, "v")

	// They must match
	if normalizedGoMod != normalizedMetadata {
		// Build error message using our ASCII module (dogfooding!)
		var msg strings.Builder
		msg.WriteString("CRITICAL: Crucible Version Mismatch Detected\n")
		msg.WriteString(strings.Repeat("═", 68) + "\n")
		msg.WriteString("\n")
		msg.WriteString(fmt.Sprintf("go.mod requires:     github.com/fulmenhq/crucible %s\n", goModVersion))
		msg.WriteString(fmt.Sprintf("metadata synced:     %s\n", metadataVersion))
		msg.WriteString("\n")
		msg.WriteString("This means your embedded assets (docs/schemas/configs) are from\n")
		msg.WriteString(fmt.Sprintf("Crucible %s but runtime imports will use %s\n", metadataVersion, goModVersion))
		msg.WriteString("\n")
		msg.WriteString(strings.Repeat("═", 68) + "\n")
		msg.WriteString("TO FIX:\n")
		msg.WriteString(strings.Repeat("═", 68) + "\n")
		msg.WriteString("\n")
		msg.WriteString("After running 'make sync', you must ALSO update go.mod:\n")
		msg.WriteString("\n")
		msg.WriteString(fmt.Sprintf("  go get github.com/fulmenhq/crucible@%s\n", metadataVersion))
		msg.WriteString("  go mod tidy\n")
		msg.WriteString("  make check-all\n")

		// Draw box using our ASCII module
		box := ascii.DrawBox(msg.String(), 0)
		t.Fatalf("\n%s", box)
	}

	t.Logf("✓ Crucible versions match: %s", goModVersion)
}

// readCrucibleVersionFromGoMod parses go.mod and extracts the Crucible version.
func readCrucibleVersionFromGoMod(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading go.mod: %w", err)
	}

	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", fmt.Errorf("parsing go.mod: %w", err)
	}

	for _, req := range modFile.Require {
		if req.Mod.Path == "github.com/fulmenhq/crucible" {
			return req.Mod.Version, nil
		}
	}

	return "", fmt.Errorf("crucible not found in go.mod requires")
}

// readCrucibleVersionFromMetadata reads the Crucible version from synced metadata.
func readCrucibleVersionFromMetadata(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading metadata: %w", err)
	}

	var metadata struct {
		Sources []struct {
			Name    string `yaml:"name"`
			Version string `yaml:"version"`
		} `yaml:"sources"`
	}

	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return "", fmt.Errorf("parsing metadata YAML: %w", err)
	}

	for _, src := range metadata.Sources {
		if src.Name == "crucible" {
			return src.Version, nil
		}
	}

	return "", fmt.Errorf("crucible source not found in metadata.yaml")
}
