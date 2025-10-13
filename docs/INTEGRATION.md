# Gofulmen Integration Guide

**For Library Consumers**: This guide helps you integrate gofulmen packages into your Go applications.

## What is Gofulmen?

Gofulmen is the **Go foundation library** for the Fulmen ecosystem, providing:

- **Structured logging** - Production-ready zap-based logging with schema validation
- **Safe filesystem ops** - Pathfinder for secure file discovery
- **Config management** - XDG-compliant configuration paths
- **ASCII utilities** - Terminal-aware box drawing and formatting
- **Crucible access** - All schemas, standards, and docs via single import

## Quick Start

### Installation

```bash
go get github.com/fulmenhq/gofulmen
```

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/fulmenhq/gofulmen/config"
    "github.com/fulmenhq/gofulmen/logging"
    "github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
    // Config: Get standard directories
    configDir := config.GetAppConfigDir("myapp")

    // Logging: Create structured logger
    logConfig := logging.DefaultConfig("myapp")
    logger, _ := logging.New(logConfig)
    defer logger.Sync()

    logger.Info("Application started")

    // Pathfinder: Safe file discovery
    ctx := context.Background()
    finder := pathfinder.NewFinder()
    files, _ := finder.FindGoFiles(ctx, ".")

    logger.WithFields(map[string]any{
        "count": len(files),
    }).Info("Files discovered")
}
```

## Available Packages

### Logging (`logging/`)

Production-ready structured logging with severity levels, multiple sinks, and schema validation.

**Features**:

- 7 severity levels (TRACE, DEBUG, INFO, WARN, ERROR, FATAL, NONE)
- Console (stderr) and file sinks with rotation
- Context enrichment (fields, errors, components)
- Schema-validated configuration

**Example**:

```go
import "github.com/fulmenhq/gofulmen/logging"

config := logging.DefaultConfig("my-service")
logger, _ := logging.New(config)
defer logger.Sync()

logger.Info("User logged in")
logger.WithFields(map[string]any{
    "userId": "user-123",
    "ip": "192.168.1.1",
}).Info("Request processed")
```

[📚 Full Logging Documentation](../logging/README.md)

### Pathfinder (`pathfinder/`)

Safe filesystem discovery with security-first design.

**Features**:

- Path validation (prevents traversal attacks)
- Pattern-based discovery (glob support)
- Context-aware operations
- Built-in safety checks

**Example**:

```go
import "github.com/fulmenhq/gofulmen/pathfinder"

finder := pathfinder.NewFinder()
goFiles, _ := finder.FindGoFiles(ctx, ".")

for _, file := range goFiles {
    fmt.Println(file.RelativePath)
}
```

[📚 Full Pathfinder Documentation](../pathfinder/README.md)

### Config (`config/`)

XDG-compliant configuration management.

**Features**:

- XDG Base Directory compliance
- Config file discovery with fallbacks
- Application-agnostic path helpers
- Ecosystem-friendly (Fulmen tools share `~/.config/fulmen/`)

**Example**:

```go
import "github.com/fulmenhq/gofulmen/config"

// Your app's config directory
configDir := config.GetAppConfigDir("myapp")
// Returns: ~/.config/myapp

// Search for config files (with fallbacks)
paths := config.GetAppConfigPaths("myapp", "legacy-name")
// Searches: ~/.config/myapp/, ~/.myapp/, ./myapp.yaml, etc.

// XDG directories
xdg := config.GetXDGBaseDirs()
fmt.Printf("Config: %s\n", xdg.ConfigHome)  // ~/.config
fmt.Printf("Data: %s\n", xdg.DataHome)      // ~/.local/share
fmt.Printf("Cache: %s\n", xdg.CacheHome)    // ~/.cache
```

[📚 Full Config Documentation](../config/README.md)

### ASCII (`ascii/`)

Terminal formatting and Unicode utilities.

**Features**:

- Unicode box drawing
- Text alignment and padding
- Multi-byte character support
- Terminal-specific width handling

**Example**:

```go
import "github.com/fulmenhq/gofulmen/ascii"

lines := []string{
    "Welcome to MyApp",
    "Version 1.0.0",
}
ascii.DrawBox(lines)
```

[📚 Full ASCII Documentation](../ascii/README.md)

### Crucible (`crucible/`)

Unified access to Crucible schemas, standards, and documentation.

**Features**:

- Complete schema registry access
- Embedded documentation
- Version tracking
- Convenience helpers

**Example**:

```go
import "github.com/fulmenhq/gofulmen/crucible"

// Access schemas
eventSchema, _ := crucible.GetLoggingEventSchema()

// Access docs
ecosystemDoc, _ := crucible.GetLibraryEcosystemDoc()

// Version info
fmt.Println(crucible.GetVersionString())
// Output: gofulmen/0.1.0 crucible/2025.10.0
```

[📚 Full Crucible Documentation](../crucible/README.md)

## Integration Patterns

### CLI Application Setup

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/fulmenhq/gofulmen/config"
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // CLI-friendly logger (INFO level, stderr only)
    logger := logging.NewCLI("my-tool", logging.INFO)
    defer logger.Sync()

    // Context with cancellation
    ctx, cancel := signal.NotifyContext(
        context.Background(),
        syscall.SIGINT,
        syscall.SIGTERM,
    )
    defer cancel()

    // Run application
    if err := run(ctx, logger); err != nil {
        logger.WithError(err).Fatal("Command failed")
    }
}

func run(ctx context.Context, logger *logging.Logger) error {
    // Your CLI logic
    return nil
}
```

### Service Application Setup

```go
package main

import (
    "context"
    "os"
    "path/filepath"

    "github.com/fulmenhq/gofulmen/config"
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // Load config from standard location
    xdg := config.GetXDGBaseDirs()
    configPath := filepath.Join(xdg.ConfigHome, "myservice", "logging.yaml")

    logConfig, err := logging.LoadConfig(configPath)
    if err != nil {
        // Fallback to defaults
        logConfig = logging.DefaultConfig("myservice")
    }

    logger, _ := logging.New(logConfig)
    defer logger.Sync()

    logger.Info("Service starting")

    // Your service logic
    if err := runService(context.Background(), logger); err != nil {
        logger.WithError(err).Fatal("Service failed")
    }
}
```

### Library Integration

If you're building a library that uses gofulmen:

```go
package mylib

import "github.com/fulmenhq/gofulmen/logging"

type Client struct {
    logger *logging.Logger
}

// NewClient accepts an optional logger
func NewClient(logger *logging.Logger) *Client {
    if logger == nil {
        // Use no-op logger if none provided
        logger = logging.NewCLI("mylib", logging.NONE)
    }
    return &Client{logger: logger}
}

func (c *Client) DoSomething() error {
    c.logger.Info("Doing something")
    // Your logic
    return nil
}
```

## Configuration Management

### For Fulmen Ecosystem Tools

If you're building a tool for the Fulmen ecosystem:

```go
// Use shared ecosystem location for common configs
sharedDir := config.GetFulmenConfigDir()
// Returns: ~/.config/fulmen

// Example: Terminal calibration (shared across all Fulmen tools)
terminalConfig := filepath.Join(sharedDir, "terminal-overrides.yaml")
```

### For Independent Applications

```go
// Use your own app name
configDir := config.GetAppConfigDir("myapp")
// Returns: ~/.config/myapp

dataDir := config.GetAppDataDir("myapp")
// Returns: ~/.local/share/myapp

cacheDir := config.GetAppCacheDir("myapp")
// Returns: ~/.cache/myapp
```

## Best Practices

1. **Always `defer logger.Sync()`** - Ensures logs are flushed
2. **Use structured fields** - Prefer `WithFields()` over string interpolation
3. **Configure via files** - Use YAML/JSON config for production
4. **Validate paths** - Use `pathfinder.ValidatePath()` for user input
5. **Follow XDG** - Use config package for standard locations
6. **Check versions** - Log `crucible.GetVersionString()` on startup

## Troubleshooting

### Import Errors

**Problem**: Cannot import gofulmen

```go
import "github.com/fulmenhq/gofulmen/logging" // ❌ Module not found
```

**Solution**: Run `go get`

```bash
go get github.com/fulmenhq/gofulmen
```

### Missing Schemas

**Problem**: Schema validation fails

**Solution**: Gofulmen embeds all required schemas. If you see schema errors, ensure you're using the latest version:

```bash
go get -u github.com/fulmenhq/gofulmen
```

### Config File Not Found

**Problem**: Application can't find config file

**Solution**: Use `GetAppConfigPaths()` to search multiple locations:

```go
paths := config.GetAppConfigPaths("myapp", "")
for _, path := range paths {
    if _, err := os.Stat(path); err == nil {
        // Found config at path
        break
    }
}
```

## Examples

See the [examples/](../examples/) directory for complete working examples:

- `basic-pathfinder.go` - Filesystem discovery
- More examples coming soon

## Migration from Other Libraries

### From logrus

```go
// Before (logrus)
log := logrus.New()
log.WithFields(logrus.Fields{"user": "john"}).Info("message")

// After (gofulmen)
logger := logging.NewCLI("myapp", logging.INFO)
logger.WithFields(map[string]any{"user": "john"}).Info("message")
```

### From standard log

```go
// Before (standard log)
log.Printf("User %s logged in", userID)

// After (gofulmen)
logger := logging.NewCLI("myapp", logging.INFO)
logger.WithFields(map[string]any{"userId": userID}).Info("User logged in")
```

## Support

- **Issues**: https://github.com/fulmenhq/gofulmen/issues
- **Discussions**: https://github.com/fulmenhq/gofulmen/discussions
- **Documentation**: Individual package READMEs in repository

## Next Steps

1. Install gofulmen: `go get github.com/fulmenhq/gofulmen`
2. Choose packages you need (logging, pathfinder, config, ascii)
3. Read package-specific READMEs for detailed API docs
4. Start building!

## For Gofulmen Contributors

If you're contributing to gofulmen itself, see:

- [FULDX.md](FULDX.md) - Development experience and tooling
- [ops/bootstrap-strategy.md](../ops/bootstrap-strategy.md) - Bootstrap architecture
