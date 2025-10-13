package crucible

import (
	"testing"
)

func TestGetConfigPathStandard(t *testing.T) {
	doc, err := GetConfigPathStandard()
	if err != nil {
		t.Fatalf("Failed to get config path standard: %v", err)
	}
	if len(doc) == 0 {
		t.Error("Config path standard should not be empty")
	}
	t.Logf("Config path standard: %d bytes", len(doc))
}

func TestGetConfigPathMigrationSOP(t *testing.T) {
	doc, err := GetConfigPathMigrationSOP()
	if err != nil {
		t.Fatalf("Failed to get config path migration SOP: %v", err)
	}
	if len(doc) == 0 {
		t.Error("Config path migration SOP should not be empty")
	}
	t.Logf("Config path migration SOP: %d bytes", len(doc))
}

func TestGetConfigPathSchema(t *testing.T) {
	schema, err := GetConfigPathSchema()
	if err != nil {
		t.Fatalf("Failed to get config path schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Config path schema should not be empty")
	}
	t.Logf("Config path schema: %d bytes", len(schema))
}
