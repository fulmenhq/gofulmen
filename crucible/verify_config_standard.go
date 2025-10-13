package crucible

// GetConfigPathStandard returns the Fulmen config path standard document
func GetConfigPathStandard() (string, error) {
	return GetDoc("standards/config/fulmen-config-paths.md")
}

// GetConfigPathMigrationSOP returns the config path migration SOP
func GetConfigPathMigrationSOP() (string, error) {
	return GetDoc("sop/config-path-migration.md")
}

// GetConfigPathSchema returns the Fulmen config paths JSON schema
func GetConfigPathSchema() ([]byte, error) {
	return GetSchema("config/fulmen-ecosystem/v1.0.0/fulmen-config-paths.schema.json")
}
