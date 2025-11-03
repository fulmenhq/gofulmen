// Package foundry provides standardized exit codes and metadata for Fulmen ecosystem applications.
//
// This file re-exports exit code constants from github.com/fulmenhq/crucible/foundry,
// which are generated from the canonical Foundry exit codes catalog.
//
// Re-exporting through gofulmen/foundry provides:
// - API stability if Crucible package paths change
// - Additional metadata and platform-specific helpers
// - Simplified mode mapping and BSD compatibility
//
// For the authoritative catalog and generation source, see:
// - Catalog: config/library/foundry/exit-codes.yaml in Crucible
// - Generator: scripts/codegen/generate-exit-codes.ts in Crucible
package foundry

import crucible "github.com/fulmenhq/crucible/foundry"

// ExitCode represents a standardized exit code from the Foundry catalog.
// It is an alias for int to maintain compatibility with os.Exit() and standard library.
type ExitCode = int

// Re-export all exit code constants from Crucible's generated bindings.
// These constants are generated from config/library/foundry/exit-codes.yaml.
const (
	// Standard Exit Codes (0-1)
	// POSIX standard success and generic failure codes
	ExitSuccess = crucible.ExitSuccess
	ExitFailure = crucible.ExitFailure

	// Networking & Port Management (10-19)
	// Network-related failures (ports, connectivity, etc.)
	ExitPortInUse              = crucible.ExitPortInUse
	ExitPortRangeExhausted     = crucible.ExitPortRangeExhausted
	ExitInstanceAlreadyRunning = crucible.ExitInstanceAlreadyRunning
	ExitNetworkUnreachable     = crucible.ExitNetworkUnreachable
	ExitConnectionRefused      = crucible.ExitConnectionRefused
	ExitConnectionTimeout      = crucible.ExitConnectionTimeout

	// Configuration & Validation (20-29)
	// Configuration errors, validation failures, version mismatches
	ExitConfigInvalid       = crucible.ExitConfigInvalid
	ExitMissingDependency   = crucible.ExitMissingDependency
	ExitSsotVersionMismatch = crucible.ExitSsotVersionMismatch
	ExitConfigFileNotFound  = crucible.ExitConfigFileNotFound
	ExitEnvironmentInvalid  = crucible.ExitEnvironmentInvalid

	// Runtime Errors (30-39)
	// Errors during normal operation (health checks, database, etc.)
	ExitHealthCheckFailed          = crucible.ExitHealthCheckFailed
	ExitDatabaseUnavailable        = crucible.ExitDatabaseUnavailable
	ExitExternalServiceUnavailable = crucible.ExitExternalServiceUnavailable
	ExitResourceExhausted          = crucible.ExitResourceExhausted
	ExitOperationTimeout           = crucible.ExitOperationTimeout

	// Command-Line Usage Errors (40-49)
	// Invalid arguments, missing required flags, usage errors
	ExitInvalidArgument         = crucible.ExitInvalidArgument
	ExitMissingRequiredArgument = crucible.ExitMissingRequiredArgument
	ExitUsage                   = crucible.ExitUsage

	// Permissions & File Access (50-59)
	// Permission denied, file not found, access errors
	ExitPermissionDenied  = crucible.ExitPermissionDenied
	ExitFileNotFound      = crucible.ExitFileNotFound
	ExitDirectoryNotFound = crucible.ExitDirectoryNotFound
	ExitFileReadError     = crucible.ExitFileReadError
	ExitFileWriteError    = crucible.ExitFileWriteError

	// Data & Processing Errors (60-69)
	// Data validation, parsing, transformation failures
	ExitDataInvalid          = crucible.ExitDataInvalid
	ExitParseError           = crucible.ExitParseError
	ExitTransformationFailed = crucible.ExitTransformationFailed
	ExitDataCorrupt          = crucible.ExitDataCorrupt

	// Security & Authentication (70-79)
	// Authentication failures, authorization errors, security violations
	ExitAuthenticationFailed = crucible.ExitAuthenticationFailed
	ExitAuthorizationFailed  = crucible.ExitAuthorizationFailed
	ExitSecurityViolation    = crucible.ExitSecurityViolation
	ExitCertificateInvalid   = crucible.ExitCertificateInvalid

	// Observability & Monitoring (80-89)
	// Observability infrastructure failures
	ExitMetricsUnavailable      = crucible.ExitMetricsUnavailable
	ExitTracingFailed           = crucible.ExitTracingFailed
	ExitLoggingFailed           = crucible.ExitLoggingFailed
	ExitAlertSystemFailed       = crucible.ExitAlertSystemFailed
	ExitStructuredLoggingFailed = crucible.ExitStructuredLoggingFailed

	// Testing & Validation (91-99)
	// Test execution outcomes and validation failures
	ExitTestFailure             = crucible.ExitTestFailure
	ExitTestError               = crucible.ExitTestError
	ExitTestInterrupted         = crucible.ExitTestInterrupted
	ExitTestUsageError          = crucible.ExitTestUsageError
	ExitTestNoTestsCollected    = crucible.ExitTestNoTestsCollected
	ExitCoverageThresholdNotMet = crucible.ExitCoverageThresholdNotMet

	// Signal-Induced Exits (128-165)
	// Process terminated by Unix signals (128+N pattern)
	// NOTE: These codes are POSIX-specific and may not be supported on Windows.
	// Use SupportsSignalExitCodes() to check platform compatibility.
	ExitSignalHup  = crucible.ExitSignalHup
	ExitSignalInt  = crucible.ExitSignalInt
	ExitSignalQuit = crucible.ExitSignalQuit
	ExitSignalKill = crucible.ExitSignalKill
	ExitSignalPipe = crucible.ExitSignalPipe
	ExitSignalAlrm = crucible.ExitSignalAlrm
	ExitSignalTerm = crucible.ExitSignalTerm
	ExitSignalUsr1 = crucible.ExitSignalUsr1
	ExitSignalUsr2 = crucible.ExitSignalUsr2
)
