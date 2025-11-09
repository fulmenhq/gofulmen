package signals

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogUnsupportedSignal(t *testing.T) {
	// Test with SIGHUP (has fallback metadata on Windows)
	cancel, err := logUnsupportedSignal(syscall.SIGHUP)

	assert.NoError(t, err, "Should not return error for unsupported signal")
	assert.NotNil(t, cancel, "Should return cancel function")

	// Cancel should be no-op
	cancel() // Should not panic
}

func TestLogUnsupportedSignal_UnknownSignal(t *testing.T) {
	// Create a signal that won't be in the catalog
	// Use a syscall signal with high number
	cancel, err := logUnsupportedSignal(syscall.Signal(999))

	assert.NoError(t, err, "Should not return error even for unknown signal")
	assert.NotNil(t, cancel, "Should return cancel function")
}

func TestIsWindows(t *testing.T) {
	// This test verifies the function exists and returns a boolean
	result := IsWindows()
	assert.IsType(t, false, result, "IsWindows should return a boolean")
}

func TestSupportsEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		signal os.Signal
		check  func(bool) bool
	}{
		{
			name:   "SIGTERM always supported",
			signal: syscall.SIGTERM,
			check:  func(supported bool) bool { return supported },
		},
		{
			name:   "SIGINT always supported",
			signal: syscall.SIGINT,
			check:  func(supported bool) bool { return supported },
		},
		{
			name:   "SIGQUIT always supported",
			signal: syscall.SIGQUIT,
			check:  func(supported bool) bool { return supported },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Supports(tt.signal)
			assert.True(t, tt.check(result), "Support check failed for %s", tt.name)
		})
	}
}

func TestPackageLevelConvenience(t *testing.T) {
	// Test package-level convenience functions that delegate to default manager

	// OnShutdown
	OnShutdown(func(ctx context.Context) error {
		return nil
	})

	// OnReload
	OnReload(func(ctx context.Context) error {
		return nil
	})

	// SetQuietMode
	SetQuietMode(true)
	SetQuietMode(false)

	// EnableDoubleTap
	err := EnableDoubleTap(DoubleTapConfig{
		Window:   1 * time.Second,
		Message:  "test",
		ExitCode: 130,
	})
	assert.NoError(t, err, "EnableDoubleTap should not error")

	// Version
	version, err := Version()
	assert.NoError(t, err, "Version should not error")
	assert.NotEmpty(t, version, "Version should return non-empty string")
}

func TestHandle_PackageLevel(t *testing.T) {
	// Test package-level Handle convenience function
	cancel, err := Handle(syscall.SIGTERM, func(ctx context.Context, sig os.Signal) error {
		return nil
	})

	assert.NoError(t, err, "Package-level Handle should not error")
	assert.NotNil(t, cancel, "Should return cancel function")

	cancel() // Cleanup
}
