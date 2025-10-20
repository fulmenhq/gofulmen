package schema

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCatalogListAndGet(t *testing.T) {
	catalog := DefaultCatalog()

	schemas, err := catalog.ListSchemas("pathfinder/")
	if err != nil {
		t.Fatalf("ListSchemas failed: %v", err)
	}
	if len(schemas) == 0 {
		t.Fatalf("expected schemas for prefix, got none")
	}

	found := false
	for _, desc := range schemas {
		if desc.ID == "pathfinder/v1.0.0/path-result" {
			found = true
			if desc.Path == "" {
				t.Fatalf("descriptor path empty for %s", desc.ID)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected pathfinder/v1.0.0/path-result in catalog")
	}

	desc, err := catalog.GetSchema("pathfinder/v1.0.0/path-result")
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}
	if desc.Title == "" {
		t.Fatalf("expected title for descriptor")
	}
}

func TestCatalogValidateDataByID(t *testing.T) {
	catalog := DefaultCatalog()
	id := "pathfinder/v1.0.0/path-result"

	validPayload := map[string]any{
		"relativePath": "assets/config.yaml",
		"sourcePath":   "/tmp/config.yaml",
		"logicalPath":  "assets/config.yaml",
		"loaderType":   "local",
		"metadata": map[string]any{
			"size": 10,
		},
	}

	payloadBytes, err := json.Marshal(validPayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	diags, err := catalog.ValidateDataByID(id, payloadBytes)
	if err != nil {
		t.Fatalf("ValidateDataByID returned error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid payload, got %v", diags)
	}

	invalidPayload := map[string]any{
		"sourcePath": "/tmp/config.yaml",
		"loaderType": "local",
	}
	invalidBytes, _ := json.Marshal(invalidPayload)

	diags, err = catalog.ValidateDataByID(id, invalidBytes)
	if err != nil {
		t.Fatalf("ValidateDataByID returned error for invalid payload: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for invalid payload, got none")
	}
	foundRequired := false
	for _, d := range diags {
		if d.Keyword == "required" || strings.Contains(d.Message, "missing properties") {
			foundRequired = true
			break
		}
	}
	if !foundRequired {
		t.Fatalf("missing required-field diagnostic: %v", diags)
	}
}

func TestCompareSchema(t *testing.T) {
	catalog := DefaultCatalog()
	id := "pathfinder/v1.0.0/path-result"

	desc, err := catalog.GetSchema(id)
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	original, err := os.ReadFile(desc.Path)
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	diff, err := catalog.CompareSchema(id, original)
	if err != nil {
		t.Fatalf("CompareSchema failed: %v", err)
	}
	if len(diff) != 0 {
		t.Fatalf("expected no diff for identical schema, got %v", diff)
	}

	var mutated map[string]any
	if err := json.Unmarshal(original, &mutated); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}
	mutated["title"] = "mutated title"
	mutatedBytes, _ := json.Marshal(mutated)

	diff, err = catalog.CompareSchema(id, mutatedBytes)
	if err != nil {
		t.Fatalf("CompareSchema failed: %v", err)
	}
	if len(diff) == 0 {
		t.Fatalf("expected diff when schema mutated")
	}
}
