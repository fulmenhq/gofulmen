// Package export provides functionality to export schemas from the Crucible SSOT
// with provenance metadata and validation.
package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/fulmenhq/gofulmen/crucible"
	"github.com/fulmenhq/gofulmen/schema"
)

// Export exports a schema from Crucible to a file with optional provenance and validation
func Export(ctx context.Context, opts ExportOptions) error {
	// Validate options
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("invalid export options: %w", err)
	}

	// Apply defaults
	opts.applyDefaults()

	// Load schema from Crucible
	schemaData, err := crucible.GetSchema(opts.SchemaID)
	if err != nil {
		return fmt.Errorf("%w: %q: %v", ErrSchemaNotFound, opts.SchemaID, err)
	}

	// Validate schema if requested
	if opts.ValidateSchema {
		if err := validateSchemaData(schemaData); err != nil {
			return fmt.Errorf("%w: %v", ErrSchemaValidation, err)
		}
	}

	// Build provenance metadata if requested
	var metadata *ProvenanceMetadata
	if opts.IncludeProvenance {
		metadata, err = buildProvenance(ctx, opts)
		if err != nil {
			return fmt.Errorf("failed to build provenance: %w", err)
		}
	}

	// Format the schema based on output format
	var formattedData []byte
	switch opts.Format {
	case FormatJSON:
		formattedData, err = formatJSON(schemaData, metadata, opts.ProvenanceStyle)
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}

	case FormatYAML:
		formattedData, err = formatYAML(schemaData, metadata, opts.ProvenanceStyle)
		if err != nil {
			return fmt.Errorf("failed to format YAML: %w", err)
		}

	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// Write the formatted schema to file
	if err := writeFileSafe(opts.OutPath, formattedData, opts.Overwrite); err != nil {
		// Preserve specific error types (ErrFileExists, ErrPathValidation)
		if errors.Is(err, ErrFileExists) || errors.Is(err, ErrPathValidation) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrFileWrite, err)
	}

	return nil
}

// validateSchemaData validates that schema data is valid JSON Schema
func validateSchemaData(data []byte) error {
	// Create a validator from the schema data
	validator, err := schema.NewValidator(data)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// The fact that we could create a validator means the schema is valid
	// (NewValidator parses and validates the JSON Schema structure)
	_ = validator // validator is only used for validation existence

	return nil
}

// ValidateExportedSchema validates that an exported schema file matches its source
// This function:
// 1. Validates the exported schema structure (skipped if references are unresolved)
// 2. Strips provenance metadata from the exported schema
// 3. Compares the payload byte-for-byte with the source from Crucible
func ValidateExportedSchema(ctx context.Context, schemaID, filePath string) error {
	return validateExportedSchemaWithOptions(ctx, schemaID, filePath, true)
}

// validateExportedSchemaWithOptions validates with optional structure validation
func validateExportedSchemaWithOptions(ctx context.Context, schemaID, filePath string, validateStructure bool) error {
	// Load the exported file
	exportedData, err := schema.LoadSchemaFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load exported schema: %w", err)
	}

	// Validate the exported schema structure (optional, may fail with unresolved refs)
	if validateStructure {
		if err := validateSchemaData(exportedData); err != nil {
			return fmt.Errorf("exported schema validation failed: %w", err)
		}
	}

	// Load the source schema from Crucible
	sourceData, err := crucible.GetSchema(schemaID)
	if err != nil {
		return fmt.Errorf("failed to load source schema: %w", err)
	}

	// Validate that the source schema is also valid (optional)
	if validateStructure {
		if err := validateSchemaData(sourceData); err != nil {
			return fmt.Errorf("source schema validation failed: %w", err)
		}
	}

	// Strip provenance from exported schema and compare with source
	if err := compareSchemaPayloads(exportedData, sourceData); err != nil {
		return fmt.Errorf("schema content mismatch: %w", err)
	}

	return nil
}

// compareSchemaPayloads strips provenance metadata and compares schema payloads
func compareSchemaPayloads(exportedData, sourceData []byte) error {
	// Parse exported schema
	var exportedSchema map[string]interface{}
	if err := json.Unmarshal(exportedData, &exportedSchema); err != nil {
		return fmt.Errorf("failed to parse exported schema: %w", err)
	}

	// Parse source schema
	var sourceSchema map[string]interface{}
	if err := json.Unmarshal(sourceData, &sourceSchema); err != nil {
		return fmt.Errorf("failed to parse source schema: %w", err)
	}

	// Strip provenance fields from exported schema
	delete(exportedSchema, "x-crucible-source")
	// Also strip $comment if it contains provenance
	if comment, ok := exportedSchema["$comment"].(string); ok {
		if strings.HasPrefix(comment, "x-crucible-source:") {
			delete(exportedSchema, "$comment")
		}
	}

	// Marshal both to canonical JSON for comparison
	exportedCanonical, err := json.Marshal(exportedSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal exported schema: %w", err)
	}

	sourceCanonical, err := json.Marshal(sourceSchema)
	if err != nil {
		return fmt.Errorf("failed to marshal source schema: %w", err)
	}

	// Compare byte-for-byte
	if string(exportedCanonical) != string(sourceCanonical) {
		return fmt.Errorf("schema payload differs from source (after stripping provenance)")
	}

	return nil
}
