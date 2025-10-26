package metrics

const (
	SchemaValidations          = "schema_validations"
	SchemaValidationErrors     = "schema_validation_errors"
	ConfigLoadMs               = "config_load_ms"
	ConfigLoadErrors           = "config_load_errors"
	PathfinderFindMs           = "pathfinder_find_ms"
	PathfinderValidationErrors = "pathfinder_validation_errors"
	PathfinderSecurityWarnings = "pathfinder_security_warnings"
	FoundryLookupCount         = "foundry_lookup_count"
	LoggingEmitCount           = "logging_emit_count"
	LoggingEmitLatencyMs       = "logging_emit_latency_ms"
	GoneatCommandDurationMs    = "goneat_command_duration_ms"
	FulHashHashCount           = "fulhash_hash_count"
	FulHashErrorsCount         = "fulhash_errors_count"
)

const (
	UnitCount   = "count"
	UnitMs      = "ms"
	UnitBytes   = "bytes"
	UnitPercent = "percent"
)

const (
	TagStatus    = "status"
	TagComponent = "component"
	TagOperation = "operation"
	TagCategory  = "category"
	TagVersion   = "version"
	TagSeverity  = "severity"
	TagLayer     = "layer"
	TagRoot      = "root"
	TagEndpoint  = "endpoint"
	TagHost      = "host"
	TagAlgorithm = "algorithm"
	TagErrorType = "error_type"
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)
