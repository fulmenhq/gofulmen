package appidentity

// Testing utilities for working with application identity in tests.
//
// This file provides exported test utilities that both internal tests and
// external consumers can use to create test fixtures and override identity
// in their test suites.

// NewFixture creates a test Identity with sensible defaults.
//
// The fixture can be customized by providing optional override functions.
// This is useful for creating test identities without needing YAML files.
//
// Example:
//
//	identity := NewFixture()
//	// Returns minimal valid identity with defaults
//
//	identity := NewFixture(func(id *Identity) {
//	    id.BinaryName = "mytestapp"
//	    id.EnvPrefix = "MYTESTAPP_"
//	})
//	// Returns identity with custom values
//
//	identity := NewFixture(func(id *Identity) {
//	    id.Metadata.License = "Apache-2.0"
//	    id.Metadata.ProjectURL = "https://github.com/example/app"
//	})
//	// Returns identity with metadata overrides
func NewFixture(overrides ...func(*Identity)) *Identity {
	// Start with minimal valid identity
	identity := &Identity{
		BinaryName:  "testapp",
		Vendor:      "testvendor",
		EnvPrefix:   "TESTAPP_",
		ConfigName:  "testapp",
		Description: "Test application fixture",
	}

	// Apply any overrides
	for _, override := range overrides {
		override(identity)
	}

	return identity
}

// NewCompleteFixture creates a test Identity with all fields populated.
//
// This is useful for testing code that relies on optional metadata fields.
// The fixture can be customized using override functions.
//
// Example:
//
//	identity := NewCompleteFixture()
//	// Returns identity with all fields including Python metadata
//
//	identity := NewCompleteFixture(func(id *Identity) {
//	    id.Metadata.License = "GPL-3.0"
//	})
//	// Returns complete identity with license override
func NewCompleteFixture(overrides ...func(*Identity)) *Identity {
	identity := &Identity{
		BinaryName:  "myapp",
		Vendor:      "myvendor",
		EnvPrefix:   "MYAPP_",
		ConfigName:  "myapp",
		Description: "Complete test application with all metadata",
		Metadata: Metadata{
			ProjectURL:         "https://github.com/example/myapp",
			SupportEmail:       "support@example.com",
			License:            "MIT",
			RepositoryCategory: "cli",
			TelemetryNamespace: "myapp_metrics",
			RegistryID:         "01234567-89ab-cdef-0123-456789abcdef",
			Python: &PythonMetadata{
				DistributionName: "my-app",
				PackageName:      "my_app",
				ConsoleScripts: []ConsoleScript{
					{
						Name:       "myapp",
						EntryPoint: "my_app.cli:main",
					},
					{
						Name:       "myapp-admin",
						EntryPoint: "my_app.admin:main",
					},
				},
			},
			Extras: map[string]any{
				"custom_field":  "custom_value",
				"custom_number": 42,
			},
		},
	}

	// Apply any overrides
	for _, override := range overrides {
		override(identity)
	}

	return identity
}
