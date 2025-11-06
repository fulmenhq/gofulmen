package metrics

// Core metrics from Crucible taxonomy
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

// Prometheus Exporter Metrics (Crucible v0.2.7 taxonomy)
const (
	PrometheusExporterRefreshDurationSeconds = "prometheus_exporter_refresh_duration_seconds"
	PrometheusExporterRefreshTotal           = "prometheus_exporter_refresh_total"
	PrometheusExporterRefreshErrorsTotal     = "prometheus_exporter_refresh_errors_total"
	PrometheusExporterRefreshInflight        = "prometheus_exporter_refresh_inflight"
	PrometheusExporterHTTPRequestsTotal      = "prometheus_exporter_http_requests_total"
	PrometheusExporterHTTPErrorsTotal        = "prometheus_exporter_http_errors_total"
	PrometheusExporterRestartsTotal          = "prometheus_exporter_restarts_total"
)

// Foundry Module Metrics (MIME detection)
const (
	FoundryMimeDetectionsTotalJSON      = "foundry_mime_detections_total_json"
	FoundryMimeDetectionsTotalXML       = "foundry_mime_detections_total_xml"
	FoundryMimeDetectionsTotalYAML      = "foundry_mime_detections_total_yaml"
	FoundryMimeDetectionsTotalCSV       = "foundry_mime_detections_total_csv"
	FoundryMimeDetectionsTotalPlainText = "foundry_mime_detections_total_plain_text"
	FoundryMimeDetectionsTotalUnknown   = "foundry_mime_detections_total_unknown"
	FoundryMimeDetectionMs              = "foundry_mime_detection_ms"
)

// Error Handling Module Metrics
const (
	ErrorHandlingWrapsTotal = "error_handling_wraps_total"
	ErrorHandlingWrapMs     = "error_handling_wrap_ms"
)

// FulHash Module Metrics
const (
	FulHashOperationsTotalXXH3128 = "fulhash_operations_total_xxh3_128"
	FulHashOperationsTotalSHA256  = "fulhash_operations_total_sha256"
	FulHashHashStringTotal        = "fulhash_hash_string_total"
	FulHashBytesHashedTotal       = "fulhash_bytes_hashed_total"
	FulHashOperationMs            = "fulhash_operation_ms"
)

// Metric units
const (
	UnitCount   = "count"
	UnitMs      = "ms"
	UnitSeconds = "seconds"
	UnitBytes   = "bytes"
	UnitPercent = "percent"
)

// Standard tag keys
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
	TagPhase     = "phase"
	TagResult    = "result"
	TagReason    = "reason"
	TagPath      = "path"
	TagClient    = "client"
	TagMimeType  = "mime_type"
)

// Standard tag values
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusError   = "error"
)

// Prometheus exporter phase values
const (
	PhaseCollect = "collect"
	PhaseConvert = "convert"
	PhaseExport  = "export"
)

// Prometheus exporter result values
const (
	ResultSuccess = "success"
	ResultError   = "error"
)

// Prometheus exporter error types
const (
	ErrorTypeValidation = "validation"
	ErrorTypeIO         = "io"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeOther      = "other"
)

// Prometheus exporter restart reasons
const (
	RestartReasonConfig       = "config"
	RestartReasonPanicRecover = "panic_recover"
	RestartReasonManual       = "manual"
	RestartReasonDependency   = "dependency"
)
