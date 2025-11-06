# App Identity Module

The **appidentity** package provides application identity metadata management from `.fulmen/app.yaml` files. It standardizes configuration paths, environment variables, and telemetry namespaces across Go applications in the Fulmen ecosystem.

## Purpose

The appidentity module addresses common application identity challenges:

- **Consistent Configuration**: Standard paths for config, logs, and data
- **Environment Variables**: Systematic naming with proper prefixes
- **Telemetry Integration**: Consistent namespaces for metrics and logging
- **Testing Support**: Clean test isolation without filesystem dependencies
- **Zero Dependencies**: Layer 0 module with no Fulmen dependencies

## Key Features

- **Automatic Discovery**: Searches ancestor directories for `.fulmen/app.yaml`
- **Schema Validation**: Validates against Crucible v1.0.0 app-identity schema
- **Process-Level Caching**: Thread-safe singleton with sync.Once
- **Context Injection**: Test-friendly overrides via context
- **Multiple Sources**: Context → Explicit path → Env var → Discovery
- **Rich Metadata**: Optional project URL, license, repository category, etc.

## ⚠️ YAML Structure - CRITICAL

The `.fulmen/app.yaml` file **MUST** have an `app:` key at the root level wrapping identity fields. This is required by the Crucible app-identity schema v1.0.0.

### ✅ CORRECT YAML Structure

```yaml
# .fulmen/app.yaml - Required structure
app:
  binary_name: myapp
  vendor: myvendor
  env_prefix: MYAPP_
  config_name: myapp
  description: My application description

# Optional metadata section
metadata:
  project_url: https://github.com/example/myapp
  license: MIT
  repository_category: cli
```

### ❌ INCORRECT - Fields at Root Level

```yaml
# This will NOT work - fields must be under "app:" key
binary_name: myapp # ❌ Wrong - at root level
vendor: myvendor # ❌ Wrong - at root level
env_prefix: MYAPP_ # ❌ Wrong - at root level
```

**Why the `app:` wrapper?**

The nested structure allows the YAML format to include additional top-level keys (`metadata:`, `python:`, etc.) without polluting the identity namespace. It also ensures forward compatibility with future schema extensions.

**Common Mistake:**

If you see empty fields when loading identity (`BinaryName: "", Vendor: ""`), check that your YAML has the `app:` key wrapper. See `appidentity/testdata/*.yaml` for correct examples.

## Basic Usage

### Loading Identity

```go
package main

import (
    "context"
    "log"

    "github.com/fulmenhq/gofulmen/appidentity"
)

func main() {
    ctx := context.Background()

    // Load identity (auto-discovers .fulmen/app.yaml)
    identity, err := appidentity.Get(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Use identity fields
    log.Printf("Binary: %s", identity.BinaryName)
    log.Printf("Vendor: %s", identity.Vendor)
    log.Printf("Env Prefix: %s", identity.EnvPrefix)
}
```

### Configuration Paths

```go
import (
    "github.com/fulmenhq/gofulmen/appidentity"
    "github.com/fulmenhq/gofulmen/config"
)

identity, _ := appidentity.Get(ctx)

// Get XDG-compliant config params
vendor, name := identity.ConfigParams()
configDir := config.GetAppConfigDir(vendor, name)
// Example: /home/user/.config/myvendor/myapp
```

### Environment Variables

```go
identity, _ := appidentity.Get(ctx)

// Construct environment variable names
logLevel := identity.EnvVar("LOG_LEVEL")
// Returns: "MYAPP_LOG_LEVEL" (using identity.EnvPrefix)

// Use with os.Getenv
level := os.Getenv(logLevel)
```

### Telemetry Integration

```go
identity, _ := appidentity.Get(ctx)

// Get telemetry namespace
namespace := identity.TelemetryNamespace()
// Returns: "myvendor.myapp" or custom metadata.telemetry_namespace

// Use with logging/telemetry
logger := logging.New(namespace, logging.WithProfile(logging.ProfileSimple))
```

## Discovery Precedence

Identity loading follows this precedence order (highest to lowest):

1. **Context Injection**: `appidentity.WithIdentity(ctx, identity)`
2. **Explicit Path**: `GetWithOptions(ctx, Options{ExplicitPath: "/path/to/app.yaml"})`
3. **Environment Variable**: `FULMEN_APP_IDENTITY_PATH=/path/to/app.yaml`
4. **Ancestor Search**: Searches up to 20 parent directories for `.fulmen/app.yaml`

### Explicit Path Loading

```go
// Load from specific path
identity, err := appidentity.LoadFrom(ctx, "/custom/path/app.yaml")
```

### Environment Variable

```bash
# Set environment variable
export FULMEN_APP_IDENTITY_PATH=/custom/path/app.yaml

# Load identity (uses env var)
identity, err := appidentity.Get(ctx)
```

## Testing Support

### Context Injection

```go
func TestMyFunction(t *testing.T) {
    // Create test identity
    testIdentity := appidentity.NewFixture(func(id *appidentity.Identity) {
        id.BinaryName = "testapp"
        id.Vendor = "testvendor"
        id.EnvPrefix = "TEST_"
    })

    // Inject into context
    ctx := appidentity.WithIdentity(context.Background(), testIdentity)

    // Test code uses injected identity
    identity, _ := appidentity.Get(ctx)
    // Returns testIdentity (no filesystem access)
}
```

### Complete Fixture

```go
// Create fixture with all fields populated
testIdentity := appidentity.NewCompleteFixture()
testIdentity.BinaryName = "myapp"
testIdentity.Vendor = "myvendor"

ctx := appidentity.WithIdentity(context.Background(), testIdentity)
```

### Cache Reset

```go
func TestIdentityLoading(t *testing.T) {
    // Reset cache before test
    appidentity.Reset()

    // Load identity
    identity, err := appidentity.Get(ctx)

    // Test assertions...
}
```

**Warning:** `Reset()` is NOT safe during concurrent `Get()` calls. Only use in test cleanup.

## Schema Validation

Identity files are validated against the Crucible v1.0.0 app-identity schema:

```go
identity, _ := appidentity.LoadFrom(ctx, "/path/to/app.yaml")

// Validate against schema
if err := appidentity.Validate(identity); err != nil {
    if verr, ok := err.(*appidentity.ValidationError); ok {
        for _, detail := range verr.Details {
            log.Printf("Validation error at %s: %s", detail.Field, detail.Message)
        }
    }
}
```

### Required Fields

- `app.binary_name`: Lowercase alphanumeric with hyphens (e.g., "my-app")
- `app.vendor`: Lowercase alphanumeric with underscores (e.g., "my_vendor")
- `app.env_prefix`: Uppercase, must end with underscore (e.g., "MYAPP\_")
- `app.config_name`: Lowercase alphanumeric with hyphens (e.g., "my-app")

### Field Validation Rules

```go
// Valid examples
binary_name: "myapp"           // ✅ Valid
binary_name: "my-app"          // ✅ Valid
binary_name: "MyApp"           // ❌ Invalid (must be lowercase)

vendor: "myvendor"             // ✅ Valid
vendor: "my_vendor"            // ✅ Valid
vendor: "my-vendor"            // ❌ Invalid (use underscore, not hyphen)

env_prefix: "MYAPP_"           // ✅ Valid (trailing underscore)
env_prefix: "MY_APP_"          // ✅ Valid
env_prefix: "MYAPP"            // ❌ Invalid (missing trailing underscore)

config_name: "myapp"           // ✅ Valid
config_name: "my-app"          // ✅ Valid
```

## Complete YAML Example

```yaml
# .fulmen/app.yaml
app:
  binary_name: gofulmen
  vendor: fulmen
  env_prefix: GOFULMEN_
  config_name: gofulmen
  description: Go foundation library for FulmenHQ ecosystem

metadata:
  project_url: https://github.com/fulmenhq/gofulmen
  support_email: support@fulmen.dev
  license: MIT
  repository_category: library
  telemetry_namespace: fulmen.gofulmen

  # Python packaging (optional)
  python:
    distribution_name: gofulmen
    package_name: gofulmen
    console_scripts:
      - name: gofulmen
        entry_point: gofulmen.cli:main

  # Custom extensibility fields
  deployment_zone: us-west-2
  team: platform
```

## Integration Examples

### With Config Module

```go
identity, _ := appidentity.Get(ctx)
vendor, name := identity.ConfigParams()

// XDG config directory
configDir := config.GetAppConfigDir(vendor, name)
configPath := filepath.Join(configDir, "config.yaml")

// Load config
cfg, err := config.LoadFrom(configPath)
```

### With Logging Module

```go
identity, _ := appidentity.Get(ctx)

// Use identity for service name
serviceName := identity.ServiceName()  // "myvendor.myapp"

logger := logging.New(serviceName,
    logging.WithProfile(logging.ProfileStructured))

logger.Info("Application started",
    logging.WithField("binary", identity.BinaryName),
    logging.WithField("version", "1.0.0"))
```

### With CLI Flags

```go
identity, _ := appidentity.Get(ctx)

// Construct flag names
flagPrefix := identity.FlagsPrefix()  // "myapp-"

configFlag := flag.String(flagPrefix+"config", "", "Config file path")
verboseFlag := flag.Bool(flagPrefix+"verbose", false, "Verbose output")
```

## Error Handling

### Not Found Error

```go
identity, err := appidentity.Get(ctx)
if err != nil {
    if notFound, ok := err.(*appidentity.NotFoundError); ok {
        log.Printf("Identity file not found. Searched:")
        for _, path := range notFound.SearchedPaths {
            log.Printf("  - %s", path)
        }
        log.Printf("See: https://docs.fulmen.dev/app-identity")
    }
}
```

### Validation Error

```go
identity, err := appidentity.Get(ctx)
if err != nil {
    if valErr, ok := err.(*appidentity.ValidationError); ok {
        log.Printf("Validation failed:")
        for _, detail := range valErr.Details {
            log.Printf("  %s: %s", detail.Field, detail.Message)
        }
    }
}
```

### Malformed Error

```go
identity, err := appidentity.Get(ctx)
if err != nil {
    if malformed, ok := err.(*appidentity.MalformedError); ok {
        log.Printf("YAML parsing failed in %s: %v", malformed.Path, malformed.Err)
    }
}
```

## API Reference

### Core Functions

- `Get(ctx) (*Identity, error)` - Load identity with standard discovery
- `GetWithOptions(ctx, Options) (*Identity, error)` - Load with custom options
- `LoadFrom(ctx, path) (*Identity, error)` - Load from explicit path (no caching)
- `Must(ctx) *Identity` - Load or panic (for initialization code)
- `Reset()` - Clear process-level cache (testing only)

### Context Functions

- `WithIdentity(ctx, *Identity) context.Context` - Inject identity for testing
- `FromContext(ctx) (*Identity, bool)` - Extract identity from context

### Testing Utilities

- `NewFixture(opts ...func(*Identity)) *Identity` - Create minimal test fixture
- `NewCompleteFixture() *Identity` - Create complete test fixture
- `Validate(*Identity) error` - Validate against schema

### Identity Methods

- `ConfigParams() (vendor, name string)` - Get config path components
- `EnvVar(suffix string) string` - Construct environment variable name
- `FlagsPrefix() string` - Get CLI flag prefix
- `TelemetryNamespace() string` - Get telemetry namespace
- `ServiceName() string` - Get service name for logging
- `Binary() string` - Get binary name

## Test Coverage

- **Coverage**: 88.4% (68 tests passing)
- **Race Detector**: Clean (no race conditions)
- **Fixtures**: 6 YAML test fixtures (valid/invalid scenarios)
- **Examples**: 8 godoc examples

## Files

- `identity.go` - Core Identity structs and methods
- `loader.go` - File discovery and YAML loading
- `validation.go` - Schema validation
- `cache.go` - Thread-safe process-level caching
- `override.go` - Context-based injection
- `testing.go` - Test utilities and fixtures
- `errors.go` - Typed error types
- `doc.go` - Package documentation

## Dependencies

- `gopkg.in/yaml.v3` - YAML parsing
- Standard library only (no Fulmen dependencies)

## References

- **Crucible Schema**: `app-identity.schema.json` (embedded, v1.0.0)
- **Test Fixtures**: `testdata/*.yaml` (examples of correct structure)
- **Gofulmen Integration**: See `docs/gofulmen_overview.md`

---

**Layer 0 Module**: No Fulmen dependencies, suitable for use in any context without import cycles.
