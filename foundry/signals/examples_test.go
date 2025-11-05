package signals_test

import (
	"fmt"

	"github.com/fulmenhq/gofulmen/foundry/signals"
)

func ExampleGetDefaultCatalog() {
	catalog := signals.GetDefaultCatalog()
	version, _ := catalog.Version()
	fmt.Printf("Signals catalog version: %s\n", version)
	// Output: Signals catalog version: v1.0.0
}

func ExampleCatalog_GetSignal() {
	catalog := signals.GetDefaultCatalog()

	// Get SIGTERM signal definition
	signal, err := catalog.GetSignal("term")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Signal: %s\n", signal.Name)
	fmt.Printf("Unix number: %d\n", signal.UnixNumber)
	fmt.Printf("Exit code: %d\n", signal.ExitCode)
	fmt.Printf("Timeout: %ds\n", signal.TimeoutSeconds)
	// Output:
	// Signal: SIGTERM
	// Unix number: 15
	// Exit code: 143
	// Timeout: 30s
}

func ExampleCatalog_GetSignalByName() {
	catalog := signals.GetDefaultCatalog()

	// Get signal by name
	signal, err := catalog.GetSignalByName("SIGINT")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Signal ID: %s\n", signal.ID)
	fmt.Printf("Default behavior: %s\n", signal.DefaultBehavior)
	// Output:
	// Signal ID: int
	// Default behavior: graceful_shutdown_with_double_tap
}

func ExampleCatalog_ListSignals() {
	catalog := signals.GetDefaultCatalog()

	signals, err := catalog.ListSignals()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Catalog contains %d signals\n", len(signals))
	for _, signal := range signals[:3] { // Show first 3
		fmt.Printf("  %s (%s)\n", signal.ID, signal.Name)
	}
	// Output:
	// Catalog contains 8 signals
	//   term (SIGTERM)
	//   int (SIGINT)
	//   hup (SIGHUP)
}

func Example_doubleTap() {
	catalog := signals.GetDefaultCatalog()

	signal, _ := catalog.GetSignal("int")

	if signal.DoubleTapWindowSeconds != nil {
		fmt.Printf("Double-tap window: %ds\n", *signal.DoubleTapWindowSeconds)
		fmt.Printf("Message: %s\n", signal.DoubleTapMessage)
	}
	// Output:
	// Double-tap window: 2s
	// Message: Press Ctrl+C again within 2s to force quit
}

func Example_windowsFallback() {
	catalog := signals.GetDefaultCatalog()

	signal, _ := catalog.GetSignal("hup")

	if signal.WindowsFallback != nil {
		fmt.Printf("Fallback behavior: %s\n", signal.WindowsFallback.FallbackBehavior)
		fmt.Printf("Telemetry event: %s\n", signal.WindowsFallback.TelemetryEvent)
		fmt.Printf("Operation hint: %s\n", signal.WindowsFallback.OperationHint)
	}
	// Output:
	// Fallback behavior: http_admin_endpoint
	// Telemetry event: fulmen.signal.unsupported
	// Operation hint: POST /admin/signal with signal=HUP
}

func Example_reloadSemantics() {
	catalog := signals.GetDefaultCatalog()

	signal, _ := catalog.GetSignal("hup")

	fmt.Printf("Reload strategy: %s\n", signal.ReloadStrategy)
	if signal.ValidationRequired != nil && *signal.ValidationRequired {
		fmt.Println("Validation required: yes")
	}
	fmt.Println("Cleanup actions:")
	for _, action := range signal.CleanupActions {
		fmt.Printf("  - %s\n", action)
	}
	// Output:
	// Reload strategy: restart_based
	// Validation required: yes
	// Cleanup actions:
	//   - validate_new_config_against_schema
	//   - graceful_shutdown
	//   - restart_with_new_config
	//   - log_reload
}
