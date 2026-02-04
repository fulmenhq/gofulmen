package crucible

import (
	"fmt"
	"strings"

	"github.com/fulmenhq/crucible"
	internalversion "github.com/fulmenhq/gofulmen/internal/version"
)

// GofulmenVersion is deprecated. Use foundry.GofulmenVersion() instead.
// This constant is kept for backward compatibility but will be removed in v0.2.0.
const GofulmenVersion = "0.3.3"

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
	gofulmenVersion := strings.TrimPrefix(internalversion.ModuleVersion("github.com/fulmenhq/gofulmen"), "v")
	if gofulmenVersion == "" {
		gofulmenVersion = GofulmenVersion
	}
	return Version{
		Gofulmen: gofulmenVersion,
		Crucible: CrucibleVersion,
	}
}

func GetVersionString() string {
	v := GetVersion()
	return fmt.Sprintf("gofulmen/%s crucible/%s", v.Gofulmen, v.Crucible)
}
