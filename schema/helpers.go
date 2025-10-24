package schema

// ValidateSchemaByID validates the schema definition identified by ID using the default catalog.
func ValidateSchemaByID(id string) ([]Diagnostic, error) {
	catalog := globalCatalog()
	return catalog.ValidateSchemaByID(id)
}

// ValidateDataByID validates JSON bytes against the schema identified by ID using the default catalog.
func ValidateDataByID(id string, data []byte) ([]Diagnostic, error) {
	catalog := globalCatalog()
	return catalog.ValidateDataByID(id, data)
}

// ValidateFileByID validates a file (JSON or YAML) against the schema identified by ID.
func ValidateFileByID(id, path string) ([]Diagnostic, error) {
	catalog := globalCatalog()
	return catalog.ValidateFileByID(id, path)
}

// CatalogForRoot returns a catalog rooted at the provided directory. Useful for tests.
func CatalogForRoot(root string) *Catalog {
	return NewCatalog(root)
}

// DiagnosticsToStringSlice converts a slice of diagnostics to a slice of strings for error context.
func DiagnosticsToStringSlice(diags []Diagnostic) []string {
	if len(diags) == 0 {
		return nil
	}

	result := make([]string, len(diags))
	for i, diag := range diags {
		result[i] = diag.Message
	}
	return result
}
