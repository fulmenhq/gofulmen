package crucible

import (
	"fmt"

	"github.com/fulmenhq/gofulmen/schema"
)

func LoadLoggingSchemas() (*LoggingSchemasV1, error) {
	logging, err := SchemaRegistry.Observability().Logging().V1_0_0()
	if err != nil {
		return nil, fmt.Errorf("failed to load logging schemas: %w", err)
	}
	return logging, nil
}

func LoadPathfinderSchemas() (*PathfinderSchemasV1, error) {
	pathfinder, err := SchemaRegistry.Pathfinder().V1_0_0()
	if err != nil {
		return nil, fmt.Errorf("failed to load pathfinder schemas: %w", err)
	}
	return pathfinder, nil
}

func LoadSchemaValidationSchemas() (*SchemaValidationSchemasV1, error) {
	schemaVal, err := SchemaRegistry.SchemaValidation().V1_0_0()
	if err != nil {
		return nil, fmt.Errorf("failed to load schema validation schemas: %w", err)
	}
	return schemaVal, nil
}

func ValidateAgainstSchema(schemaData []byte, jsonData []byte) error {
	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	diags, err := validator.ValidateJSON(jsonData)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if verrs := schema.DiagnosticsToValidationErrors(diags); len(verrs) > 0 {
		return verrs
	}
	return nil
}

func GetLoggingEventSchema() ([]byte, error) {
	logging, err := LoadLoggingSchemas()
	if err != nil {
		return nil, err
	}
	return logging.LogEvent()
}

func GetLoggingConfigSchema() ([]byte, error) {
	logging, err := LoadLoggingSchemas()
	if err != nil {
		return nil, err
	}
	return logging.LoggerConfig()
}

func GetPathfinderFindQuerySchema() ([]byte, error) {
	pathfinder, err := LoadPathfinderSchemas()
	if err != nil {
		return nil, err
	}
	return pathfinder.FindQuery()
}

func GetPathfinderConfigSchema() ([]byte, error) {
	pathfinder, err := LoadPathfinderSchemas()
	if err != nil {
		return nil, err
	}
	return pathfinder.FinderConfig()
}

func GetGoStandards() (string, error) {
	return StandardsRegistry.Coding().Go()
}

func GetTypeScriptStandards() (string, error) {
	return StandardsRegistry.Coding().TypeScript()
}

func GetTerminalSchema() ([]byte, error) {
	return SchemaRegistry.Terminal().V1_0_0()
}

func GetTerminalCatalog() (map[string][]byte, error) {
	return SchemaRegistry.Terminal().Catalog()
}

func GetASCIIStringAnalysisSchema() ([]byte, error) {
	ascii, err := SchemaRegistry.ASCII().V1_0_0()
	if err != nil {
		return nil, err
	}
	return ascii.StringAnalysis()
}

func GetASCIIBoxCharsSchema() ([]byte, error) {
	ascii, err := SchemaRegistry.ASCII().V1_0_0()
	if err != nil {
		return nil, err
	}
	return ascii.BoxChars()
}

func GetLibraryEcosystemDoc() (string, error) {
	return GetDoc("architecture/library-ecosystem.md")
}

func GetPseudoMonorepoDoc() (string, error) {
	return GetDoc("architecture/pseudo-monorepo.md")
}

func GetSyncModelDoc() (string, error) {
	return GetDoc("architecture/sync-model.md")
}

func GetIntegrationGuideDoc() (string, error) {
	return GetDoc("guides/integration-guide.md")
}

// ListAvailableDocs returns a list of all available documentation paths
// organized by category for easy discovery
func ListAvailableDocs() map[string][]string {
	return map[string][]string{
		"architecture": {
			"architecture/library-ecosystem.md",
			"architecture/pseudo-monorepo.md",
			"architecture/sync-model.md",
		},
		"standards": {
			"standards/config/fulmen-config-paths.md",
			"standards/config/README.md",
			"standards/coding/go.md",
			"standards/coding/typescript.md",
			"standards/frontmatter-standard.md",
			"standards/agentic-attribution.md",
		},
		"guides": {
			"guides/integration-guide.md",
			"guides/sync-strategy.md",
		},
		"sop": {
			"sop/config-path-migration.md",
			"sop/repository-structure.md",
			"sop/repository-version-adoption.md",
			"sop/cicd-operations.md",
			"sop/README.md",
		},
	}
}

// ListAvailableSchemas returns a list of common schema paths
// for easy discovery
func ListAvailableSchemas() map[string][]string {
	return map[string][]string{
		"observability/logging": {
			"observability/logging/v1.0.0/definitions.schema.json",
			"observability/logging/v1.0.0/log-event.schema.json",
			"observability/logging/v1.0.0/logger-config.schema.json",
			"observability/logging/v1.0.0/middleware-config.schema.json",
		},
		"pathfinder": {
			"pathfinder/v1.0.0/find-query.schema.json",
			"pathfinder/v1.0.0/finder-config.schema.json",
			"pathfinder/v1.0.0/path-result.schema.json",
			"pathfinder/v1.0.0/error-response.schema.json",
			"pathfinder/v1.0.0/metadata.schema.json",
		},
		"config": {
			"config/fulmen-ecosystem/v1.0.0/fulmen-config-paths.schema.json",
			"config/fulmen-ecosystem/v1.0.0/README.md",
		},
		"terminal": {
			"terminal/v1.0.0/schema.json",
		},
		"ascii": {
			"ascii/v1.0.0/string-analysis.schema.json",
			"ascii/v1.0.0/box-chars.schema.json",
		},
		"schema-validation": {
			"schema-validation/v1.0.0/validator-config.schema.json",
			"schema-validation/v1.0.0/schema-registry.schema.json",
		},
	}
}
