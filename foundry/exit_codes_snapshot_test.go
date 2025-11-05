package foundry

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fulmenhq/crucible"
)

// snapshotExitCode represents an exit code entry in the canonical snapshot.
type snapshotExitCode struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Context     string `json:"context,omitempty"`
}

// snapshotData represents the structure of exit-codes.snapshot.json.
// The codes field is a map where keys are string representations of exit codes.
type snapshotData struct {
	Version string                      `json:"version"`
	Codes   map[string]snapshotExitCode `json:"codes"`
}

// TestSnapshotParity verifies that gofulmen's exit codes match Crucible's canonical snapshot.
// This test guards against drift between the catalog and our implementation.
func TestSnapshotParity(t *testing.T) {
	// Load the canonical snapshot from Crucible
	snapshotBytes, err := crucible.GetConfig("library/foundry/exit-codes.snapshot.json")
	if err != nil {
		t.Skipf("Snapshot file not available in Crucible v0.2.3: %v", err)
		return
	}

	// Parse the snapshot
	var snapshot snapshotData
	if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
		t.Fatalf("Failed to parse snapshot JSON: %v", err)
	}

	// Get all exit codes from gofulmen
	gofulmenCodes := ListExitCodes()

	// Build a map of gofulmen codes for comparison
	gofulmenMap := make(map[int]ExitCodeInfo)
	for _, info := range gofulmenCodes {
		gofulmenMap[info.Code] = info
	}

	// Verify all snapshot codes exist in gofulmen
	for codeStr, snapCode := range snapshot.Codes {
		// Parse code string to int
		var code int
		if _, err := fmt.Sscanf(codeStr, "%d", &code); err != nil {
			t.Fatalf("Invalid code key in snapshot: %q", codeStr)
		}

		t.Run(snapCode.Name, func(t *testing.T) {
			// NOTE: Crucible v0.2.5 has code mapping issues for USR1/USR2 signals.
			// Snapshot has codes 159/160 but catalog has 138/140.
			// Skip these codes until Crucible catalog is fixed.
			// See: /Users/davethompson/dev/fulmenhq/gofulmen/.plans/memos/crucible/20251105-exit-codes-snapshot-catalog-mismatch.md
			knownMissingCodes := map[int]bool{159: true, 160: true} // USR1, USR2
			if knownMissingCodes[code] {
				t.Skipf("Code %d (%s) has known Crucible catalog mapping issue", code, snapCode.Name)
				return
			}

			gofulmenInfo, found := gofulmenMap[code]
			if !found {
				t.Errorf("Snapshot code %d (%s) not found in gofulmen", code, snapCode.Name)
				return
			}

			// Verify name matches
			if gofulmenInfo.Name != snapCode.Name {
				t.Errorf("Code %d: name=%q, want %q", code, gofulmenInfo.Name, snapCode.Name)
			}

			// Verify category matches
			if gofulmenInfo.Category != snapCode.Category {
				t.Errorf("Code %d (%s): category=%q, want %q", code, snapCode.Name, gofulmenInfo.Category, snapCode.Category)
			}

			// Verify description matches
			// NOTE: Crucible v0.2.5 has a known inconsistency where signal descriptions in
			// exit-codes.yaml (verbose) don't match exit-codes.snapshot.json (simplified).
			// Skip description check for signal exit codes until Crucible catalog is updated.
			// See: https://github.com/fulmenhq/crucible/issues/TBD
			// Signal codes: 129 (HUP), 130 (INT), 131 (QUIT), 137 (KILL), 138 (USR1),
			// 140 (USR2), 141 (PIPE), 142 (ALRM), 143 (TERM)
			signalCodes := map[int]bool{129: true, 130: true, 131: true, 137: true, 138: true, 140: true, 141: true, 142: true, 143: true}
			if !signalCodes[code] && gofulmenInfo.Description != snapCode.Description {
				t.Errorf("Code %d (%s): description=%q, want %q", code, snapCode.Name, gofulmenInfo.Description, snapCode.Description)
			}
		})
	}

	// Verify snapshot version matches catalog version
	catalogVersion := ExitCodesVersion()
	if snapshot.Version != catalogVersion {
		t.Errorf("Snapshot version %q does not match catalog version %q", snapshot.Version, catalogVersion)
	}

	// Verify counts match
	if len(gofulmenCodes) != len(snapshot.Codes) {
		t.Errorf("Code count mismatch: gofulmen has %d codes, snapshot has %d", len(gofulmenCodes), len(snapshot.Codes))

		// List extra/missing codes
		snapshotMap := make(map[int]bool)
		for codeStr := range snapshot.Codes {
			var code int
			_, _ = fmt.Sscanf(codeStr, "%d", &code)
			snapshotMap[code] = true
		}

		for code := range gofulmenMap {
			if !snapshotMap[code] {
				t.Errorf("  Extra in gofulmen: code %d (%s)", code, gofulmenMap[code].Name)
			}
		}

		for codeStr, snap := range snapshot.Codes {
			var code int
			_, _ = fmt.Sscanf(codeStr, "%d", &code)
			if _, found := gofulmenMap[code]; !found {
				t.Errorf("  Missing from gofulmen: code %d (%s)", code, snap.Name)
			}
		}
	}
}

// TestSnapshotVersion verifies the snapshot version matches our catalog version.
func TestSnapshotVersion(t *testing.T) {
	snapshotBytes, err := crucible.GetConfig("library/foundry/exit-codes.snapshot.json")
	if err != nil {
		t.Skipf("Snapshot file not available in Crucible v0.2.3: %v", err)
		return
	}

	var snapshot snapshotData
	if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
		t.Fatalf("Failed to parse snapshot JSON: %v", err)
	}

	catalogVersion := ExitCodesVersion()
	if snapshot.Version != catalogVersion {
		t.Errorf("Snapshot version mismatch: got %q, want %q", snapshot.Version, catalogVersion)
	}
}
