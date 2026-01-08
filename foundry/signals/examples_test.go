package signals_test

import (
	"fmt"

	"github.com/fulmenhq/gofulmen/foundry/signals"
)

func ExampleGetDefaultCatalog() {
	catalog := signals.GetDefaultCatalog()
	version, _ := catalog.Version()
	fmt.Printf("Signals catalog version present: %t\n", version != "")
	// Output: Signals catalog version present: true
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

	signalTerm, _ := catalog.GetSignal("term")
	signalInt, _ := catalog.GetSignal("int")
	signalHup, _ := catalog.GetSignal("hup")

	fmt.Println("Signals:")
	fmt.Printf("  %s (%s)\n", signalTerm.ID, signalTerm.Name)
	fmt.Printf("  %s (%s)\n", signalInt.ID, signalInt.Name)
	fmt.Printf("  %s (%s)\n", signalHup.ID, signalHup.Name)
	// Output:
	// Signals:
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

func ExampleCatalog_ResolveSignal() {
	catalog := signals.GetDefaultCatalog()

	// ResolveSignal accepts various input formats
	variants := []string{"SIGTERM", "sigterm", "TERM", "term", "15", "  term  "}

	for _, input := range variants {
		signal := catalog.ResolveSignal(input)
		if signal != nil {
			fmt.Printf("%q -> %s\n", input, signal.Name)
		}
	}

	// Unknown signals return nil
	unknown := catalog.ResolveSignal("SIGFOO")
	fmt.Printf("Unknown signal: %v\n", unknown)
	// Output:
	// "SIGTERM" -> SIGTERM
	// "sigterm" -> SIGTERM
	// "TERM" -> SIGTERM
	// "term" -> SIGTERM
	// "15" -> SIGTERM
	// "  term  " -> SIGTERM
	// Unknown signal: <nil>
}

func ExampleCatalog_ListSignalNames() {
	catalog := signals.GetDefaultCatalog()

	names, err := catalog.ListSignalNames()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Has SIGTERM: %t\n", contains(names, "SIGTERM"))
	fmt.Printf("Has SIGINT: %t\n", contains(names, "SIGINT"))
	fmt.Printf("Has SIGHUP: %t\n", contains(names, "SIGHUP"))
	// Output:
	// Has SIGTERM: true
	// Has SIGINT: true
	// Has SIGHUP: true
}

func ExampleCatalog_MatchSignalNames() {
	catalog := signals.GetDefaultCatalog()

	usrSignals, _ := catalog.MatchSignalNames("*USR*")
	fmt.Printf("Matches SIGUSR1: %t\n", contains(usrSignals, "SIGUSR1"))
	fmt.Printf("Matches SIGUSR2: %t\n", contains(usrSignals, "SIGUSR2"))

	shortSignals, _ := catalog.MatchSignalNames("SIG???")
	fmt.Printf("Matches SIGINT: %t\n", contains(shortSignals, "SIGINT"))
	fmt.Printf("Matches SIGHUP: %t\n", contains(shortSignals, "SIGHUP"))

	trimmedSignals, _ := catalog.MatchSignalNames("  sigterm  ")
	fmt.Printf("Trimmed pattern matches SIGTERM: %t\n", contains(trimmedSignals, "SIGTERM"))
	// Output:
	// Matches SIGUSR1: true
	// Matches SIGUSR2: true
	// Matches SIGINT: true
	// Matches SIGHUP: true
	// Trimmed pattern matches SIGTERM: true
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
