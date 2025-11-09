package signals

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	fsignals "github.com/fulmenhq/gofulmen/foundry/signals"
)

// HandlerFunc is a function that handles a signal.
type HandlerFunc func(ctx context.Context, sig os.Signal) error

// CleanupFunc is a function that performs cleanup during shutdown.
type CleanupFunc func(ctx context.Context) error

// ReloadFunc is a function that handles config reload with validation.
type ReloadFunc func(ctx context.Context) error

// CancelFunc cancels a signal registration.
type CancelFunc func()

// Manager manages signal handlers and cleanup chains.
type Manager struct {
	mu               sync.RWMutex
	handlers         map[os.Signal][]HandlerFunc
	shutdownHandlers []CleanupFunc
	reloadHandlers   []ReloadFunc
	doubleTapConfig  *DoubleTapConfig
	doubleTapTimer   *time.Timer
	doubleTapActive  bool
	catalog          *fsignals.Catalog
	signalChan       chan os.Signal
	stopChan         chan struct{}
	running          bool
	quietMode        bool
}

// DoubleTapConfig configures Ctrl+C double-tap behavior.
type DoubleTapConfig struct {
	// Window is the time window for detecting a second Ctrl+C.
	// Default: 2 seconds (from catalog)
	Window time.Duration

	// Message to display after first Ctrl+C.
	// Default: from catalog
	Message string

	// ExitCode to use when force quitting.
	// Default: 130 (from catalog)
	ExitCode int
}

// NewManager creates a new signal manager.
func NewManager() *Manager {
	return &Manager{
		handlers:         make(map[os.Signal][]HandlerFunc),
		shutdownHandlers: make([]CleanupFunc, 0),
		reloadHandlers:   make([]ReloadFunc, 0),
		catalog:          fsignals.GetDefaultCatalog(),
		signalChan:       make(chan os.Signal, 1),
		stopChan:         make(chan struct{}),
	}
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

// GetDefaultManager returns the default singleton manager.
func GetDefaultManager() *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewManager()
	})
	return defaultManager
}

// Handle registers a handler for a specific signal.
//
// Returns a CancelFunc that can be called to unregister the handler.
//
// Example:
//
//	cancel, err := signals.Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
//	    log.Println("Received SIGTERM")
//	    return nil
//	})
//	defer cancel()
func Handle(sig os.Signal, handler HandlerFunc) (CancelFunc, error) {
	return GetDefaultManager().Handle(sig, handler)
}

// Handle registers a handler for a specific signal on this manager.
func (m *Manager) Handle(sig os.Signal, handler HandlerFunc) (CancelFunc, error) {
	if !Supports(sig) {
		// Log warning and emit telemetry for unsupported signals
		return logUnsupportedSignal(sig)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[sig] = append(m.handlers[sig], handler)
	idx := len(m.handlers[sig]) - 1

	// Return cancel function
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if handlers, exists := m.handlers[sig]; exists && idx < len(handlers) {
			m.handlers[sig] = append(handlers[:idx], handlers[idx+1:]...)
		}
	}, nil
}

// OnShutdown registers a cleanup function to be called during graceful shutdown.
//
// Cleanup functions are executed in reverse registration order (LIFO).
//
// Example:
//
//	signals.OnShutdown(func(ctx context.Context) error {
//	    return server.Shutdown(ctx)
//	})
func OnShutdown(handler CleanupFunc) {
	GetDefaultManager().OnShutdown(handler)
}

// OnShutdown registers a cleanup function on this manager.
func (m *Manager) OnShutdown(handler CleanupFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownHandlers = append(m.shutdownHandlers, handler)
}

// OnReload registers a config reload handler.
//
// Reload handlers are executed in registration order. If any handler returns
// an error, the reload is aborted and the process continues with the old config.
//
// Example:
//
//	signals.OnReload(func(ctx context.Context) error {
//	    return config.Reload(ctx)
//	})
func OnReload(handler ReloadFunc) {
	GetDefaultManager().OnReload(handler)
}

// OnReload registers a reload handler on this manager.
func (m *Manager) OnReload(handler ReloadFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloadHandlers = append(m.reloadHandlers, handler)
}

// EnableDoubleTap enables Ctrl+C double-tap behavior.
//
// After the first SIGINT, waits for the configured window. If a second SIGINT
// is received within the window, immediately exits with the configured exit code.
//
// Example:
//
//	signals.EnableDoubleTap(signals.DoubleTapConfig{
//	    Window:  2 * time.Second,
//	    Message: "Press Ctrl+C again to force quit",
//	})
func EnableDoubleTap(config DoubleTapConfig) error {
	return GetDefaultManager().EnableDoubleTap(config)
}

// EnableDoubleTap enables double-tap on this manager.
func (m *Manager) EnableDoubleTap(config DoubleTapConfig) error {
	// Load defaults from catalog if not specified
	if config.Window == 0 || config.Message == "" || config.ExitCode == 0 {
		sig, err := m.catalog.GetSignal("int")
		if err != nil {
			return fmt.Errorf("failed to load SIGINT config from catalog: %w", err)
		}

		if config.Window == 0 && sig.DoubleTapWindowSeconds != nil {
			config.Window = time.Duration(*sig.DoubleTapWindowSeconds) * time.Second
		}
		if config.Message == "" {
			config.Message = sig.DoubleTapMessage
		}
		if config.ExitCode == 0 && sig.DoubleTapExitCode != nil {
			config.ExitCode = *sig.DoubleTapExitCode
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.doubleTapConfig = &config
	return nil
}

// SetQuietMode enables or disables quiet mode.
//
// In quiet mode, double-tap messages are not printed to stderr.
// Useful for non-interactive services.
func SetQuietMode(quiet bool) {
	GetDefaultManager().SetQuietMode(quiet)
}

// SetQuietMode sets quiet mode on this manager.
func (m *Manager) SetQuietMode(quiet bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quietMode = quiet
}

// Listen starts listening for signals and blocks until a shutdown signal is received.
//
// By default, registers handlers for SIGTERM and SIGINT. Call Handle() before Listen()
// to register custom handlers.
//
// Example:
//
//	if err := signals.Listen(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
func Listen(ctx context.Context) error {
	return GetDefaultManager().Listen(ctx)
}

// Listen starts listening on this manager.
func (m *Manager) Listen(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("already listening")
	}
	m.running = true
	m.mu.Unlock()

	// Register default signals if no handlers registered
	m.mu.RLock()
	hasHandlers := len(m.handlers) > 0
	m.mu.RUnlock()

	if !hasHandlers {
		// Register default SIGTERM and SIGINT handlers
		signal.Notify(m.signalChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	} else {
		// Register for all signals that have handlers
		m.mu.RLock()
		signals := make([]os.Signal, 0, len(m.handlers))
		for sig := range m.handlers {
			signals = append(signals, sig)
		}
		m.mu.RUnlock()
		signal.Notify(m.signalChan, signals...)
	}

	// Wait for signal or context cancellation
	select {
	case sig := <-m.signalChan:
		return m.handleSignal(ctx, sig)
	case <-ctx.Done():
		return ctx.Err()
	case <-m.stopChan:
		return nil
	}
}

// Stop stops the signal listener.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		signal.Stop(m.signalChan)
		close(m.stopChan)
		m.running = false
	}
}

// handleSignal dispatches a signal to registered handlers.
func (m *Manager) handleSignal(ctx context.Context, sig os.Signal) error {
	// Check for double-tap (SIGINT only)
	if sig == syscall.SIGINT {
		if m.handleDoubleTap() {
			// Force exit
			exitCode := 130
			if m.doubleTapConfig != nil && m.doubleTapConfig.ExitCode != 0 {
				exitCode = m.doubleTapConfig.ExitCode
			}
			os.Exit(exitCode)
		}
	}

	// Execute custom handlers
	m.mu.RLock()
	handlers := m.handlers[sig]
	m.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(ctx, sig); err != nil {
			return fmt.Errorf("signal handler failed: %w", err)
		}
	}

	// Execute shutdown or reload based on signal
	switch sig {
	case syscall.SIGHUP:
		return m.executeReload(ctx)
	case syscall.SIGTERM, syscall.SIGINT:
		return m.executeShutdown(ctx)
	}

	return nil
}

// handleDoubleTap manages double-tap logic for SIGINT.
// Returns true if force exit should occur.
func (m *Manager) handleDoubleTap() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.doubleTapConfig == nil {
		return false
	}

	if m.doubleTapActive {
		// Second tap within window - force exit
		if m.doubleTapTimer != nil {
			m.doubleTapTimer.Stop()
		}
		return true
	}

	// First tap - start timer
	m.doubleTapActive = true
	if !m.quietMode {
		fmt.Fprintln(os.Stderr, m.doubleTapConfig.Message)
	}

	m.doubleTapTimer = time.AfterFunc(m.doubleTapConfig.Window, func() {
		m.mu.Lock()
		m.doubleTapActive = false
		m.mu.Unlock()
	})

	return false
}

// executeShutdown runs all cleanup handlers in reverse order.
func (m *Manager) executeShutdown(ctx context.Context) error {
	m.mu.RLock()
	handlers := make([]CleanupFunc, len(m.shutdownHandlers))
	copy(handlers, m.shutdownHandlers)
	m.mu.RUnlock()

	// Execute in reverse order (LIFO)
	for i := len(handlers) - 1; i >= 0; i-- {
		if err := handlers[i](ctx); err != nil {
			return fmt.Errorf("cleanup handler failed: %w", err)
		}
	}

	return nil
}

// executeReload runs all reload handlers in order.
func (m *Manager) executeReload(ctx context.Context) error {
	m.mu.RLock()
	handlers := make([]ReloadFunc, len(m.reloadHandlers))
	copy(handlers, m.reloadHandlers)
	m.mu.RUnlock()

	// Execute in registration order with fail-fast
	for _, handler := range handlers {
		if err := handler(ctx); err != nil {
			return fmt.Errorf("reload handler failed: %w", err)
		}
	}

	return nil
}

// Supports returns true if the signal is supported on the current platform.
//
// Example:
//
//	if signals.Supports(syscall.SIGHUP) {
//	    // Register HUP handler
//	}
func Supports(sig os.Signal) bool {
	// Get signal name
	sigName := sig.String()

	// Map common signal names to catalog IDs
	idMap := map[string]string{
		"terminated":            "term",
		"interrupt":             "int",
		"hangup":                "hup",
		"quit":                  "quit",
		"broken pipe":           "pipe",
		"alarm clock":           "alrm",
		"user defined signal 1": "usr1",
		"user defined signal 2": "usr2",
	}

	catalog := fsignals.GetDefaultCatalog()
	for name, id := range idMap {
		if sigName == name {
			signal, err := catalog.GetSignal(id)
			if err != nil {
				return false
			}
			// Check if Windows event is defined (nil means unsupported on Windows)
			if IsWindows() && signal.WindowsEvent == nil {
				return false
			}
			return true
		}
	}

	// For syscall signals, try numeric lookup
	if sysSig, ok := sig.(syscall.Signal); ok {
		// Common Unix signals are supported unless on Windows with no event
		commonSignals := map[syscall.Signal]bool{
			syscall.SIGTERM: true,
			syscall.SIGINT:  true,
			syscall.SIGQUIT: true,
		}
		if commonSignals[sysSig] {
			return true
		}

		// SIGHUP, SIGPIPE, SIGALRM not supported on Windows
		if IsWindows() {
			unsupportedOnWindows := map[syscall.Signal]bool{
				syscall.SIGHUP:  true,
				syscall.SIGPIPE: true,
			}
			if unsupportedOnWindows[sysSig] {
				return false
			}
		}
		return true
	}

	// Unknown signal - assume unsupported
	return false
}

// Version returns the signal catalog version.
func Version() (string, error) {
	catalog := fsignals.GetDefaultCatalog()
	return catalog.Version()
}
