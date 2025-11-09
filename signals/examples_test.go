package signals_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fulmenhq/gofulmen/signals"
)

func ExampleOnShutdown() {
	// Register a cleanup function for graceful shutdown
	signals.OnShutdown(func(ctx context.Context) error {
		fmt.Println("Shutting down gracefully...")
		return nil
	})

	// Register another cleanup function
	// These execute in reverse order (LIFO)
	signals.OnShutdown(func(ctx context.Context) error {
		fmt.Println("Closing connections...")
		return nil
	})

	// In a real application, you would call signals.Listen(ctx)
	// Output:
}

func ExampleOnReload() {
	// Register a config reload handler
	signals.OnReload(func(ctx context.Context) error {
		fmt.Println("Reloading configuration...")
		// Validate new config
		// If validation fails, return error to abort reload
		return nil
	})

	// Reload handlers execute in registration order with fail-fast
	// Output:
}

func ExampleEnableDoubleTap() {
	// Enable Ctrl+C double-tap with custom settings
	err := signals.EnableDoubleTap(signals.DoubleTapConfig{
		Window:   2 * time.Second,
		Message:  "Press Ctrl+C again within 2s to force quit",
		ExitCode: 130,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Now pressing Ctrl+C twice within 2 seconds will force immediate exit
	// Output:
}

func ExampleEnableDoubleTap_defaults() {
	// Enable double-tap with catalog defaults
	err := signals.EnableDoubleTap(signals.DoubleTapConfig{})
	if err != nil {
		log.Fatal(err)
	}

	// Uses defaults from Crucible catalog:
	// - Window: 2 seconds
	// - Message: "Press Ctrl+C again within 2s to force quit"
	// - Exit code: 130
	// Output:
}

func ExampleSetQuietMode() {
	// Disable double-tap messages for non-interactive services
	signals.SetQuietMode(true)

	err := signals.EnableDoubleTap(signals.DoubleTapConfig{})
	if err != nil {
		log.Fatal(err)
	}

	// Now double-tap message won't be printed to stderr
	// Output:
}

func ExampleNewHTTPHandler() {
	// Create HTTP handler for /admin/signal endpoint
	config := signals.HTTPConfig{
		TokenAuth: os.Getenv("SIGNAL_ADMIN_TOKEN"),
		RateLimit: 6, // requests per minute
		RateBurst: 3,
	}

	handler := signals.NewHTTPHandler(config)

	// Wire to your HTTP server
	http.Handle("/admin/signal", handler)

	// Example request:
	// POST /admin/signal
	// Authorization: Bearer <token>
	// {"signal": "SIGHUP", "reason": "config reload", "requester": "admin"}
	// Output:
}

func ExampleHTTPConfig_mTLS() {
	// Configure with mTLS verification
	config := signals.HTTPConfig{
		MTLSVerify: true,
		RateLimit:  6,
		RateBurst:  3,
	}

	handler := signals.NewHTTPHandler(config)

	// Your TLS server config should verify client certificates
	server := &http.Server{
		Addr:    ":8443",
		Handler: handler,
		// TLSConfig: ... (configure client cert verification)
	}

	_ = server // Use server
	// Output:
}

func ExampleSupports() {
	// Check if a signal is supported on the current platform
	if signals.Supports(os.Signal(nil)) {
		fmt.Println("Signal supported")
	}

	// On Windows, SIGHUP is not supported
	// On Unix, all standard signals are supported
	// Output:
}

func ExampleVersion() {
	// Get the signal catalog version
	version, err := signals.Version()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Signals catalog version: %s\n", version)
	// Output: Signals catalog version: v1.0.0
}

// Example_basicUsage demonstrates a complete signal handling setup.
func Example_basicUsage() {
	// Register shutdown cleanup
	signals.OnShutdown(func(ctx context.Context) error {
		fmt.Println("Closing database...")
		return nil
	})

	signals.OnShutdown(func(ctx context.Context) error {
		fmt.Println("Stopping workers...")
		return nil
	})

	// Enable double-tap for force quit
	_ = signals.EnableDoubleTap(signals.DoubleTapConfig{
		Window:  2 * time.Second,
		Message: "Force quit with Ctrl+C again",
	})

	// In a real app, you would:
	// ctx := context.Background()
	// if err := signals.Listen(ctx); err != nil {
	//     log.Fatal(err)
	// }

	// Output:
}

// Example_httpEndpoint demonstrates setting up the HTTP admin endpoint.
func Example_httpEndpoint() {
	// Set up signal handlers
	signals.OnShutdown(func(ctx context.Context) error {
		fmt.Println("Shutting down...")
		return nil
	})

	signals.OnReload(func(ctx context.Context) error {
		fmt.Println("Reloading config...")
		return nil
	})

	// Create HTTP handler
	config := signals.HTTPConfig{
		TokenAuth: "my-secret-token",
		RateLimit: 6,
		RateBurst: 3,
	}

	handler := signals.NewHTTPHandler(config)

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle("/admin/signal", handler)

	// In production:
	// http.ListenAndServe(":8080", mux)

	// Example usage on Windows (since SIGHUP is unsupported):
	// curl -X POST http://localhost:8080/admin/signal \
	//   -H "Authorization: Bearer my-secret-token" \
	//   -H "Content-Type: application/json" \
	//   -d '{"signal": "SIGHUP", "reason": "config reload"}'

	// Output:
}

// Example_windowsFallback demonstrates Windows signal fallback behavior.
func Example_windowsFallback() {
	// On Windows, some signals are unsupported
	// The library logs INFO messages with operation hints

	// Register handlers (works on all platforms)
	signals.OnShutdown(func(ctx context.Context) error {
		return nil
	})

	// On Windows, registering SIGHUP logs:
	// INFO: SIGHUP unavailable on Windows - use HTTP endpoint for config reload
	// INFO: Hint: POST /admin/signal with signal=HUP

	// The handler still returns without error (graceful degradation)
	// Applications should provide HTTP endpoint as fallback

	// Output:
}
