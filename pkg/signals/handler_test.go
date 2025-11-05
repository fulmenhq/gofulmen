package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m, "NewManager should return a non-nil manager")
	assert.NotNil(t, m.handlers, "Handlers map should be initialized")
	assert.NotNil(t, m.shutdownHandlers, "Shutdown handlers should be initialized")
	assert.NotNil(t, m.reloadHandlers, "Reload handlers should be initialized")
}

func TestGetDefaultManager(t *testing.T) {
	m1 := GetDefaultManager()
	m2 := GetDefaultManager()
	assert.Same(t, m1, m2, "GetDefaultManager should return same instance")
}

func TestHandle(t *testing.T) {
	m := NewManager()

	handler := func(ctx context.Context, sig os.Signal) error {
		return nil
	}

	cancel, err := m.Handle(syscall.SIGTERM, handler)
	require.NoError(t, err, "Handle should not return error for supported signal")
	assert.NotNil(t, cancel, "Cancel function should be returned")

	// Verify handler was registered
	m.mu.RLock()
	assert.Len(t, m.handlers[syscall.SIGTERM], 1, "Should have one handler registered")
	m.mu.RUnlock()

	// Test cancel
	cancel()
	m.mu.RLock()
	assert.Len(t, m.handlers[syscall.SIGTERM], 0, "Handler should be removed after cancel")
	m.mu.RUnlock()
}

func TestOnShutdown(t *testing.T) {
	m := NewManager()

	handler := func(ctx context.Context) error {
		return nil
	}

	m.OnShutdown(handler)

	m.mu.RLock()
	assert.Len(t, m.shutdownHandlers, 1, "Should have one shutdown handler registered")
	m.mu.RUnlock()
}

func TestOnReload(t *testing.T) {
	m := NewManager()

	handler := func(ctx context.Context) error {
		return nil
	}

	m.OnReload(handler)

	m.mu.RLock()
	assert.Len(t, m.reloadHandlers, 1, "Should have one reload handler registered")
	m.mu.RUnlock()
}

func TestEnableDoubleTap(t *testing.T) {
	m := NewManager()

	config := DoubleTapConfig{
		Window:   2 * time.Second,
		Message:  "Test message",
		ExitCode: 130,
	}

	err := m.EnableDoubleTap(config)
	require.NoError(t, err, "EnableDoubleTap should not return error")

	m.mu.RLock()
	assert.NotNil(t, m.doubleTapConfig, "Double-tap config should be set")
	assert.Equal(t, 2*time.Second, m.doubleTapConfig.Window, "Window should match")
	assert.Equal(t, "Test message", m.doubleTapConfig.Message, "Message should match")
	assert.Equal(t, 130, m.doubleTapConfig.ExitCode, "Exit code should match")
	m.mu.RUnlock()
}

func TestEnableDoubleTap_CatalogDefaults(t *testing.T) {
	m := NewManager()

	// Enable with zero values - should load from catalog
	err := m.EnableDoubleTap(DoubleTapConfig{})
	require.NoError(t, err, "EnableDoubleTap should not return error")

	m.mu.RLock()
	assert.NotNil(t, m.doubleTapConfig, "Double-tap config should be set")
	assert.Equal(t, 2*time.Second, m.doubleTapConfig.Window, "Window should be from catalog (2s)")
	assert.NotEmpty(t, m.doubleTapConfig.Message, "Message should be from catalog")
	assert.Equal(t, 130, m.doubleTapConfig.ExitCode, "Exit code should be from catalog (130)")
	m.mu.RUnlock()
}

func TestSetQuietMode(t *testing.T) {
	m := NewManager()

	assert.False(t, m.quietMode, "Quiet mode should be false by default")

	m.SetQuietMode(true)
	assert.True(t, m.quietMode, "Quiet mode should be enabled")

	m.SetQuietMode(false)
	assert.False(t, m.quietMode, "Quiet mode should be disabled")
}

func TestExecuteShutdown(t *testing.T) {
	m := NewManager()

	var order []int

	m.OnShutdown(func(ctx context.Context) error {
		order = append(order, 1)
		return nil
	})

	m.OnShutdown(func(ctx context.Context) error {
		order = append(order, 2)
		return nil
	})

	m.OnShutdown(func(ctx context.Context) error {
		order = append(order, 3)
		return nil
	})

	ctx := context.Background()
	err := m.executeShutdown(ctx)
	require.NoError(t, err, "Shutdown should execute without error")

	// Verify LIFO order (reverse registration)
	assert.Equal(t, []int{3, 2, 1}, order, "Shutdown handlers should execute in LIFO order")
}

func TestExecuteReload(t *testing.T) {
	m := NewManager()

	var order []int

	m.OnReload(func(ctx context.Context) error {
		order = append(order, 1)
		return nil
	})

	m.OnReload(func(ctx context.Context) error {
		order = append(order, 2)
		return nil
	})

	m.OnReload(func(ctx context.Context) error {
		order = append(order, 3)
		return nil
	})

	ctx := context.Background()
	err := m.executeReload(ctx)
	require.NoError(t, err, "Reload should execute without error")

	// Verify FIFO order (registration order)
	assert.Equal(t, []int{1, 2, 3}, order, "Reload handlers should execute in registration order")
}

func TestExecuteReload_FailFast(t *testing.T) {
	m := NewManager()

	var executed []int

	m.OnReload(func(ctx context.Context) error {
		executed = append(executed, 1)
		return nil
	})

	m.OnReload(func(ctx context.Context) error {
		executed = append(executed, 2)
		return assert.AnError // Fail here
	})

	m.OnReload(func(ctx context.Context) error {
		executed = append(executed, 3)
		return nil
	})

	ctx := context.Background()
	err := m.executeReload(ctx)
	require.Error(t, err, "Reload should return error on handler failure")

	// Verify only first two executed (fail-fast)
	assert.Equal(t, []int{1, 2}, executed, "Should stop after first failure")
}

func TestHandleDoubleTap(t *testing.T) {
	m := NewManager()

	config := DoubleTapConfig{
		Window:   100 * time.Millisecond,
		Message:  "Test",
		ExitCode: 130,
	}
	err := m.EnableDoubleTap(config)
	require.NoError(t, err)

	// First tap should not force exit
	shouldExit := m.handleDoubleTap()
	assert.False(t, shouldExit, "First tap should not force exit")

	m.mu.RLock()
	active := m.doubleTapActive
	m.mu.RUnlock()
	assert.True(t, active, "Double-tap should be active after first tap")

	// Second tap within window should force exit
	shouldExit = m.handleDoubleTap()
	assert.True(t, shouldExit, "Second tap within window should force exit")
}

func TestHandleDoubleTap_WindowExpiry(t *testing.T) {
	m := NewManager()

	config := DoubleTapConfig{
		Window:   50 * time.Millisecond,
		Message:  "Test",
		ExitCode: 130,
	}
	err := m.EnableDoubleTap(config)
	require.NoError(t, err)

	// First tap
	shouldExit := m.handleDoubleTap()
	assert.False(t, shouldExit, "First tap should not force exit")

	// Wait for window to expire
	time.Sleep(100 * time.Millisecond)

	m.mu.RLock()
	active := m.doubleTapActive
	m.mu.RUnlock()
	assert.False(t, active, "Double-tap should be inactive after window expiry")

	// New tap after window should not force exit (treated as new first tap)
	shouldExit = m.handleDoubleTap()
	assert.False(t, shouldExit, "Tap after window should be treated as new first tap")
}

func TestHandleDoubleTap_Disabled(t *testing.T) {
	m := NewManager()

	// Double-tap not enabled
	shouldExit := m.handleDoubleTap()
	assert.False(t, shouldExit, "Should not force exit when double-tap disabled")
}

func TestSupports(t *testing.T) {
	tests := []struct {
		name      string
		signal    os.Signal
		supported bool
	}{
		{"SIGTERM supported", syscall.SIGTERM, true},
		{"SIGINT supported", syscall.SIGINT, true},
		{"SIGQUIT supported", syscall.SIGQUIT, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Supports(tt.signal)
			assert.Equal(t, tt.supported, result, "Supports should return %v for %s", tt.supported, tt.name)
		})
	}
}

func TestVersion(t *testing.T) {
	version, err := Version()
	require.NoError(t, err, "Version should not return error")
	assert.Equal(t, "v1.0.0", version, "Version should match catalog version")
}

func TestConcurrentHandlerRegistration(t *testing.T) {
	m := NewManager()

	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			_, err := m.Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
				return nil
			})
			assert.NoError(t, err)
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	m.mu.RLock()
	assert.Len(t, m.handlers[syscall.SIGTERM], goroutines, "Should have all handlers registered")
	m.mu.RUnlock()
}

func TestConcurrentShutdownRegistration(t *testing.T) {
	m := NewManager()

	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			m.OnShutdown(func(ctx context.Context) error {
				return nil
			})
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	m.mu.RLock()
	assert.Len(t, m.shutdownHandlers, goroutines, "Should have all shutdown handlers registered")
	m.mu.RUnlock()
}
