// Package signals provides signal handling utilities for graceful shutdown,
// config reload, and cross-platform signal management.
//
// This package implements the Fulmen signal handling architecture as defined
// in the Crucible catalog (v1.0.0), providing:
//
//   - Graceful shutdown with cleanup chains
//   - Ctrl+C double-tap semantics (2-second window for force quit)
//   - Config reload via restart with validation hooks
//   - Windows fallback behavior with logging and telemetry
//   - HTTP admin endpoint helpers for Windows signal simulation
//
// # Basic Usage
//
// Register a shutdown handler:
//
//	signals.OnShutdown(func(ctx context.Context) error {
//	    log.Println("Shutting down...")
//	    return server.Shutdown(ctx)
//	})
//
// Enable Ctrl+C double-tap:
//
//	signals.EnableDoubleTap(signals.DoubleTapConfig{
//	    Window:  2 * time.Second,
//	    Message: "Press Ctrl+C again to force quit",
//	})
//
// Register a config reload handler:
//
//	signals.OnReload(func(ctx context.Context) error {
//	    return config.Reload(ctx)
//	})
//
// Start listening for signals:
//
//	if err := signals.Listen(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
//
// # Unix vs Windows
//
// On Unix systems, the package registers OS signal handlers for SIGTERM, SIGINT,
// SIGHUP, etc. On Windows, some signals are unsupported:
//
//   - SIGTERM/SIGINT: Supported via CTRL_CLOSE_EVENT and CTRL_C_EVENT
//   - SIGHUP: Not supported - use HTTP admin endpoint fallback
//   - SIGPIPE: Not supported - use exception handling
//
// The package automatically logs INFO messages and emits telemetry events when
// unsupported signals are registered on Windows.
//
// # Cleanup Chains
//
// Cleanup functions are executed in reverse registration order (LIFO) with
// configurable timeouts:
//
//	signals.OnShutdown(func(ctx context.Context) error {
//	    // First registered, last executed
//	    return closeDatabase(ctx)
//	})
//
//	signals.OnShutdown(func(ctx context.Context) error {
//	    // Last registered, first executed
//	    return stopWorkers(ctx)
//	})
//
// # Config Reload
//
// Config reload enforces validation before restart:
//
//	signals.OnReload(func(ctx context.Context) error {
//	    // Validate new config against schema
//	    if err := validateConfig(newConfig); err != nil {
//	        return fmt.Errorf("validation failed: %w", err)
//	    }
//	    // Proceed with restart
//	    return restartProcess(ctx)
//	})
//
// On validation failure, the process continues with the old config.
//
// # Thread Safety
//
// All public APIs are thread-safe and can be called concurrently.
package signals
