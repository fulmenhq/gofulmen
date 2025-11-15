package fulpack

import (
	"time"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// globalTelemetrySystem is the package-level telemetry system.
// Initialized on first use, gracefully degrades if telemetry is unavailable.
var globalTelemetrySystem *telemetry.System

// initTelemetry initializes the global telemetry system if not already initialized.
func initTelemetry() {
	if globalTelemetrySystem != nil {
		return
	}

	config := telemetry.DefaultConfig()
	config.Enabled = true
	telSys, err := telemetry.NewSystem(config)
	if err != nil {
		// Gracefully degrade - operate without telemetry
		return
	}

	globalTelemetrySystem = telSys
}

// emitOperationMetrics emits standard operation telemetry.
func emitOperationMetrics(operation Operation, format ArchiveFormat, duration time.Duration, entryCount int, bytesProcessed int64, err error) {
	initTelemetry()
	if globalTelemetrySystem == nil {
		return
	}

	status := metrics.StatusSuccess
	if err != nil {
		status = metrics.StatusError
	}

	tags := map[string]string{
		metrics.TagOperation: string(operation),
		"format":             string(format),
		metrics.TagStatus:    status,
	}

	// Operation counter
	_ = globalTelemetrySystem.Counter(metrics.FulpackOperationsTotal, 1, tags)

	// Duration histogram
	_ = globalTelemetrySystem.Histogram(metrics.FulpackOperationMs, duration, tags)

	// Bytes processed counter
	if bytesProcessed > 0 {
		_ = globalTelemetrySystem.Counter(metrics.FulpackBytesProcessedTotal, float64(bytesProcessed), tags)
	}

	// Entry count counter
	if entryCount > 0 {
		_ = globalTelemetrySystem.Counter(metrics.FulpackEntriesTotal, float64(entryCount), tags)
	}

	// Error counter
	if err != nil {
		errorTags := map[string]string{
			metrics.TagOperation: string(operation),
			"format":             string(format),
		}
		if ferr, ok := err.(*FulpackError); ok {
			errorTags[metrics.TagErrorType] = ferr.Code
		} else {
			errorTags[metrics.TagErrorType] = "unknown"
		}
		_ = globalTelemetrySystem.Counter(metrics.FulpackErrorsTotal, 1, errorTags)
	}
}
