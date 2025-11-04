package appidentity_test

import (
	"context"
	"fmt"
	"log"

	"github.com/fulmenhq/gofulmen/appidentity"
)

// ExampleGet demonstrates loading application identity from .fulmen/app.yaml.
func ExampleGet() {
	ctx := context.Background()

	// Get loads identity from .fulmen/app.yaml using discovery rules
	identity, err := appidentity.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to load identity: %v", err)
	}

	fmt.Printf("Application: %s\n", identity.Binary())
	fmt.Printf("Vendor: %s\n", identity.VendorName())
	fmt.Printf("Env prefix: %s\n", identity.EnvPrefix)
}

// ExampleMust demonstrates the panic variant of Get.
func ExampleMust() {
	ctx := context.Background()

	// Must panics if identity cannot be loaded
	identity := appidentity.Must(ctx)

	fmt.Printf("Binary: %s\n", identity.Binary())
}

// ExampleLoadFrom demonstrates loading identity from an explicit path.
func ExampleLoadFrom() {
	ctx := context.Background()

	// LoadFrom loads identity from a specific file path
	identity, err := appidentity.LoadFrom(ctx, ".fulmen/app.yaml")
	if err != nil {
		log.Fatalf("Failed to load identity: %v", err)
	}

	fmt.Printf("Loaded from explicit path: %s\n", identity.Binary())
}

// ExampleWithIdentity demonstrates context-based identity override for testing.
func ExampleWithIdentity() {
	ctx := context.Background()

	// Create a test identity
	testIdentity := &appidentity.Identity{
		BinaryName:  "testapp",
		Vendor:      "testvendor",
		EnvPrefix:   "TESTAPP_",
		ConfigName:  "testapp",
		Description: "Test application",
	}

	// Inject identity into context for testing
	ctx = appidentity.WithIdentity(ctx, testIdentity)

	// Get now returns the injected identity
	identity, _ := appidentity.Get(ctx)
	fmt.Printf("Test identity: %s\n", identity.Binary())

	// Output: Test identity: testapp
}

// ExampleIdentity_EnvVar demonstrates constructing environment variable names.
func ExampleIdentity_EnvVar() {
	identity := &appidentity.Identity{
		BinaryName:  "myapp",
		EnvPrefix:   "MYAPP_",
		ConfigName:  "myapp",
		Description: "Example application",
	}

	// Construct environment variable names
	configPathVar := identity.EnvVar("CONFIG_PATH")
	logLevelVar := identity.EnvVar("LOG_LEVEL")

	fmt.Printf("Config path var: %s\n", configPathVar)
	fmt.Printf("Log level var: %s\n", logLevelVar)

	// Output:
	// Config path var: MYAPP_CONFIG_PATH
	// Log level var: MYAPP_LOG_LEVEL
}

// ExampleIdentity_ConfigParams demonstrates getting config path parameters.
func ExampleIdentity_ConfigParams() {
	identity := &appidentity.Identity{
		BinaryName:  "myapp",
		Vendor:      "myvendor",
		ConfigName:  "myapp",
		Description: "Example application",
	}

	// Get parameters for config path derivation
	vendor, name := identity.ConfigParams()

	// These can be passed to configpaths package (avoiding direct dependency)
	fmt.Printf("Vendor: %s, Config name: %s\n", vendor, name)

	// Output: Vendor: myvendor, Config name: myapp
}

// ExampleIdentity_FlagsPrefix demonstrates CLI flag prefix derivation.
func ExampleIdentity_FlagsPrefix() {
	identity := &appidentity.Identity{
		BinaryName:  "gofulmen",
		ConfigName:  "gofulmen",
		Description: "Example application",
	}

	// Get CLI flag prefix
	prefix := identity.FlagsPrefix()

	fmt.Printf("Flags prefix: %s\n", prefix)
	fmt.Printf("Example flag: %sconfig\n", prefix)

	// Output:
	// Flags prefix: gofulmen-
	// Example flag: gofulmen-config
}

// ExampleIdentity_TelemetryNamespace demonstrates telemetry namespace derivation.
func ExampleIdentity_TelemetryNamespace() {
	// Without explicit namespace - uses binary name
	identity1 := &appidentity.Identity{
		BinaryName:  "myapp",
		ConfigName:  "myapp",
		Description: "Example application",
	}

	fmt.Printf("Default namespace: %s\n", identity1.TelemetryNamespace())

	// With explicit namespace override
	identity2 := &appidentity.Identity{
		BinaryName:  "myapp",
		ConfigName:  "myapp",
		Description: "Example application",
		Metadata: appidentity.Metadata{
			TelemetryNamespace: "custom_metrics",
		},
	}

	fmt.Printf("Custom namespace: %s\n", identity2.TelemetryNamespace())

	// Output:
	// Default namespace: myapp
	// Custom namespace: custom_metrics
}

// ExampleIdentity_ServiceName demonstrates service name for logging.
func ExampleIdentity_ServiceName() {
	identity := &appidentity.Identity{
		BinaryName:  "myapp",
		ConfigName:  "myapp",
		Description: "Example application",
	}

	// Get service name for structured logging
	serviceName := identity.ServiceName()

	fmt.Printf("Service name: %s\n", serviceName)

	// Output: Service name: myapp
}

// ExampleNewFixture demonstrates creating test fixtures.
func ExampleNewFixture() {
	// Create default test fixture
	identity := appidentity.NewFixture()
	fmt.Printf("Default: %s\n", identity.Binary())

	// Create fixture with overrides
	customIdentity := appidentity.NewFixture(func(id *appidentity.Identity) {
		id.BinaryName = "customapp"
		id.Metadata.License = "Apache-2.0"
	})
	fmt.Printf("Custom: %s (license: %s)\n",
		customIdentity.Binary(),
		customIdentity.Metadata.License)

	// Output:
	// Default: testapp
	// Custom: customapp (license: Apache-2.0)
}

// ExampleNewCompleteFixture demonstrates creating complete test fixtures.
func ExampleNewCompleteFixture() {
	// Create fixture with all fields populated
	identity := appidentity.NewCompleteFixture()

	fmt.Printf("Binary: %s\n", identity.Binary())
	fmt.Printf("License: %s\n", identity.Metadata.License)
	fmt.Printf("Python dist: %s\n", identity.Metadata.Python.DistributionName)

	// Output:
	// Binary: myapp
	// License: MIT
	// Python dist: my-app
}

// ExampleValidate demonstrates validating an identity file.
func ExampleValidate() {
	ctx := context.Background()

	// Validate identity file against schema
	if err := appidentity.Validate(ctx, ".fulmen/app.yaml"); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Println("Identity file is valid")
}
