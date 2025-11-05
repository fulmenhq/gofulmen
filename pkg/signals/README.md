# Signals Package

Cross-platform signal handling for graceful shutdown, config reload, and operational control.

## Features

- **Graceful Shutdown**: Register cleanup chains that execute in LIFO order with context support
- **Config Reload**: SIGHUP-triggered reload with validation hooks and restart semantics
- **Ctrl+C Double-Tap**: Configurable window (default 2s) for force quit on stuck processes
- **Windows Fallback**: HTTP admin endpoint for signals unsupported on Windows (SIGHUP, SIGPIPE)
- **Rate Limiting**: Built-in request throttling for HTTP endpoint (default 6/min, burst 3)
- **Thread-Safe**: All APIs are safe for concurrent use

## Signal Catalog

Signal definitions and behaviors come from Crucible catalog (v1.0.0), ensuring cross-language parity with pyfulmen and tsfulmen.

Supported signals:

- **SIGTERM**: Graceful termination (30s timeout)
- **SIGINT**: User interrupt with double-tap (5s timeout, 2s window)
- **SIGHUP**: Config reload via restart (validation required)
- **SIGQUIT**: Immediate termination (0s timeout)
- **SIGUSR1/SIGUSR2**: User-defined signals

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "github.com/fulmenhq/gofulmen/pkg/signals"
)

func main() {
    // Register shutdown cleanup (LIFO order)
    signals.OnShutdown(func(ctx context.Context) error {
        log.Println("Closing database...")
        return db.Close()
    })

    signals.OnShutdown(func(ctx context.Context) error {
        log.Println("Stopping workers...")
        return workers.Stop(ctx)
    })

    // Enable double-tap for force quit
    signals.EnableDoubleTap(signals.DoubleTapConfig{})

    // Start listening
    ctx := context.Background()
    if err := signals.Listen(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Config Reload

```go
// Register reload handler with validation
signals.OnReload(func(ctx context.Context) error {
    // Load and validate new config
    newConfig, err := config.Load()
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // If validation passes, restart with new config
    return restartProcess(ctx, newConfig)
})

// On Unix: kill -HUP <pid>
// On Windows: Use HTTP endpoint (see below)
```

### Windows HTTP Fallback

Some signals (SIGHUP, SIGPIPE) are not supported on Windows. The HTTP admin endpoint provides a fallback:

```go
config := signals.HTTPConfig{
    TokenAuth: os.Getenv("SIGNAL_ADMIN_TOKEN"),
    RateLimit: 6,  // requests per minute
    RateBurst: 3,
}

handler := signals.NewHTTPHandler(config)
http.Handle("/admin/signal", handler)
http.ListenAndServe(":8080", nil)
```

**Trigger reload on Windows:**

```powershell
Invoke-WebRequest -Uri http://localhost:8080/admin/signal `
  -Method POST `
  -Headers @{"Authorization"="Bearer <token>"} `
  -Body '{"signal":"SIGHUP","reason":"config reload"}' `
  -ContentType "application/json"
```

### Advanced Configuration

#### Custom Double-Tap Settings

```go
signals.EnableDoubleTap(signals.DoubleTapConfig{
    Window:   3 * time.Second,
    Message:  "Force quit: press Ctrl+C again",
    ExitCode: 130,
})
```

#### Quiet Mode (Non-Interactive Services)

```go
// Disable stderr messages for non-interactive deployments
signals.SetQuietMode(true)
```

#### mTLS Authentication

```go
config := signals.HTTPConfig{
    MTLSVerify: true,  // Verify client certificates
    RateLimit:  6,
    RateBurst:  3,
}

handler := signals.NewHTTPHandler(config)

// Configure your TLS server to require client certs
server := &http.Server{
    Addr:    ":8443",
    Handler: handler,
    TLSConfig: &tls.Config{
        ClientAuth: tls.RequireAndVerifyClientCert,
        // ... additional cert config
    },
}
server.ListenAndServeTLS("cert.pem", "key.pem")
```

## API Reference

### Registration

```go
// OnShutdown registers a cleanup function (LIFO execution)
func OnShutdown(handler CleanupFunc)

// OnReload registers a config reload handler (FIFO with fail-fast)
func OnReload(handler ReloadFunc)

// Handle registers a handler for a specific signal
func Handle(sig os.Signal, handler HandlerFunc) (CancelFunc, error)
```

### Configuration

```go
// EnableDoubleTap enables Ctrl+C double-tap behavior
func EnableDoubleTap(config DoubleTapConfig) error

// SetQuietMode disables stderr messages
func SetQuietMode(quiet bool)
```

### Lifecycle

```go
// Listen starts listening for signals (blocking)
func Listen(ctx context.Context) error

// Supports checks if a signal is supported on the current platform
func Supports(sig os.Signal) bool

// Version returns the signal catalog version
func Version() (string, error)
```

### HTTP Endpoint

```go
// NewHTTPHandler creates an HTTP handler for /admin/signal
func NewHTTPHandler(config HTTPConfig) *HTTPHandler

// HTTPConfig configures the endpoint
type HTTPConfig struct {
    TokenAuth  string  // Bearer token
    MTLSVerify bool    // Require client certificates
    RateLimit  int     // Requests per minute (default: 6)
    RateBurst  int     // Burst size (default: 3)
    Manager    *Manager // Custom manager (optional)
}

// SignalRequest is the HTTP request format
type SignalRequest struct {
    Signal             string `json:"signal"`                        // Required: "SIGTERM", "SIGHUP", etc.
    Reason             string `json:"reason,omitempty"`              // Optional audit message
    GracePeriodSeconds *int   `json:"grace_period_seconds,omitempty"` // Optional timeout
    Requester          string `json:"requester,omitempty"`           // Optional requester ID
}
```

## Platform Differences

### Unix (Linux, macOS, BSD)

All standard signals are supported via OS signal handlers:

- SIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGUSR1, SIGUSR2

### Windows

Supported via Windows console events:

- SIGTERM (CTRL_CLOSE_EVENT)
- SIGINT (CTRL_C_EVENT)
- SIGQUIT (CTRL_BREAK_EVENT)

**Not supported** (use HTTP endpoint):

- SIGHUP → Logs INFO with hint to use HTTP endpoint
- SIGPIPE → Handled via exception handling

## Error Handling

### Cleanup Chain Failures

If a cleanup handler returns an error, the shutdown process stops and the error is returned:

```go
signals.OnShutdown(func(ctx context.Context) error {
    if err := db.Close(); err != nil {
        return fmt.Errorf("database shutdown failed: %w", err)
    }
    return nil
})
```

### Reload Validation Failures

If a reload handler returns an error, the reload is aborted and the process continues with the old config:

```go
signals.OnReload(func(ctx context.Context) error {
    if err := validateNewConfig(); err != nil {
        // Log error, continue with old config
        return fmt.Errorf("config validation failed: %w", err)
    }
    return restartWithNewConfig()
})
```

## Testing

The package is designed for testing without triggering actual OS signals:

```go
func TestMyHandler(t *testing.T) {
    m := signals.NewManager()

    var called bool
    m.OnShutdown(func(ctx context.Context) error {
        called = true
        return nil
    })

    // Test cleanup execution
    err := m.(*signals.Manager).executeShutdown(context.Background())
    assert.NoError(t, err)
    assert.True(t, called)
}
```

**Note**: Full `Listen()` integration testing requires signal injection utilities (documented in test polish phase).

## Best Practices

1. **Always validate before reload**: SIGHUP handlers should validate new config before restarting
2. **Use context timeouts**: Provide reasonable timeouts for cleanup operations
3. **Log errors**: Capture cleanup/reload failures for debugging
4. **Test cleanup chains**: Unit test shutdown logic without signals
5. **Secure HTTP endpoint**: Always use token auth or mTLS in production
6. **Monitor rate limits**: HTTP endpoint has built-in rate limiting to prevent abuse

## Examples

See `examples_test.go` for additional usage patterns:

- Basic shutdown handling
- Config reload with validation
- HTTP endpoint setup
- Windows fallback behavior
- mTLS configuration

## Related

- [Crucible Signal Handling Spec](https://github.com/fulmenhq/crucible) - Cross-language standard
- [Foundry Signals Catalog](../../foundry/signals/) - Signal metadata accessor
- [Listen() Testing Strategy](../../.plans/active/v0.1.9/listen-testing-strategy.md) - Test architecture

## Version

Signal catalog version: **v1.0.0** (Crucible v0.2.6)
