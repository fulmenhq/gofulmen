package appidentity

import (
	"testing"
)

// TestIdentityStructure verifies the Identity struct can be created and accessed.
func TestIdentityStructure(t *testing.T) {
	tests := []struct {
		name     string
		identity Identity
	}{
		{
			name: "minimal identity",
			identity: Identity{
				BinaryName: "testapp",
				Vendor:     "testvendor",
				EnvPrefix:  "TESTAPP_",
				ConfigName: "testapp",
			},
		},
		{
			name: "complete identity",
			identity: Identity{
				BinaryName:  "myapp",
				Vendor:      "myvendor",
				EnvPrefix:   "MYAPP_",
				ConfigName:  "myapp",
				Description: "Test application",
				Metadata: Metadata{
					ProjectURL:         "https://github.com/example/myapp",
					SupportEmail:       "support@example.com",
					License:            "MIT",
					RepositoryCategory: "cli",
					TelemetryNamespace: "myapp.metrics",
					RegistryID:         "01234567-89ab-cdef-0123-456789abcdef",
					Python: &PythonMetadata{
						DistributionName: "myapp",
						PackageName:      "myapp",
						ConsoleScripts: []ConsoleScript{
							{Name: "myapp", EntryPoint: "myapp.cli:main"},
						},
					},
					Extras: map[string]any{
						"custom_field": "custom_value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify struct fields are accessible
			if tt.identity.BinaryName == "" {
				t.Error("BinaryName should not be empty")
			}
			if tt.identity.Vendor == "" {
				t.Error("Vendor should not be empty")
			}
			if tt.identity.EnvPrefix == "" {
				t.Error("EnvPrefix should not be empty")
			}
			if tt.identity.ConfigName == "" {
				t.Error("ConfigName should not be empty")
			}
		})
	}
}

// TestIdentityGetters verifies all getter methods return correct values.
func TestIdentityGetters(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
		Metadata: Metadata{
			TelemetryNamespace: "custom.namespace",
		},
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Binary", identity.Binary(), "testapp"},
		{"VendorName", identity.VendorName(), "testvendor"},
		{"FlagsPrefix", identity.FlagsPrefix(), "testapp-"},
		{"ServiceName", identity.ServiceName(), "testapp"},
		{"TelemetryNamespace", identity.TelemetryNamespace(), "custom.namespace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestIdentityEnvVar verifies environment variable construction.
func TestIdentityEnvVar(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"CONFIG_PATH", "TESTAPP_CONFIG_PATH"},
		{"LOG_LEVEL", "TESTAPP_LOG_LEVEL"},
		{"DEBUG", "TESTAPP_DEBUG"},
		{"", "TESTAPP_"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := identity.EnvVar(tt.key)
			if got != tt.expected {
				t.Errorf("EnvVar(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}

// TestTelemetryNamespaceDefault verifies the default telemetry namespace behavior.
func TestTelemetryNamespaceDefault(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
		// No TelemetryNamespace set in Metadata
	}

	got := identity.TelemetryNamespace()
	expected := "testapp"

	if got != expected {
		t.Errorf("TelemetryNamespace() = %q, want %q (should default to BinaryName)", got, expected)
	}
}

// TestTelemetryNamespaceCustom verifies custom telemetry namespace.
func TestTelemetryNamespaceCustom(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
		Metadata: Metadata{
			TelemetryNamespace: "custom.metrics.namespace",
		},
	}

	got := identity.TelemetryNamespace()
	expected := "custom.metrics.namespace"

	if got != expected {
		t.Errorf("TelemetryNamespace() = %q, want %q", got, expected)
	}
}

// TestConfigParams verifies config parameter extraction.
func TestConfigParams(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
	}

	vendor, configName := identity.ConfigParams()

	if vendor != "testvendor" {
		t.Errorf("ConfigParams() vendor = %q, want %q", vendor, "testvendor")
	}

	if configName != "testapp" {
		t.Errorf("ConfigParams() configName = %q, want %q", configName, "testapp")
	}
}

// TestPythonMetadata verifies Python metadata handling.
func TestPythonMetadata(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
		Metadata: Metadata{
			Python: &PythonMetadata{
				DistributionName: "test-app",
				PackageName:      "test_app",
				ConsoleScripts: []ConsoleScript{
					{Name: "testapp", EntryPoint: "test_app.cli:main"},
					{Name: "testapp-admin", EntryPoint: "test_app.admin:main"},
				},
			},
		},
	}

	if identity.Metadata.Python == nil {
		t.Fatal("Python metadata should not be nil")
	}

	if identity.Metadata.Python.DistributionName != "test-app" {
		t.Errorf("DistributionName = %q, want %q",
			identity.Metadata.Python.DistributionName, "test-app")
	}

	if identity.Metadata.Python.PackageName != "test_app" {
		t.Errorf("PackageName = %q, want %q",
			identity.Metadata.Python.PackageName, "test_app")
	}

	if len(identity.Metadata.Python.ConsoleScripts) != 2 {
		t.Errorf("ConsoleScripts length = %d, want %d",
			len(identity.Metadata.Python.ConsoleScripts), 2)
	}
}

// TestExtrasMetadata verifies custom extras handling.
func TestExtrasMetadata(t *testing.T) {
	identity := Identity{
		BinaryName: "testapp",
		Vendor:     "testvendor",
		EnvPrefix:  "TESTAPP_",
		ConfigName: "testapp",
		Metadata: Metadata{
			Extras: map[string]any{
				"custom_string": "value",
				"custom_int":    42,
				"custom_bool":   true,
				"custom_map": map[string]any{
					"nested": "data",
				},
			},
		},
	}

	if identity.Metadata.Extras == nil {
		t.Fatal("Extras should not be nil")
	}

	if len(identity.Metadata.Extras) != 4 {
		t.Errorf("Extras length = %d, want %d", len(identity.Metadata.Extras), 4)
	}

	if identity.Metadata.Extras["custom_string"] != "value" {
		t.Errorf("Extras[custom_string] = %v, want %q",
			identity.Metadata.Extras["custom_string"], "value")
	}

	if identity.Metadata.Extras["custom_int"] != 42 {
		t.Errorf("Extras[custom_int] = %v, want %d",
			identity.Metadata.Extras["custom_int"], 42)
	}

	if identity.Metadata.Extras["custom_bool"] != true {
		t.Errorf("Extras[custom_bool] = %v, want %v",
			identity.Metadata.Extras["custom_bool"], true)
	}
}
