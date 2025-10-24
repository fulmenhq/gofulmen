# Pathfinder Library

Gofulmen's `pathfinder` package provides safe filesystem discovery and traversal operations for Go applications. This is the core implementation that provides a clean, secure API for path operations.

## Purpose

The pathfinder library addresses common filesystem operation challenges in Go applications:

- **Safe Path Handling**: Prevents path traversal attacks and validates paths
- **File Discovery**: Find files matching patterns with built-in safety checks
- **Security First**: All operations include validation and safety measures
- **Simple API**: Easy-to-use interface for common discovery operations

## Key Features

- **Path Validation**: Detect and prevent path traversal attacks (".." sequences)
- **Pattern-based Discovery**: Glob pattern matching for file discovery
- **Security Checks**: Built-in validation for all path operations
- **Extensible Design**: Clean interfaces for future enhancements
- **Context Support**: Proper context handling for cancellable operations

## Basic Usage

### Path Validation

```go
package main

import (
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
    // Validate safe paths
    safePaths := []string{"valid/path", "another/valid"}
    for _, path := range safePaths {
        if err := pathfinder.ValidatePath(path); err != nil {
            log.Printf("Invalid path %s: %v", path, err)
        } else {
            fmt.Printf("Path %s is safe\n", path)
        }
    }

    // These will fail validation
    unsafePaths := []string{"../escape", "", "/"}
    for _, path := range unsafePaths {
        if err := pathfinder.ValidatePath(path); err != nil {
            fmt.Printf("Path %s rejected: %v\n", path, err)
        }
    }
}
```

### File Discovery

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
    ctx := context.Background()
    finder := pathfinder.NewFinder()

    // Find Go files
    goFiles, err := finder.FindGoFiles(ctx, ".")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d Go files:\n", len(goFiles))
    for _, file := range goFiles {
        fmt.Printf("  %s\n", file.RelativePath)
    }

    // Find config files
    configFiles, err := finder.FindConfigFiles(ctx, ".")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d config files:\n", len(configFiles))
    for _, file := range configFiles {
        fmt.Printf("  %s\n", file.RelativePath)
    }
}
```

### Advanced Discovery

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
    ctx := context.Background()
    finder := pathfinder.NewFinder()

    // Custom discovery query
    query := pathfinder.FindQuery{
        Root:          ".",
        Include:       []string{"*.md", "*.txt"},
        Exclude:       []string{"*.tmp"},
        MaxDepth:      3,
        IncludeHidden: false,
        ErrorHandler: func(path string, err error) error {
            fmt.Printf("Error processing %s: %v\n", path, err)
            return nil // Continue processing
        },
    }

    results, err := finder.FindFiles(ctx, query)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d files:\n", len(results))
    for _, result := range results {
        fmt.Printf("  %s -> %s\n", result.RelativePath, result.SourcePath)
    }
}
```

### Checksum Calculation Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fulmenhq/gofulmen/pathfinder"
)

func main() {
    ctx := context.Background()
    finder := pathfinder.NewFinder()

    // Discovery with checksum calculation
    query := pathfinder.FindQuery{
        Root:               ".",
        Include:            []string{"*.go"},
        CalculateChecksums: true,
        ChecksumAlgorithm:  "xxh3-128", // or "sha256"
    }

    results, err := finder.FindFiles(ctx, query)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d Go files with checksums:\n", len(results))
    for _, result := range results {
        checksum := result.Metadata["checksum"]
        algorithm := result.Metadata["checksumAlgorithm"]
        fmt.Printf("  %s: %s (%s)\n", result.RelativePath, checksum, algorithm)
    }
}
```

## API Reference

### Core Functions

#### pathfinder.ValidatePath(path string) error

Validates that a path is safe to use, checking for path traversal attempts and invalid patterns.

**Parameters:**

- `path`: The path string to validate

**Returns:**

- `error`: nil if path is safe, error describing the issue otherwise

#### pathfinder.IsSafePath(path string) bool

Convenience function that returns true if the path passes validation.

**Parameters:**

- `path`: The path string to check

**Returns:**

- `bool`: true if path is safe, false otherwise

### Finder API

#### NewFinder() \*Finder

Creates a new Finder instance with default configuration.

**Returns:**

- `*Finder`: A new finder instance

#### (\*Finder).FindFiles(ctx context.Context, query FindQuery) ([]PathResult, error)

Performs file discovery based on the provided query parameters.

**Parameters:**

- `ctx`: Context for cancellation
- `query`: FindQuery specifying discovery parameters

**Returns:**

- `[]PathResult`: Slice of discovered file results
- `error`: Any error during discovery

#### (\*Finder).FindGoFiles(ctx context.Context, root string) ([]PathResult, error)

Convenience method to find Go source files (\*.go).

**Parameters:**

- `ctx`: Context for cancellation
- `root`: Root directory to search from

**Returns:**

- `[]PathResult`: Slice of discovered Go files
- `error`: Any error during discovery

#### (\*Finder).FindConfigFiles(ctx context.Context, root string) ([]PathResult, error)

Convenience method to find common configuration files.

**Parameters:**

- `ctx`: Context for cancellation
- `root`: Root directory to search from

**Returns:**

- `[]PathResult`: Slice of discovered config files
- `error`: Any error during discovery

#### (\*Finder).FindSchemaFiles(ctx context.Context, root string) ([]PathResult, error)

Convenience method to find JSON Schema files.

**Parameters:**

- `ctx`: Context for cancellation
- `root`: Root directory to search from

**Returns:**

- `[]PathResult`: Slice of discovered schema files
- `error`: Any error during discovery

#### (\*Finder).FindByExtension(ctx context.Context, root string, exts []string) ([]PathResult, error)

Finds files with specific extensions.

**Parameters:**

- `ctx`: Context for cancellation
- `root`: Root directory to search from
- `exts`: Slice of file extensions (without dots)

**Returns:**

- `[]PathResult`: Slice of discovered files
- `error`: Any error during discovery

### Data Types

#### FindQuery

Specifies parameters for file discovery operations.

```go
type FindQuery struct {
    Root               string                                      // Root directory to search from
    Include            []string                                    // Patterns to include (e.g., "*.go")
    Exclude            []string                                    // Patterns to exclude
    MaxDepth           int                                         // Maximum directory depth (0 = unlimited)
    FollowSymlinks     bool                                        // Whether to follow symbolic links
    IncludeHidden      bool                                        // Whether to include hidden files/directories
    CalculateChecksums bool                                        // Whether to calculate file checksums
    ChecksumAlgorithm  string                                      // Checksum algorithm ("xxh3-128" or "sha256", default "xxh3-128")
    ErrorHandler       func(path string, err error) error          // Error handler function
    ProgressCallback   func(processed int, total int, currentPath string) // Progress callback
}
```

#### PathResult

Represents a discovered file or directory.

```go
type PathResult struct {
    RelativePath string            // Path relative to search root
    SourcePath   string            // Absolute path to the file
    LogicalPath  string            // Logical path (defaults to RelativePath)
    LoaderType   string            // Type of loader used ("local")
    Metadata     map[string]any    // Additional metadata (size, mtime, checksum, checksumAlgorithm)
}
```

**Metadata Fields:**

- `size`: File size in bytes (int64)
- `mtime`: File modification time in RFC3339Nano format (string)
- `checksum`: File checksum in "algorithm:hex" format (string, when CalculateChecksums=true)
- `checksumAlgorithm`: Checksum algorithm used ("xxh3-128" or "sha256", when CalculateChecksums=true)
- `checksumError`: Error message if checksum calculation failed (string, optional)

## Security Considerations

- **Path Traversal Protection**: All paths are validated to prevent ".." traversal attacks
- **Input Validation**: All user-provided paths are sanitized
- **Safe Defaults**: Conservative defaults that prioritize security
- **Error Handling**: Comprehensive error reporting for security events

## Testing

```bash
go test ./pathfinder/...
```

## Telemetry and Error Handling

### Structured Error Envelopes

The pathfinder package returns structured error envelopes (`*errors.ErrorEnvelope`) for comprehensive error tracking:

```go
import (
    "context"
    "fmt"

    "github.com/fulmenhq/gofulmen/errors"
    "github.com/fulmenhq/gofulmen/pathfinder"
)

func example() {
    ctx := context.Background()
    finder := pathfinder.NewFinder()

    results, err := finder.FindFiles(ctx, query)
    if err != nil {
        if envelope, ok := err.(*errors.ErrorEnvelope); ok {
            fmt.Printf("Error Code: %s\n", envelope.Code)
            fmt.Printf("Severity: %s\n", envelope.Severity)
            fmt.Printf("Context: %+v\n", envelope.Context)
        }
        return err
    }
}
```

### Metrics Emission

Pathfinder automatically emits telemetry metrics:

- `pathfinder_find_ms`: Histogram of file discovery duration
- `pathfinder_validation_errors`: Counter of validation failures
- `pathfinder_security_warnings`: Counter of security events (path traversal attempts)

All metrics include relevant tags (root, status, error_type) for filtering and aggregation.

### Validation Functions with Envelopes

Validation functions provide both simple and envelope variants:

```go
// Simple validation (backward compatible)
err := pathfinder.ValidatePathResult(result)

// With structured error envelope
err := pathfinder.ValidatePathResultWithEnvelope(result, correlationID)

// Batch validation
err := pathfinder.ValidatePathResultsWithEnvelope(results, correlationID)

// Security validation
err := pathfinder.ValidatePathWithinRootWithEnvelope(absPath, absRoot, correlationID)
```

## Future Enhancements

- Advanced pattern matching with regular expressions
- Directory traversal with configurable safety limits
- Multiple filesystem loader support (remote, cloud storage)
- Caching layer for performance
- Audit logging for compliance
- Schema-based discovery with validation
- Performance optimizations for large directory trees
