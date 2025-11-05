package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

// TestSignalInjector_Basic tests basic signal injection functionality.
func TestSignalInjector_Basic(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	signalReceived := make(chan os.Signal, 1)
	manager.OnShutdown(func(ctx context.Context) error {
		return nil
	})

	_, _ = manager.Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
		signalReceived <- sig
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start Listen in background
	go func() {
		_ = manager.Listen(ctx)
	}()

	// Wait for Listen to be ready
	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Inject signal
	if err := injector.Inject(syscall.SIGTERM); err != nil {
		t.Fatalf("Inject() failed: %v", err)
	}

	// Verify signal was received
	select {
	case sig := <-signalReceived:
		if sig != syscall.SIGTERM {
			t.Errorf("Expected SIGTERM, got %v", sig)
		}
	case <-time.After(time.Second):
		t.Fatal("Signal not received within timeout")
	}
}

// TestSignalInjector_WaitForListen tests WaitForListen timeout behavior.
func TestSignalInjector_WaitForListen(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	// Should timeout if Listen never starts
	err := injector.WaitForListen(10 * time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
	if err.Error() != "timeout waiting for Listen() to start" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestSignalInjector_InjectBeforeListen tests injecting before Listen starts.
func TestSignalInjector_InjectBeforeListen(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	// Should fail if manager not running
	err := injector.Inject(syscall.SIGTERM)
	if err == nil {
		t.Fatal("Expected error when injecting before Listen(), got nil")
	}
	if err.Error() != "manager not running - call Listen() first" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestSignalInjector_IsRunning tests the IsRunning check.
func TestSignalInjector_IsRunning(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	if injector.IsRunning() {
		t.Error("Manager should not be running initially")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = manager.Listen(ctx)
	}()

	// Wait for Listen to start
	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	if !injector.IsRunning() {
		t.Error("Manager should be running after Listen() starts")
	}

	manager.Stop()

	// Give Stop() time to complete
	time.Sleep(50 * time.Millisecond)

	if injector.IsRunning() {
		t.Error("Manager should not be running after Stop()")
	}
}

// TestSignalInjector_MultipleHandlers tests multiple signal types with different handlers.
func TestSignalInjector_MultipleHandlers(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	usr1Received := make(chan bool, 1)
	usr2Received := make(chan bool, 1)

	// Register handlers for different signals
	_, _ = manager.Handle(syscall.SIGUSR1, func(ctx context.Context, sig os.Signal) error {
		usr1Received <- true
		return nil
	})

	_, _ = manager.Handle(syscall.SIGUSR2, func(ctx context.Context, sig os.Signal) error {
		usr2Received <- true
		return nil
	})

	// Test SIGUSR1
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		_ = manager.Listen(ctx)
	}()

	if err := injector.WaitForListen(500 * time.Millisecond); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	if err := injector.Inject(syscall.SIGUSR1); err != nil {
		t.Fatalf("Inject(SIGUSR1) failed: %v", err)
	}

	select {
	case <-usr1Received:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("SIGUSR1 handler not called")
	}

	// Note: Listen() exits after first signal, so we only test one injection per Listen() session
}

// TestSignalInjector_ShutdownChain tests injecting shutdown signal with cleanup.
func TestSignalInjector_ShutdownChain(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	cleanupOrder := make([]int, 0)
	cleanupChan := make(chan struct{})

	manager.OnShutdown(func(ctx context.Context) error {
		cleanupOrder = append(cleanupOrder, 1)
		return nil
	})

	manager.OnShutdown(func(ctx context.Context) error {
		cleanupOrder = append(cleanupOrder, 2)
		close(cleanupChan)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = manager.Listen(ctx)
	}()

	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Inject SIGTERM to trigger shutdown
	if err := injector.Inject(syscall.SIGTERM); err != nil {
		t.Fatalf("Inject() failed: %v", err)
	}

	// Wait for cleanup to complete
	select {
	case <-cleanupChan:
		// Cleanup completed
	case <-time.After(time.Second):
		t.Fatal("Cleanup not completed within timeout")
	}

	// Verify LIFO order (2, 1)
	if len(cleanupOrder) != 2 {
		t.Fatalf("Expected 2 cleanup calls, got %d", len(cleanupOrder))
	}
	if cleanupOrder[0] != 2 || cleanupOrder[1] != 1 {
		t.Errorf("Expected cleanup order [2, 1], got %v", cleanupOrder)
	}
}

// TestSignalInjector_ReloadChain tests injecting SIGHUP for reload.
func TestSignalInjector_ReloadChain(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	reloadCalled := make(chan bool, 1)

	manager.OnReload(func(ctx context.Context) error {
		reloadCalled <- true
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = manager.Listen(ctx)
	}()

	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Inject SIGHUP to trigger reload
	if err := injector.Inject(syscall.SIGHUP); err != nil {
		t.Fatalf("Inject() failed: %v", err)
	}

	// Verify reload was called
	select {
	case called := <-reloadCalled:
		if !called {
			t.Error("Reload should have been called")
		}
	case <-time.After(time.Second):
		t.Fatal("Reload not called within timeout")
	}
}

// TestSignalInjector_InjectWithContext tests context-based injection.
func TestSignalInjector_InjectWithContext(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	signalReceived := make(chan os.Signal, 1)

	_, _ = manager.Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
		signalReceived <- sig
		return nil
	})

	listenCtx, listenCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer listenCancel()

	go func() {
		_ = manager.Listen(listenCtx)
	}()

	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Test successful injection with context
	injectCtx, injectCancel := context.WithTimeout(context.Background(), time.Second)
	defer injectCancel()

	if err := injector.InjectWithContext(injectCtx, syscall.SIGTERM); err != nil {
		t.Fatalf("InjectWithContext() failed: %v", err)
	}

	select {
	case sig := <-signalReceived:
		if sig != syscall.SIGTERM {
			t.Errorf("Expected SIGTERM, got %v", sig)
		}
	case <-time.After(time.Second):
		t.Fatal("Signal not received within timeout")
	}
}

// TestSignalInjector_InjectAsync tests async injection.
func TestSignalInjector_InjectAsync(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	signalReceived := make(chan os.Signal, 1)

	_, _ = manager.Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
		signalReceived <- sig
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = manager.Listen(ctx)
	}()

	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Inject async
	injector.InjectAsync(syscall.SIGTERM)

	// Verify signal received
	select {
	case sig := <-signalReceived:
		if sig != syscall.SIGTERM {
			t.Errorf("Expected SIGTERM, got %v", sig)
		}
	case <-time.After(time.Second):
		t.Fatal("Signal not received within timeout")
	}
}

// TestSignalInjector_GetManager tests GetManager accessor.
func TestSignalInjector_GetManager(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	if injector.GetManager() != manager {
		t.Error("GetManager() should return the same manager instance")
	}
}

// TestSignalInjector_StopAfter tests delayed stop functionality.
func TestSignalInjector_StopAfter(t *testing.T) {
	manager := NewManager()
	injector := NewInjector(manager)

	ctx := context.Background()

	started := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		close(started)
		if err := manager.Listen(ctx); err != nil && err != context.Canceled {
			t.Errorf("Listen() failed: %v", err)
		}
		close(stopped)
	}()

	<-started

	if err := injector.WaitForListen(time.Second); err != nil {
		t.Fatalf("WaitForListen() failed: %v", err)
	}

	// Schedule stop after 100ms
	injector.StopAfter(100 * time.Millisecond)

	// Wait for Listen to stop
	select {
	case <-stopped:
		// Success - Listen stopped
	case <-time.After(time.Second):
		t.Fatal("Listen did not stop within timeout")
	}
}
