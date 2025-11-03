package crucible

import (
	"fmt"

	"github.com/fulmenhq/crucible"
)

// GofulmenVersion is deprecated. Use foundry.GofulmenVersion() instead.
// This constant is kept for backward compatibility but will be removed in v0.2.0.
const GofulmenVersion = "0.1.8"

const (
	CrucibleVersion = crucible.Version
)

var (
	SchemaRegistry    = crucible.SchemaRegistry
	StandardsRegistry = crucible.StandardsRegistry
	ConfigRegistry    = crucible.ConfigRegistry
)

type Schemas = crucible.Schemas
type Standards = crucible.Standards
type Config = crucible.Config

type TerminalSchemas = crucible.TerminalSchemas
type PathfinderSchemas = crucible.PathfinderSchemas
type PathfinderSchemasV1 = crucible.PathfinderSchemasV1
type ASCIISchemas = crucible.ASCIISchemas
type ASCIISchemasV1 = crucible.ASCIISchemasV1
type SchemaValidationSchemas = crucible.SchemaValidationSchemas
type SchemaValidationSchemasV1 = crucible.SchemaValidationSchemasV1
type ObservabilitySchemas = crucible.ObservabilitySchemas
type LoggingSchemas = crucible.LoggingSchemas
type LoggingSchemasV1 = crucible.LoggingSchemasV1
type CodingStandards = crucible.CodingStandards

func GetSchema(schemaPath string) ([]byte, error) {
	return crucible.GetSchema(schemaPath)
}

func GetDoc(docPath string) (string, error) {
	return crucible.GetDoc(docPath)
}

func ListSchemas(basePath string) ([]string, error) {
	return crucible.ListSchemas(basePath)
}

func ParseJSONSchema(data []byte) (map[string]any, error) {
	return crucible.ParseJSONSchema(data)
}

func GetConfig(configPath string) ([]byte, error) {
	return crucible.GetConfig(configPath)
}

func ListConfigs(basePath string) ([]string, error) {
	return crucible.ListConfigs(basePath)
}

type Version struct {
	Gofulmen string `json:"gofulmen"`
	Crucible string `json:"crucible"`
}

func GetVersion() Version {
	return Version{
		Gofulmen: GofulmenVersion,
		Crucible: CrucibleVersion,
	}
}

func GetVersionString() string {
	return fmt.Sprintf("gofulmen/%s crucible/%s", GofulmenVersion, CrucibleVersion)
}
