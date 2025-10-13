package crucible

import (
	"testing"
)

func TestListAvailableDocs(t *testing.T) {
	catalog := ListAvailableDocs()

	if len(catalog) == 0 {
		t.Error("Doc catalog should not be empty")
	}

	// Check for expected categories
	expectedCategories := []string{"architecture", "standards", "guides", "sop"}
	for _, cat := range expectedCategories {
		if docs, ok := catalog[cat]; !ok || len(docs) == 0 {
			t.Errorf("Category %s should exist with docs", cat)
		}
	}

	// Verify specific important docs
	if docs, ok := catalog["standards"]; ok {
		found := false
		for _, doc := range docs {
			if doc == "standards/config/fulmen-config-paths.md" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should include config path standard")
		}
	}

	t.Logf("Available doc categories: %d", len(catalog))
	for cat, docs := range catalog {
		t.Logf("  %s: %d docs", cat, len(docs))
	}
}

func TestListAvailableSchemas(t *testing.T) {
	catalog := ListAvailableSchemas()

	if len(catalog) == 0 {
		t.Error("Schema catalog should not be empty")
	}

	// Check for expected categories
	expectedCategories := []string{"observability/logging", "pathfinder", "config", "terminal", "ascii"}
	for _, cat := range expectedCategories {
		if schemas, ok := catalog[cat]; !ok || len(schemas) == 0 {
			t.Errorf("Category %s should exist with schemas", cat)
		}
	}

	// Verify specific important schemas
	if schemas, ok := catalog["observability/logging"]; ok {
		found := false
		for _, schema := range schemas {
			if schema == "observability/logging/v1.0.0/log-event.schema.json" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Should include log event schema")
		}
	}

	t.Logf("Available schema categories: %d", len(catalog))
	for cat, schemas := range catalog {
		t.Logf("  %s: %d schemas", cat, len(schemas))
	}
}

func TestDocCatalogAccessibility(t *testing.T) {
	_ = ListAvailableDocs()

	// Try to access a few docs from the catalog
	testDocs := []string{
		"architecture/library-ecosystem.md",
		"standards/config/fulmen-config-paths.md",
		"sop/config-path-migration.md",
	}

	for _, docPath := range testDocs {
		doc, err := GetDoc(docPath)
		if err != nil {
			t.Errorf("Doc listed in catalog should be accessible: %s: %v", docPath, err)
		}
		if len(doc) == 0 {
			t.Errorf("Doc should not be empty: %s", docPath)
		}
	}
}

func TestSchemaCatalogAccessibility(t *testing.T) {
	_ = ListAvailableSchemas()

	// Try to access a few schemas from the catalog
	testSchemas := []string{
		"observability/logging/v1.0.0/log-event.schema.json",
		"pathfinder/v1.0.0/find-query.schema.json",
		"config/fulmen-ecosystem/v1.0.0/fulmen-config-paths.schema.json",
	}

	for _, schemaPath := range testSchemas {
		schema, err := GetSchema(schemaPath)
		if err != nil {
			t.Errorf("Schema listed in catalog should be accessible: %s: %v", schemaPath, err)
		}
		if len(schema) == 0 {
			t.Errorf("Schema should not be empty: %s", schemaPath)
		}
	}
}
