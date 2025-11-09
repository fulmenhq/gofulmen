package signals

import (
	"context"
	"errors"
	"os"
	"time"
)

// SignalInjector provides test utilities for injecting signals into a Manager
// without relying on OS signal delivery. This allows testing Listen() and
// signal dispatch logic in a controlled, parallel-safe manner.
//
// Example:
//
//	manager := signals.NewManager()
//	injector := signals.NewInjector(manager)
//
//	// Start Listen in background
//	go manager.Listen(ctx)
//
//	// Wait for listener to be ready
//	injector.WaitForListen(time.Second)
//
//	// Inject test signal
//	injector.Inject(syscall.SIGTERM)
type SignalInjector struct {
	manager *Manager
}

// NewInjector creates a new SignalInjector for the given Manager.
func NewInjector(m *Manager) *SignalInjector {
	return &SignalInjector{
		manager: m,
	}
}

// Inject sends a signal directly to the manager's signal channel,
// bypassing OS signal delivery. This triggers the normal signal
// handling logic (handlers, cleanup chains, etc.) as if the signal
// came from the OS.
//
// Returns an error if the manager is not running or if injection fails.
func (i *SignalInjector) Inject(sig os.Signal) error {
	i.manager.mu.RLock()
	running := i.manager.running
	i.manager.mu.RUnlock()

	if !running {
		return errors.New("manager not running - call Listen() first")
	}

	select {
	case i.manager.signalChan <- sig:
		return nil
	case <-time.After(100 * time.Millisecond):
		return errors.New("signal injection timed out - channel may be blocked")
	}
}

// WaitForListen waits for the manager's Listen() method to start and be
// ready to receive signals. This is useful in tests to avoid race conditions
// between starting Listen() in a goroutine and injecting signals.
//
// Returns an error if the timeout expires before Listen() is ready.
func (i *SignalInjector) WaitForListen(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Millisecond)
	defer ticker.Stop()

	for {
		i.manager.mu.RLock()
		running := i.manager.running
		i.manager.mu.RUnlock()

		if running {
			return nil
		}

		if time.Now().After(deadline) {
			return errors.New("timeout waiting for Listen() to start")
		}

		select {
		case <-ticker.C:
			// Continue waiting
		case <-i.manager.stopChan:
			return errors.New("manager stopped before Listen() started")
		}
	}
}

// IsRunning returns true if the manager's Listen() method is currently running.
func (i *SignalInjector) IsRunning() bool {
	i.manager.mu.RLock()
	defer i.manager.mu.RUnlock()
	return i.manager.running
}

// InjectAsync injects a signal asynchronously without blocking. This is useful
// when you want to inject a signal and immediately continue test execution.
// The signal will be processed by Listen() when it's ready.
func (i *SignalInjector) InjectAsync(sig os.Signal) {
	go func() {
		select {
		case i.manager.signalChan <- sig:
		case <-time.After(100 * time.Millisecond):
			// Signal dropped - manager may have stopped
		}
	}()
}

// GetManager returns the underlying Manager for direct access in tests.
// This allows tests to inspect manager state or call methods directly.
func (i *SignalInjector) GetManager() *Manager {
	return i.manager
}

// StopAfter stops the manager after the specified duration. This is useful
// for tests that need to cleanly terminate Listen() after a series of
// signal injections.
func (i *SignalInjector) StopAfter(d time.Duration) {
	time.AfterFunc(d, func() {
		i.manager.Stop()
	})
}

// InjectWithContext injects a signal with a context for timeout control.
// Returns an error if the context expires before injection completes.
func (i *SignalInjector) InjectWithContext(ctx context.Context, sig os.Signal) error {
	i.manager.mu.RLock()
	running := i.manager.running
	i.manager.mu.RUnlock()

	if !running {
		return errors.New("manager not running - call Listen() first")
	}

	select {
	case i.manager.signalChan <- sig:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
