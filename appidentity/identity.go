package appidentity

// Identity represents application identity metadata from .fulmen/app.yaml.
//
// Identity provides core application metadata including binary name, vendor,
// environment variable prefixes, and configuration naming conventions. This
// information is used throughout the application for consistent configuration,
// logging, and telemetry setup.
type Identity struct {
	// BinaryName is the application's binary name (e.g., "gofulmen").
	// Must be lowercase, alphanumeric with hyphens only.
	BinaryName string `yaml:"binary_name" json:"binary_name"`

	// Vendor is the organization or vendor name (e.g., "fulmenhq").
	// Must be lowercase, alphanumeric with underscores only.
	Vendor string `yaml:"vendor" json:"vendor"`

	// EnvPrefix is the environment variable prefix (e.g., "GOFULMEN_").
	// Must be uppercase and end with an underscore.
	EnvPrefix string `yaml:"env_prefix" json:"env_prefix"`

	// ConfigName is the base name for configuration directories (e.g., "gofulmen").
	// Used for XDG config dir discovery.
	ConfigName string `yaml:"config_name" json:"config_name"`

	// Description is a one-line description of the application.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Metadata contains optional extended metadata.
	// Note: metadata is excluded from JSON marshaling of Identity because
	// the schema expects it as a sibling of "app", not nested within it.
	// Use the identityFile wrapper for proper YAML/JSON structure.
	Metadata Metadata `yaml:"metadata,omitempty" json:"-"`
}

// Metadata holds optional identity metadata for enhanced application information.
type Metadata struct {
	// ProjectURL is the primary project URL (e.g., GitHub repository).
	ProjectURL string `yaml:"project_url,omitempty" json:"project_url,omitempty"`

	// SupportEmail is the support or contact email address.
	SupportEmail string `yaml:"support_email,omitempty" json:"support_email,omitempty"`

	// License is the SPDX license identifier (e.g., "MIT", "Apache-2.0").
	License string `yaml:"license,omitempty" json:"license,omitempty"`

	// RepositoryCategory classifies the repository type
	// (e.g., "cli", "workhorse", "service", "library").
	RepositoryCategory string `yaml:"repository_category,omitempty" json:"repository_category,omitempty"`

	// TelemetryNamespace is the optional metrics namespace override.
	// If not set, BinaryName is used.
	TelemetryNamespace string `yaml:"telemetry_namespace,omitempty" json:"telemetry_namespace,omitempty"`

	// RegistryID is an optional UUIDv7 for future multi-app registry support.
	RegistryID string `yaml:"registry_id,omitempty" json:"registry_id,omitempty"`

	// Python contains Python-specific packaging metadata (optional).
	Python *PythonMetadata `yaml:"python,omitempty" json:"python,omitempty"`

	// Extras holds additional properties for extensibility.
	// Applications can store custom metadata here beyond the standard fields.
	//
	// The yaml:",inline" tag causes the YAML decoder to capture any metadata
	// fields not explicitly defined in the struct. For example, if your
	// .fulmen/app.yaml contains:
	//
	//   metadata:
	//     license: MIT              # Standard field → Metadata.License
	//     custom_field: value       # Custom field → Metadata.Extras["custom_field"]
	//     deployment_zone: us-west  # Custom field → Metadata.Extras["deployment_zone"]
	//
	// Standard fields (project_url, license, etc.) are marshaled to their
	// dedicated struct fields, while unknown fields are captured in Extras.
	// This enables forward-compatible extensibility without schema changes.
	Extras map[string]any `yaml:",inline" json:"-"`
}

// PythonMetadata contains Python-specific packaging metadata.
//
// This is used for applications that support both Go and Python implementations
// to maintain consistent identity across language boundaries.
type PythonMetadata struct {
	// DistributionName is the PyPI distribution name (e.g., "my-package").
	DistributionName string `yaml:"distribution_name,omitempty" json:"distribution_name,omitempty"`

	// PackageName is the Python import name (e.g., "my_package").
	PackageName string `yaml:"package_name,omitempty" json:"package_name,omitempty"`

	// ConsoleScripts lists console_scripts entry points.
	ConsoleScripts []ConsoleScript `yaml:"console_scripts,omitempty" json:"console_scripts,omitempty"`
}

// ConsoleScript represents a Python console_scripts entry point.
type ConsoleScript struct {
	// Name is the console script command name.
	Name string `yaml:"name" json:"name"`

	// EntryPoint is the module:function reference (e.g., "mypackage.cli:main").
	EntryPoint string `yaml:"entry_point" json:"entry_point"`
}

// Binary returns the application's binary name.
func (i *Identity) Binary() string {
	return i.BinaryName
}

// Vendor returns the vendor/organization name.
func (i *Identity) VendorName() string {
	return i.Vendor
}

// EnvVar constructs an environment variable name by appending the key to EnvPrefix.
//
// Example:
//
//	identity.EnvVar("CONFIG_PATH") // Returns "GOFULMEN_CONFIG_PATH"
func (i *Identity) EnvVar(key string) string {
	return i.EnvPrefix + key
}

// FlagsPrefix returns the CLI flags prefix derived from the binary name.
//
// Example:
//
//	identity.FlagsPrefix() // Returns "gofulmen-"
func (i *Identity) FlagsPrefix() string {
	return i.BinaryName + "-"
}

// TelemetryNamespace returns the telemetry metrics namespace.
//
// If Metadata.TelemetryNamespace is set, it returns that value.
// Otherwise, it returns the BinaryName as the default namespace.
func (i *Identity) TelemetryNamespace() string {
	if i.Metadata.TelemetryNamespace != "" {
		return i.Metadata.TelemetryNamespace
	}
	return i.BinaryName
}

// ServiceName returns the service name for logging and telemetry.
//
// This is used as the service identifier in structured logging and distributed tracing.
func (i *Identity) ServiceName() string {
	return i.BinaryName
}

// ConfigParams returns the vendor and config name parameters for config path derivation.
//
// This method provides the necessary parameters for integration with configpaths package
// without creating a direct dependency. Consumers can use:
//
//	vendor, name := identity.ConfigParams()
//	dirs := configpaths.GetAppConfigDirs(vendor, name)
func (i *Identity) ConfigParams() (vendor, configName string) {
	return i.Vendor, i.ConfigName
}
