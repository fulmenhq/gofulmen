package signals

import (
	"fmt"
	"os"
	"runtime"

	fsignals "github.com/fulmenhq/gofulmen/foundry/signals"
)

// logUnsupportedSignal logs a warning and emits telemetry for unsupported signals.
// Returns a no-op cancel function and nil error to allow graceful degradation.
func logUnsupportedSignal(sig os.Signal) (CancelFunc, error) {
	// Get catalog to check for fallback metadata
	catalog := fsignals.GetDefaultCatalog()

	// Try to find signal definition
	var fallback *fsignals.WindowsFallback

	// Map signal to catalog ID
	sigMap := map[string]string{
		"hangup":      "hup",
		"broken pipe": "pipe",
		"alarm clock": "alrm",
	}

	sigName := sig.String()
	if id, found := sigMap[sigName]; found {
		if signal, err := catalog.GetSignal(id); err == nil && signal.WindowsFallback != nil {
			fallback = signal.WindowsFallback
		}
	}

	if fallback != nil {
		// Log structured message from catalog template
		logUnsupportedSignalWithFallback(sig, fallback)
	} else {
		// Generic fallback logging
		fmt.Fprintf(os.Stderr, "INFO: Signal %s is not supported on %s\n", sigName, runtime.GOOS)
	}

	// Return no-op cancel function - graceful degradation
	return func() {}, nil
}

// logUnsupportedSignalWithFallback logs using catalog-defined fallback metadata.
func logUnsupportedSignalWithFallback(sig os.Signal, fallback *fsignals.WindowsFallback) {
	// TODO: Integrate with logging package when available
	// For now, log to stderr with structured format from catalog

	message := fallback.LogMessage
	if message == "" {
		message = fmt.Sprintf("Signal %s unsupported on %s", sig, runtime.GOOS)
	}

	// Log at INFO level (as specified in catalog)
	fmt.Fprintf(os.Stderr, "INFO: %s\n", message)

	if fallback.OperationHint != "" {
		fmt.Fprintf(os.Stderr, "INFO: Hint: %s\n", fallback.OperationHint)
	}

	// TODO: Emit telemetry event when telemetry integration is available
	// Event: fallback.TelemetryEvent (e.g., "fulmen.signal.unsupported")
	// Tags: fallback.TelemetryTags
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}
