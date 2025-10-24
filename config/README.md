# Config Library

Gofulmen's `config` package provides XDG-compliant configuration management with flexible, application-agnostic path resolution.

## Purpose

The config library addresses common configuration needs in Go applications:

- **XDG Base Directory Support**: Follows XDG standards for config file locations
- **Application-Agnostic**: Use with any app name, not just Fulmen ecosystem
- **File Discovery**: Automatic discovery of configuration files in standard locations
- **Legacy Support**: Backward compatibility with old config locations
- **Ecosystem Conventions**: Optional Fulmen ecosystem defaults
- **Structured Error Reporting**: Error envelopes with correlation ID tracking
- **Telemetry Integration**: Automatic metrics emission for config operations

## Key Features

- **XDG Compliance**: Proper handling of XDG_CONFIG_HOME, XDG_DATA_HOME, XDG_CACHE_HOME
- **Parameterized Paths**: `GetAppConfigDir("myapp")` for any application
- **Config Discovery**: Searches multiple standard locations for config files
- **Fulmen Defaults**: Convenience functions for Fulmen ecosystem tools
- **No Hard-Coded Names**: Library doesn't force "gofulmen" or "fulmen" on consumers
- **Error Envelopes**: Structured error reporting with detailed context
- **Telemetry Metrics**: Automatic emission of config load duration and errors

## Basic Usage

### Loading Configuration

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/config"
)

func main() {
    // Load configuration from default locations
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    fmt.Println("Configuration loaded successfully")
}
```

### Getting XDG Directories

```go
package main

import (
    "fmt"

    "github.com/fulmenhq/gofulmen/config"
)

func main() {
    // Get XDG base directories
    xdg := config.GetXDGBaseDirs()

    fmt.Printf("Config home: %s\n", xdg.ConfigHome)
    fmt.Printf("Data home: %s\n", xdg.DataHome)
    fmt.Printf("Cache home: %s\n", xdg.CacheHome)

    // Get gofulmen-specific directories
    fmt.Printf("Gofulmen config: %s\n", config.GetGofulmenConfigDir())
    fmt.Printf("Gofulmen data: %s\n", config.GetGofulmenDataDir())
    fmt.Printf("Gofulmen cache: %s\n", config.GetGofulmenCacheDir())
}
```

### Config File Discovery

```go
package main

import (
    "fmt"

    "github.com/fulmenhq/gofulmen/config"
)

func main() {
    // Get possible config file paths
    paths := config.GetConfigPaths()

    fmt.Println("Looking for config in:")
    for _, path := range paths {
        fmt.Printf("  - %s\n", path)
    }
}
```

## API Reference

### config.LoadConfig() (\*Config, error)

Loads configuration from default locations following XDG standards.

**Returns:**

- `*Config`: Configuration instance (currently empty struct)
- `error`: Any error during loading

### config.GetXDGBaseDirs() XDGBaseDirs

Returns the XDG Base Directory paths for the current user.

**Returns:**

- `XDGBaseDirs`: Struct with ConfigHome, DataHome, CacheHome paths

### config.GetConfigPaths() []string

Returns a list of possible configuration file paths in order of precedence.

**Returns:**

- `[]string`: Slice of absolute paths where config files are searched

### config.GetGofulmenConfigDir() string

Returns the gofulmen-specific configuration directory.

**Returns:**

- `string`: Path to ~/.config/gofulmen (or equivalent)

## Configuration File Locations

The library searches for configuration files in the following order:

1. `$XDG_CONFIG_HOME/gofulmen/config.json`
2. `~/.config/gofulmen/config.json`
3. `~/.gofulmen.json`
4. `./gofulmen.json`
5. `./.gofulmen.json`

## Testing

```bash
go test ./config/...
```

## Three-Layer Configuration (Preview)

Use `LoadLayeredConfig` to merge Crucible defaults, user overrides, and runtime overrides while validating against the synced schema catalog:

```go
opts := config.LayeredConfigOptions{
    Category:     "sample",
    Version:      "v1.0.0",
    DefaultsFile: "sample-defaults.yaml",
    SchemaID:     "sample/v1.0.0/schema",
}

overrides := map[string]any{
    "settings": map[string]any{
        "retries": 5,
    },
}

merged, diagnostics, err := config.LoadLayeredConfig(opts, overrides)
if err != nil {
    log.Fatal(err)
}

for _, d := range diagnostics {
    fmt.Printf("%s: %s\n", d.Pointer, d.Message)
}

fmt.Printf("Retries => %v\n", merged["settings"].(map[string]any)["retries"])

envOverrides, err := config.LoadEnvOverrides([]config.EnvVarSpec{
    {Name: "APP_RETRIES", Path: []string{"settings", "retries"}, Type: config.EnvInt},
})
if err != nil {
    log.Fatal(err)
}

merged, _, _ = config.LoadLayeredConfig(opts, envOverrides)
fmt.Printf("Retries (env) => %v\n", merged["settings"].(map[string]any)["retries"])
```

Advanced scenarios can set `DefaultsRoot`, `UserPaths`, or pass a custom `*schema.Catalog` through `LayeredConfigOptions`.

## Telemetry and Error Handling

### Structured Error Envelopes

The config package returns structured error envelopes (`*errors.ErrorEnvelope`) for comprehensive error tracking:

```go
import (
    "fmt"

    "github.com/fulmenhq/gofulmen/config"
    "github.com/fulmenhq/gofulmen/errors"
)

func example() {
    opts := config.LayeredConfigOptions{
        Category:     "myapp",
        Version:      "v1.0.0",
        DefaultsFile: "defaults.yaml",
        SchemaID:     "myapp/v1.0.0/schema",
    }

    merged, diags, err := config.LoadLayeredConfigWithEnvelope(opts, "req-123")
    if err != nil {
        if envelope, ok := err.(*errors.ErrorEnvelope); ok {
            fmt.Printf("Error Code: %s\n", envelope.Code)
            fmt.Printf("Severity: %s\n", envelope.Severity)
            fmt.Printf("Correlation ID: %s\n", envelope.CorrelationID)
            fmt.Printf("Context: %+v\n", envelope.Context)
        }
        return err
    }
}
```

### Metrics Emission

Config operations automatically emit telemetry metrics:

- `config_load_ms`: Histogram of config loading duration (tagged by category, version, status)
- `config_load_errors`: Counter of config loading errors (tagged by category, version, error_type, error_code)

All metrics include relevant tags for filtering and aggregation.

### Envelope Variants

Functions provide both simple and envelope variants:

```go
// Simple variants (backward compatible)
merged, diags, err := config.LoadLayeredConfig(opts)
overrides, err := config.LoadEnvOverrides(specs)
xdg := config.GetXDGBaseDirs()

// Envelope variants with structured errors
merged, diags, err := config.LoadLayeredConfigWithEnvelope(opts, correlationID)
overrides, err := config.LoadEnvOverridesWithEnvelope(specs, correlationID)
xdg, err := config.GetXDGBaseDirsWithEnvelope(correlationID)
```

## Future Enhancements

- Schema validation for configuration files
- Hot reloading of configuration
- Support for YAML and TOML formats
- Environment variable prefix support
- Configuration profiles
