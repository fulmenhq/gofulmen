package schema

import (
	"encoding/json"
	"testing"
)

func TestMergeJSONSchemas(t *testing.T) {
	base := []byte(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	overlay := []byte(`{"properties":{"age":{"type":"integer"}}}`)

	mergedBytes, err := MergeJSONSchemas(base, overlay)
	if err != nil {
		t.Fatalf("MergeJSONSchemas returned error: %v", err)
	}

	var merged map[string]any
	if err := json.Unmarshal(mergedBytes, &merged); err != nil {
		t.Fatalf("failed to unmarshal merged schema: %v", err)
	}

	props := merged["properties"].(map[string]any)
	if _, ok := props["name"]; !ok {
		t.Fatalf("expected name property retained")
	}
	if _, ok := props["age"]; !ok {
		t.Fatalf("expected age property merged")
	}
}

func TestDiffSchemas(t *testing.T) {
	left := []byte(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	right := []byte(`{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"}}}`)

	diffs, err := DiffSchemas(left, right)
	if err != nil {
		t.Fatalf("DiffSchemas returned error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatalf("expected diffs when schema changes")
	}
}
