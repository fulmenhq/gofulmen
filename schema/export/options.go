package export

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Format represents the output format for exported schemas
type Format int

const (
	// FormatAuto detects format from file extension
	FormatAuto Format = iota
	// FormatJSON exports as JSON
	FormatJSON
	// FormatYAML exports as YAML
	FormatYAML
)

// String returns the string representation of the format
func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	case FormatYAML:
		return "yaml"
	case FormatAuto:
		return "auto"
	default:
		return "unknown"
	}
}

// ProvenanceStyle represents how provenance metadata is embedded
type ProvenanceStyle int

const (
	// ProvenanceObject adds x-crucible-source object to JSON root
	ProvenanceObject ProvenanceStyle = iota
	// ProvenanceComment adds provenance as $comment field
	ProvenanceComment
	// ProvenanceNone omits provenance metadata
	ProvenanceNone
)

// String returns the string representation of the provenance style
func (p ProvenanceStyle) String() string {
	switch p {
	case ProvenanceObject:
		return "object"
	case ProvenanceComment:
		return "comment"
	case ProvenanceNone:
		return "none"
	default:
		return "unknown"
	}
}

// IdentityProvider provides application identity information for provenance
type IdentityProvider interface {
	GetIdentity(ctx context.Context) (*Identity, error)
}

// Identity represents application identity for provenance metadata
type Identity struct {
	Vendor string `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	Binary string `json:"binary,omitempty" yaml:"binary,omitempty"`
}

// ExportOptions configures schema export behavior
type ExportOptions struct {
	// SchemaID is the Crucible schema identifier (e.g., "logging/v1.0.0/config")
	// This is REQUIRED.
	SchemaID string

	// OutPath is the destination file path where the schema will be written
	// This is REQUIRED.
	OutPath string

	// Format specifies the output format (JSON or YAML)
	// Default: FormatAuto (detects from file extension)
	Format Format

	// IncludeProvenance controls whether provenance metadata is included
	// Default: true
	IncludeProvenance bool

	// ProvenanceStyle controls how provenance is embedded
	// Default: ProvenanceObject
	ProvenanceStyle ProvenanceStyle

	// ValidateSchema controls whether the schema is validated before export
	// Default: true
	ValidateSchema bool

	// Overwrite controls whether existing files can be overwritten
	// Default: false (refuse to overwrite existing files)
	Overwrite bool

	// IdentityProvider optionally provides application identity for provenance
	// Default: nil (no identity information included)
	IdentityProvider IdentityProvider
}

// Validate checks that the export options are valid
func (o *ExportOptions) Validate() error {
	if o.SchemaID == "" {
		return fmt.Errorf("SchemaID is required")
	}

	if o.OutPath == "" {
		return fmt.Errorf("OutPath is required")
	}

	// Validate SchemaID format (basic check)
	parts := strings.Split(o.SchemaID, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid SchemaID format: expected 'category/version/name' or 'category/name', got %q", o.SchemaID)
	}

	return nil
}

// applyDefaults applies default values to unset options
func (o *ExportOptions) applyDefaults() {
	// Default format detection
	if o.Format == FormatAuto {
		ext := strings.ToLower(filepath.Ext(o.OutPath))
		switch ext {
		case ".yaml", ".yml":
			o.Format = FormatYAML
		case ".json":
			o.Format = FormatJSON
		default:
			// Default to JSON if extension is ambiguous
			o.Format = FormatJSON
		}
	}

	// Note: IncludeProvenance defaults to false (zero value)
	// This is intentional - we only include provenance if explicitly requested
	// or if the caller sets it to true

	// ProvenanceStyle defaults to ProvenanceObject (zero value is correct)
}

// NewExportOptions creates ExportOptions with defaults applied
func NewExportOptions(schemaID, outPath string) ExportOptions {
	opts := ExportOptions{
		SchemaID:          schemaID,
		OutPath:           outPath,
		Format:            FormatAuto,
		IncludeProvenance: true, // Default to including provenance
		ProvenanceStyle:   ProvenanceObject,
		ValidateSchema:    true, // Default to validation
		Overwrite:         false,
		IdentityProvider:  NewDefaultIdentityProvider(), // Use default identity provider
	}
	opts.applyDefaults()
	return opts
}
