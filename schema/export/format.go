package export

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// formatJSON formats schema data as JSON with optional provenance
func formatJSON(schemaData []byte, metadata *ProvenanceMetadata, style ProvenanceStyle) ([]byte, error) {
	// Parse the schema JSON
	var schemaObj map[string]interface{}
	if err := json.Unmarshal(schemaData, &schemaObj); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// Add provenance if metadata is provided
	if metadata != nil {
		switch style {
		case ProvenanceObject:
			// Add x-crucible-source as a top-level field
			schemaObj["x-crucible-source"] = metadata

		case ProvenanceComment:
			// Add provenance as $comment field
			commentStr := formatProvenanceComment(metadata)
			schemaObj["$comment"] = commentStr

		case ProvenanceNone:
			// No provenance added
		}
	}

	// Marshal with 2-space indentation for readability
	output, err := json.MarshalIndent(schemaObj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Add trailing newline for POSIX compliance
	output = append(output, '\n')

	return output, nil
}

// formatYAML formats schema data as YAML with optional provenance
func formatYAML(schemaData []byte, metadata *ProvenanceMetadata, style ProvenanceStyle) ([]byte, error) {
	// Parse the schema JSON into a generic structure
	var schemaObj interface{}
	if err := json.Unmarshal(schemaData, &schemaObj); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	var buf bytes.Buffer

	// Add provenance as YAML comment front-matter if metadata is provided
	if metadata != nil && style != ProvenanceNone {
		switch style {
		case ProvenanceObject:
			// Write provenance as YAML comment
			buf.WriteString("# x-crucible-source:\n")
			buf.WriteString(fmt.Sprintf("#   schema_id: %s\n", metadata.SchemaID))
			buf.WriteString(fmt.Sprintf("#   crucible_version: %s\n", metadata.CrucibleVersion))
			buf.WriteString(fmt.Sprintf("#   gofulmen_version: %s\n", metadata.GofulmenVersion))
			if metadata.GitRevision != "" {
				buf.WriteString(fmt.Sprintf("#   git_revision: %s\n", metadata.GitRevision))
			}
			buf.WriteString(fmt.Sprintf("#   exported_at: %s\n", metadata.ExportedAt.Format("2006-01-02T15:04:05Z07:00")))
			if metadata.Identity != nil {
				if metadata.Identity.Vendor != "" || metadata.Identity.Binary != "" {
					buf.WriteString("#   identity:\n")
					if metadata.Identity.Vendor != "" {
						buf.WriteString(fmt.Sprintf("#     vendor: %s\n", metadata.Identity.Vendor))
					}
					if metadata.Identity.Binary != "" {
						buf.WriteString(fmt.Sprintf("#     binary: %s\n", metadata.Identity.Binary))
					}
				}
			}
			buf.WriteString("---\n")
		case ProvenanceComment:
			// Write compact provenance comment
			buf.WriteString("# ")
			buf.WriteString(formatProvenanceComment(metadata))
			buf.WriteString("\n---\n")
		}
	}

	// Marshal to YAML
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(schemaObj); err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close YAML encoder: %w", err)
	}

	return buf.Bytes(), nil
}
