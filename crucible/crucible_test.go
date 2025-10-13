package crucible

import (
	"testing"
)

func TestGetVersion(t *testing.T) {
	v := GetVersion()
	if v.Gofulmen == "" {
		t.Error("Gofulmen version should not be empty")
	}
	if v.Crucible == "" {
		t.Error("Crucible version should not be empty")
	}
}

func TestGetVersionString(t *testing.T) {
	vs := GetVersionString()
	if vs == "" {
		t.Error("Version string should not be empty")
	}
	t.Logf("Version: %s", vs)
}

func TestSchemaRegistry(t *testing.T) {
	if SchemaRegistry == nil {
		t.Fatal("SchemaRegistry should not be nil")
	}

	if SchemaRegistry.Terminal() == nil {
		t.Error("Terminal schemas should be accessible")
	}
	if SchemaRegistry.Pathfinder() == nil {
		t.Error("Pathfinder schemas should be accessible")
	}
	if SchemaRegistry.ASCII() == nil {
		t.Error("ASCII schemas should be accessible")
	}
	if SchemaRegistry.SchemaValidation() == nil {
		t.Error("SchemaValidation schemas should be accessible")
	}
	if SchemaRegistry.Observability() == nil {
		t.Error("Observability schemas should be accessible")
	}
}

func TestStandardsRegistry(t *testing.T) {
	if StandardsRegistry == nil {
		t.Fatal("StandardsRegistry should not be nil")
	}

	if StandardsRegistry.Coding() == nil {
		t.Error("Coding standards should be accessible")
	}
}

func TestLoadLoggingSchemas(t *testing.T) {
	schemas, err := LoadLoggingSchemas()
	if err != nil {
		t.Fatalf("Failed to load logging schemas: %v", err)
	}
	if schemas == nil {
		t.Error("Logging schemas should not be nil")
	}

	eventSchema, err := schemas.LogEvent()
	if err != nil {
		t.Errorf("Failed to get log event schema: %v", err)
	}
	if len(eventSchema) == 0 {
		t.Error("Log event schema should not be empty")
	}

	configSchema, err := schemas.LoggerConfig()
	if err != nil {
		t.Errorf("Failed to get logger config schema: %v", err)
	}
	if len(configSchema) == 0 {
		t.Error("Logger config schema should not be empty")
	}
}

func TestLoadPathfinderSchemas(t *testing.T) {
	schemas, err := LoadPathfinderSchemas()
	if err != nil {
		t.Fatalf("Failed to load pathfinder schemas: %v", err)
	}
	if schemas == nil {
		t.Error("Pathfinder schemas should not be nil")
	}

	findQuerySchema, err := schemas.FindQuery()
	if err != nil {
		t.Errorf("Failed to get find query schema: %v", err)
	}
	if len(findQuerySchema) == 0 {
		t.Error("Find query schema should not be empty")
	}
}

func TestLoadSchemaValidationSchemas(t *testing.T) {
	schemas, err := LoadSchemaValidationSchemas()
	if err != nil {
		t.Fatalf("Failed to load schema validation schemas: %v", err)
	}
	if schemas == nil {
		t.Error("Schema validation schemas should not be nil")
	}

	validatorConfig, err := schemas.ValidatorConfig()
	if err != nil {
		t.Errorf("Failed to get validator config schema: %v", err)
	}
	if len(validatorConfig) == 0 {
		t.Error("Validator config schema should not be empty")
	}
}

func TestGetLoggingEventSchema(t *testing.T) {
	schema, err := GetLoggingEventSchema()
	if err != nil {
		t.Fatalf("Failed to get logging event schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Logging event schema should not be empty")
	}
}

func TestGetLoggingConfigSchema(t *testing.T) {
	schema, err := GetLoggingConfigSchema()
	if err != nil {
		t.Fatalf("Failed to get logging config schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Logging config schema should not be empty")
	}
}

func TestGetPathfinderFindQuerySchema(t *testing.T) {
	schema, err := GetPathfinderFindQuerySchema()
	if err != nil {
		t.Fatalf("Failed to get pathfinder find query schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Pathfinder find query schema should not be empty")
	}
}

func TestGetPathfinderConfigSchema(t *testing.T) {
	schema, err := GetPathfinderConfigSchema()
	if err != nil {
		t.Fatalf("Failed to get pathfinder config schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Pathfinder config schema should not be empty")
	}
}

func TestGetGoStandards(t *testing.T) {
	standards, err := GetGoStandards()
	if err != nil {
		t.Fatalf("Failed to get Go standards: %v", err)
	}
	if standards == "" {
		t.Error("Go standards should not be empty")
	}
}

func TestGetTypeScriptStandards(t *testing.T) {
	standards, err := GetTypeScriptStandards()
	if err != nil {
		t.Fatalf("Failed to get TypeScript standards: %v", err)
	}
	if standards == "" {
		t.Error("TypeScript standards should not be empty")
	}
}

func TestGetTerminalSchema(t *testing.T) {
	schema, err := GetTerminalSchema()
	if err != nil {
		t.Fatalf("Failed to get terminal schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Terminal schema should not be empty")
	}
}

func TestGetTerminalCatalog(t *testing.T) {
	catalog, err := GetTerminalCatalog()
	if err != nil {
		t.Fatalf("Failed to get terminal catalog: %v", err)
	}
	if len(catalog) == 0 {
		t.Error("Terminal catalog should not be empty")
	}
}

func TestGetASCIIStringAnalysisSchema(t *testing.T) {
	schema, err := GetASCIIStringAnalysisSchema()
	if err != nil {
		t.Fatalf("Failed to get ASCII string analysis schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("ASCII string analysis schema should not be empty")
	}
}

func TestGetASCIIBoxCharsSchema(t *testing.T) {
	schema, err := GetASCIIBoxCharsSchema()
	if err != nil {
		t.Fatalf("Failed to get ASCII box chars schema: %v", err)
	}
	if len(schema) == 0 {
		t.Error("ASCII box chars schema should not be empty")
	}
}

func TestGetSchema(t *testing.T) {
	schema, err := GetSchema("terminal/v1.0.0/schema.json")
	if err != nil {
		t.Fatalf("Failed to get schema by path: %v", err)
	}
	if len(schema) == 0 {
		t.Error("Schema should not be empty")
	}
}

func TestListSchemas(t *testing.T) {
	schemas, err := ListSchemas("observability/logging/v1.0.0")
	if err != nil {
		t.Fatalf("Failed to list schemas: %v", err)
	}
	if len(schemas) == 0 {
		t.Error("Should have at least one schema")
	}
	t.Logf("Found %d schemas", len(schemas))
}

func TestParseJSONSchema(t *testing.T) {
	schemaData, err := GetTerminalSchema()
	if err != nil {
		t.Fatalf("Failed to get terminal schema: %v", err)
	}

	parsed, err := ParseJSONSchema(schemaData)
	if err != nil {
		t.Fatalf("Failed to parse JSON schema: %v", err)
	}
	if len(parsed) == 0 {
		t.Error("Parsed schema should not be empty")
	}
}

func TestValidateAgainstSchema(t *testing.T) {
	t.Skip("ValidateAgainstSchema requires full schema resolution setup - tested via logging package")
}

func TestGetArchitectureDocs(t *testing.T) {
	doc, err := GetDoc("architecture/pseudo-monorepo.md")
	if err != nil {
		t.Fatalf("Failed to get pseudo-monorepo doc: %v", err)
	}
	if len(doc) == 0 {
		t.Error("Pseudo-monorepo doc should not be empty")
	}
	t.Logf("Pseudo-monorepo doc: %d bytes", len(doc))

	syncDoc, err := GetDoc("architecture/sync-model.md")
	if err != nil {
		t.Fatalf("Failed to get sync-model doc: %v", err)
	}
	if len(syncDoc) == 0 {
		t.Error("Sync-model doc should not be empty")
	}
	t.Logf("Sync-model doc: %d bytes", len(syncDoc))

	ecosystemDoc, err := GetDoc("architecture/library-ecosystem.md")
	if err != nil {
		t.Fatalf("Failed to get library-ecosystem doc: %v", err)
	}
	if len(ecosystemDoc) == 0 {
		t.Error("Library-ecosystem doc should not be empty")
	}
	t.Logf("Library-ecosystem doc: %d bytes", len(ecosystemDoc))
}

func TestArchitectureDocHelpers(t *testing.T) {
	ecosystemDoc, err := GetLibraryEcosystemDoc()
	if err != nil {
		t.Fatalf("Failed to get library ecosystem doc: %v", err)
	}
	if len(ecosystemDoc) == 0 {
		t.Error("Library ecosystem doc should not be empty")
	}
	t.Logf("GetLibraryEcosystemDoc(): %d bytes", len(ecosystemDoc))

	pseudoDoc, err := GetPseudoMonorepoDoc()
	if err != nil {
		t.Fatalf("Failed to get pseudo-monorepo doc: %v", err)
	}
	if len(pseudoDoc) == 0 {
		t.Error("Pseudo-monorepo doc should not be empty")
	}
	t.Logf("GetPseudoMonorepoDoc(): %d bytes", len(pseudoDoc))

	syncDoc, err := GetSyncModelDoc()
	if err != nil {
		t.Fatalf("Failed to get sync-model doc: %v", err)
	}
	if len(syncDoc) == 0 {
		t.Error("Sync-model doc should not be empty")
	}
	t.Logf("GetSyncModelDoc(): %d bytes", len(syncDoc))

	guideDoc, err := GetIntegrationGuideDoc()
	if err != nil {
		t.Fatalf("Failed to get integration guide doc: %v", err)
	}
	if len(guideDoc) == 0 {
		t.Error("Integration guide doc should not be empty")
	}
	t.Logf("GetIntegrationGuideDoc(): %d bytes", len(guideDoc))
}
