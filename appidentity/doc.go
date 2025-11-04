// Package appidentity provides application identity metadata discovery and access.
//
// This package enables Go applications to discover and load standardized identity
// metadata from .fulmen/app.yaml files. The identity metadata includes information
// about the application's name, vendor, environment variable prefixes, and configuration
// naming conventions.
//
// # Discovery
//
// Identity files are discovered using the following precedence order:
//
//  1. Context injection (for testing) - WithIdentity(ctx, identity)
//  2. Explicit path via Options - GetWithOptions(ctx, Options{ExplicitPath: path})
//  3. Environment variable - FULMEN_APP_IDENTITY_PATH
//  4. Nearest ancestor search - Walk from cwd to filesystem root looking for .fulmen/app.yaml
//
// # Usage
//
// Basic usage with automatic discovery:
//
//	identity, err := appidentity.Get(ctx)
//	if err != nil {
//	    return fmt.Errorf("failed to load identity: %w", err)
//	}
//	fmt.Println("Binary:", identity.Binary())
//
// Using Must variant (panics on error):
//
//	identity := appidentity.Must(ctx)
//	fmt.Println("Binary:", identity.Binary())
//
// Loading from explicit path:
//
//	identity, err := appidentity.LoadFrom(ctx, "/path/to/.fulmen/app.yaml")
//	if err != nil {
//	    return fmt.Errorf("failed to load identity: %w", err)
//	}
//
// # Caching
//
// Identity is loaded once per process and cached. Subsequent calls to Get()
// return the cached instance. This behavior is thread-safe using sync.Once.
//
// For testing, use WithIdentity() to inject a custom identity via context:
//
//	testIdentity := &appidentity.Identity{
//	    BinaryName: "testapp",
//	    Vendor:     "testvendor",
//	    EnvPrefix:  "TESTAPP_",
//	}
//	ctx = appidentity.WithIdentity(ctx, testIdentity)
//	identity, _ := appidentity.Get(ctx) // Returns testIdentity
//
// # Integration
//
// The Identity struct provides helper methods for integration with configuration,
// logging, and telemetry systems:
//
//	// Environment variables
//	envVar := identity.EnvVar("CONFIG_PATH") // Returns "MYAPP_CONFIG_PATH"
//
//	// CLI flags
//	flagPrefix := identity.FlagsPrefix() // Returns "myapp-"
//
//	// Telemetry
//	namespace := identity.TelemetryNamespace() // Returns custom or binary name
//
// # Validation
//
// Identity files are validated against the Crucible schema on load. Validation
// errors include field-level diagnostics with specific constraint violations.
//
// To validate an identity file without loading it:
//
//	if err := appidentity.Validate(ctx, "/path/to/.fulmen/app.yaml"); err != nil {
//	    fmt.Println("Validation failed:", err)
//	}
//
// # Layer 0 Module
//
// This package is a Layer 0 module with no dependencies on other Fulmen packages
// (except schema for validation). Other packages like config, logging, and telemetry
// can safely depend on appidentity without creating import cycles.
package appidentity
