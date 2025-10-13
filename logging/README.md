# Logging Library

Gofulmen's `logging` package provides structured logging with severity-based filtering, multiple output sinks, and schema-validated configuration. Built on zap for performance, this package implements the Fulmen logging standard with first-class support for observability and compliance.

## Purpose

The logging library addresses structured logging needs for Go applications:

- **Structured Logging**: JSON-formatted logs with consistent schema
- **Severity-Based Filtering**: Numeric severity levels with comparison operators
- **Multiple Sinks**: Console (stderr), file, and extensible output targets
- **Schema Validation**: Config validated against crucible logging schemas
- **Performance**: Built on uber/zap for high-performance logging
- **Observability**: Designed for log aggregation and analysis

## Key Features

- **7 Severity Levels**: TRACE (0), DEBUG (10), INFO (20), WARN (30), ERROR (40), FATAL (50), NONE (60)
- **Numeric Comparison**: Severity filters support GE, LE, GT, LT, EQ, NE operators
- **Console Sink**: Stderr-only output (stdout forbidden per standard)
- **File Sink**: Rotating file output with lumberjack
- **Context Enrichment**: WithFields, WithError, WithComponent helpers
- **Static Fields**: Service, environment, version baked into all logs
- **Dynamic Level Changes**: Runtime severity adjustment
- **Schema Validated**: Config validated against crucible observability schemas

## Basic Usage

### Simple Logger

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // Create default logger
    config := logging.DefaultConfig("my-service")
    logger, err := logging.New(config)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Log at different severities
    logger.Info("Application started")
    logger.Debug("Debug information")
    logger.Warn("Warning message")
    logger.Error("Error occurred")
}
```

### CLI-Friendly Logger

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // Create CLI logger with human-readable console output
    logger := logging.NewCLI("my-cli", logging.INFO)
    defer logger.Sync()

    logger.Info("CLI operation started")
    logger.WithFields(map[string]any{
        "file": "config.yaml",
        "lines": 42,
    }).Info("Config loaded")
}
```

### Loading Config from File

```go
package main

import (
    "log"
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    // Load and validate config from file
    config, err := logging.LoadConfig("logger-config.yaml")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    logger, err := logging.New(config)
    if err != nil {
        log.Fatalf("Failed to create logger: %v", err)
    }
    defer logger.Sync()

    logger.Info("Logger configured from file")
}
```

### Context Enrichment

```go
package main

import (
    "errors"
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    logger := logging.NewCLI("my-app", logging.INFO)
    defer logger.Sync()

    // Add fields to log entry
    logger.WithFields(map[string]any{
        "userId": "user-123",
        "action": "login",
    }).Info("User action")

    // Log errors with context
    err := errors.New("database connection failed")
    logger.WithError(err).Error("Database error")

    // Add component context
    logger.WithComponent("pathfinder").Info("Path discovery started")
}
```

### Multiple Sinks

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    config := &logging.LoggerConfig{
        DefaultLevel: "INFO",
        Service:      "my-service",
        Environment:  "production",
        Sinks: []logging.SinkConfig{
            {
                Type:   "console",
                Format: "json",
                Console: &logging.ConsoleSinkConfig{
                    Stream:   "stderr",
                    Colorize: false,
                },
            },
            {
                Type:   "file",
                Level:  "WARN", // Only WARN+ to file
                Format: "json",
                File: &logging.FileSinkConfig{
                    Path:       "/var/log/myapp.log",
                    MaxSize:    100, // MB
                    MaxAge:     30,  // days
                    MaxBackups: 5,
                    Compress:   true,
                },
            },
        },
    }

    logger, err := logging.New(config)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Logged to console only")
    logger.Error("Logged to both console and file")
}
```

### Dynamic Level Changes

```go
package main

import (
    "github.com/fulmenhq/gofulmen/logging"
)

func main() {
    logger := logging.NewCLI("my-app", logging.INFO)
    defer logger.Sync()

    logger.Debug("Not visible at INFO level")

    // Enable debug logging
    logger.SetLevel(logging.DEBUG)
    logger.Debug("Now visible at DEBUG level")

    // Disable logging
    logger.SetLevel(logging.NONE)
    logger.Info("Not logged (NONE level)")
}
```

## Configuration File Format

### YAML Example

```yaml
defaultLevel: INFO
service: my-service
environment: production
enableCaller: false
enableStacktrace: true

staticFields:
  version: "1.0.0"
  region: "us-east-1"

sinks:
  - type: console
    format: json
    console:
      stream: stderr
      colorize: false

  - type: file
    level: WARN
    format: json
    file:
      path: /var/log/myapp.log
      maxSize: 100
      maxAge: 30
      maxBackups: 5
      compress: true
```

### JSON Example

```json
{
  "defaultLevel": "INFO",
  "service": "my-service",
  "environment": "production",
  "enableCaller": false,
  "enableStacktrace": true,
  "staticFields": {
    "version": "1.0.0",
    "region": "us-east-1"
  },
  "sinks": [
    {
      "type": "console",
      "format": "json",
      "console": {
        "stream": "stderr",
        "colorize": false
      }
    }
  ]
}
```

## API Reference

### Logger Creation

#### logging.New(config *LoggerConfig) (*Logger, error)

Creates a new logger from validated configuration.

**Parameters:**

- `config`: Logger configuration (required)

**Returns:**

- `*Logger`: Configured logger instance
- `error`: Validation or initialization error

#### logging.NewCLI(service string, level Severity) \*Logger

Creates a CLI-friendly logger with console output.

**Parameters:**

- `service`: Service name
- `level`: Minimum severity level

**Returns:**

- `*Logger`: CLI logger with human-readable format

#### logging.DefaultConfig(service string) \*LoggerConfig

Returns default logger configuration.

**Parameters:**

- `service`: Service name

**Returns:**

- `*LoggerConfig`: Default configuration with JSON console sink

### Configuration

#### logging.LoadConfig(path string) (\*LoggerConfig, error)

Loads and validates configuration from file (JSON or YAML).

**Parameters:**

- `path`: Path to config file

**Returns:**

- `*LoggerConfig`: Validated configuration
- `error`: Load or validation error

#### logging.ValidateConfig(jsonData []byte) error

Validates logger config against crucible schema.

**Parameters:**

- `jsonData`: JSON-encoded config

**Returns:**

- `error`: Validation error if invalid

### Logging Methods

#### (\*Logger).Info(msg string, keyvals ...any)

Logs at INFO severity.

#### (\*Logger).Debug(msg string, keyvals ...any)

Logs at DEBUG severity.

#### (\*Logger).Warn(msg string, keyvals ...any)

Logs at WARN severity.

#### (\*Logger).Error(msg string, keyvals ...any)

Logs at ERROR severity.

#### (\*Logger).Fatal(msg string, keyvals ...any)

Logs at FATAL severity and exits.

### Context Enrichment

#### (*Logger).WithFields(fields map[string]any) *Logger

Returns logger with additional fields.

**Parameters:**

- `fields`: Key-value pairs to add to log context

**Returns:**

- `*Logger`: Logger with enriched context

#### (*Logger).WithError(err error) *Logger

Returns logger with error field.

**Parameters:**

- `err`: Error to log

**Returns:**

- `*Logger`: Logger with error context

#### (*Logger).WithComponent(component string) *Logger

Returns logger with component field.

**Parameters:**

- `component`: Component name (e.g., "pathfinder", "config")

**Returns:**

- `*Logger`: Logger with component context

### Level Control

#### (\*Logger).SetLevel(level Severity)

Changes minimum severity level dynamically.

**Parameters:**

- `level`: New minimum severity level

#### (\*Logger).Sync() error

Flushes buffered logs to outputs. Call before exit.

**Returns:**

- `error`: Flush error if any

### Severity

#### type Severity

String enum with numeric levels:

- `TRACE` (0): Detailed trace information
- `DEBUG` (10): Debug information
- `INFO` (20): Informational messages
- `WARN` (30): Warning messages
- `ERROR` (40): Error messages
- `FATAL` (50): Fatal errors (exits process)
- `NONE` (60): Disable all logging

#### logging.ParseSeverity(s string) Severity

Parses severity string (defaults to INFO if unknown).

#### (Severity).Level() int

Returns numeric severity level (0-60).

#### (Severity).IsEnabled(minLevel Severity) bool

Checks if severity meets minimum level.

### Severity Filtering

#### logging.MinLevel(level Severity) \*SeverityFilter

Creates filter for level >= threshold (GE operator).

#### logging.MaxLevel(level Severity) \*SeverityFilter

Creates filter for level <= threshold (LE operator).

#### logging.OnlyLevel(level Severity) \*SeverityFilter

Creates filter for exact level match (EQ operator).

### Data Types

#### LoggerConfig

```go
type LoggerConfig struct {
    DefaultLevel     string         // Default severity level
    Service          string         // Service name (required)
    Environment      string         // Environment (dev, prod, etc.)
    Sinks            []SinkConfig   // Output sinks
    StaticFields     map[string]any // Fields added to all logs
    EnableCaller     bool           // Include caller location
    EnableStacktrace bool           // Include stack traces
}
```

#### SinkConfig

```go
type SinkConfig struct {
    Type    string             // "console" or "file"
    Level   string             // Minimum level (optional, inherits default)
    Format  string             // "json", "text", or "console"
    Console *ConsoleSinkConfig // Console config (if type=console)
    File    *FileSinkConfig    // File config (if type=file)
}
```

#### ConsoleSinkConfig

```go
type ConsoleSinkConfig struct {
    Stream   string // Must be "stderr" (stdout forbidden)
    Colorize bool   // Enable color output
}
```

#### FileSinkConfig

```go
type FileSinkConfig struct {
    Path       string // Log file path
    MaxSize    int    // Max size in MB before rotation
    MaxAge     int    // Max age in days
    MaxBackups int    // Max old files to keep
    Compress   bool   // Compress rotated files
}
```

## Log Event Schema

All log events conform to the crucible logging schema:

```json
{
  "severity": "INFO",
  "timestamp": "2025-10-02T13:34:00.979218-04:00",
  "message": "Application started",
  "service": "my-service",
  "environment": "production",
  "component": "pathfinder",
  "userId": "user-123",
  "error": "connection timeout"
}
```

**Required Fields:**

- `severity`: Severity level string (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)
- `timestamp`: RFC3339Nano timestamp
- `message`: Log message (1-32KB)
- `service`: Service name

**Optional Fields:**

- `environment`: Environment name
- `component`: Component/package name
- `error`: Error message
- `tags`: Array of strings (max 50)
- `context`: Object with max 100 fields
- Custom fields via WithFields

## Security & Best Practices

- **Console Output**: Only stderr permitted (stdout reserved for data)
- **Message Limits**: 1-32KB per message (enforced by schema)
- **Tag Limits**: Max 50 tags per event
- **Context Limits**: Max 100 fields in context object
- **No Secrets**: Never log credentials, tokens, or sensitive data
- **Structured Data**: Use fields instead of string interpolation
- **Error Context**: Use WithError for error logging
- **Component Context**: Use WithComponent for package identification
- **Sync on Exit**: Always defer logger.Sync()

## Performance

- Built on uber/zap for high-performance logging
- Zero-allocation in hot paths
- Buffered writes with configurable flushing
- Async file rotation with lumberjack
- Minimal overhead for disabled log levels

## Testing

```bash
go test ./logging/...
```

## Schema Validation

Configuration is validated against crucible observability schemas:

- `schemas/observability/logging/v1.0.0/logger-config.schema.json`
- `schemas/observability/logging/v1.0.0/log-event.schema.json`
- `schemas/observability/logging/v1.0.0/severity-filter.schema.json`

Access schemas via crucible:

```go
logging, _ := crucible.SchemaRegistry.Observability().Logging().V1_0_0()
schemaData, _ := logging.LoggerConfig()
```

## Future Enhancements

- Sampling for high-volume logs
- Log aggregation integration (Loki, CloudWatch)
- OpenTelemetry bridge
- Custom sink plugins
- Correlation ID middleware
- Rate limiting/throttling
