package export

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fulmenhq/gofulmen/foundry"
)

// ProvenanceMetadata contains metadata about the schema export
type ProvenanceMetadata struct {
	SchemaID        string    `json:"schema_id" yaml:"schema_id"`
	CrucibleVersion string    `json:"crucible_version" yaml:"crucible_version"`
	GofulmenVersion string    `json:"gofulmen_version" yaml:"gofulmen_version"`
	GitRevision     string    `json:"git_revision,omitempty" yaml:"git_revision,omitempty"`
	ExportedAt      time.Time `json:"exported_at" yaml:"exported_at"`
	Identity        *Identity `json:"identity,omitempty" yaml:"identity,omitempty"`
}

// buildProvenance creates provenance metadata for a schema export
func buildProvenance(ctx context.Context, opts ExportOptions) (*ProvenanceMetadata, error) {
	metadata := &ProvenanceMetadata{
		SchemaID:        opts.SchemaID,
		CrucibleVersion: foundry.CrucibleVersion(),
		GofulmenVersion: foundry.GofulmenVersion(),
		ExportedAt:      time.Now().UTC(),
	}

	// Add git revision if available (best-effort)
	if gitRev := getGitRevision(); gitRev != "" {
		metadata.GitRevision = gitRev
	}

	// Add identity if provider is available
	if opts.IdentityProvider != nil {
		identity, err := opts.IdentityProvider.GetIdentity(ctx)
		if err == nil && identity != nil {
			metadata.Identity = identity
		}
		// Silently ignore identity errors - it's optional metadata
	}

	return metadata, nil
}

// getGitRevision attempts to get the current git revision (short SHA)
// Returns empty string if git is unavailable or not in a git repository
func getGitRevision() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// formatProvenanceComment formats provenance metadata as a compact comment string
func formatProvenanceComment(metadata *ProvenanceMetadata) string {
	parts := []string{
		fmt.Sprintf("schema_id=%s", metadata.SchemaID),
		fmt.Sprintf("crucible=%s", metadata.CrucibleVersion),
		fmt.Sprintf("gofulmen=%s", metadata.GofulmenVersion),
	}

	if metadata.GitRevision != "" {
		parts = append(parts, fmt.Sprintf("git=%s", metadata.GitRevision))
	}

	parts = append(parts, fmt.Sprintf("exported=%s", metadata.ExportedAt.Format(time.RFC3339)))

	if metadata.Identity != nil {
		if metadata.Identity.Vendor != "" {
			parts = append(parts, fmt.Sprintf("vendor=%s", metadata.Identity.Vendor))
		}
		if metadata.Identity.Binary != "" {
			parts = append(parts, fmt.Sprintf("binary=%s", metadata.Identity.Binary))
		}
	}

	return "x-crucible-source: " + strings.Join(parts, " ")
}
