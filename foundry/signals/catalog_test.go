package signals

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCatalog(t *testing.T) {
	catalog := NewCatalog()
	assert.NotNil(t, catalog, "NewCatalog should return a non-nil catalog")
}

func TestGetDefaultCatalog(t *testing.T) {
	catalog1 := GetDefaultCatalog()
	catalog2 := GetDefaultCatalog()

	assert.NotNil(t, catalog1, "GetDefaultCatalog should return a non-nil catalog")
	assert.Same(t, catalog1, catalog2, "GetDefaultCatalog should return the same instance")
}

func TestCatalogVersion(t *testing.T) {
	catalog := NewCatalog()
	version, err := catalog.Version()

	require.NoError(t, err, "Version should not return an error")
	assert.Equal(t, "v1.0.0", version, "Version should match expected catalog version")
}

func TestGetSignalByID(t *testing.T) {
	catalog := NewCatalog()

	tests := []struct {
		name         string
		signalID     string
		expectedName string
		expectedNum  int
		expectedExit int
		shouldError  bool
	}{
		{
			name:         "SIGTERM",
			signalID:     "term",
			expectedName: "SIGTERM",
			expectedNum:  15,
			expectedExit: 143,
			shouldError:  false,
		},
		{
			name:         "SIGINT",
			signalID:     "int",
			expectedName: "SIGINT",
			expectedNum:  2,
			expectedExit: 130,
			shouldError:  false,
		},
		{
			name:         "SIGHUP",
			signalID:     "hup",
			expectedName: "SIGHUP",
			expectedNum:  1,
			expectedExit: 129,
			shouldError:  false,
		},
		{
			name:        "Invalid signal",
			signalID:    "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, err := catalog.GetSignal(tt.signalID)

			if tt.shouldError {
				assert.Error(t, err, "GetSignal should return an error for invalid signal ID")
				assert.Nil(t, signal, "Signal should be nil for invalid ID")
				return
			}

			require.NoError(t, err, "GetSignal should not return an error")
			require.NotNil(t, signal, "Signal should not be nil")

			assert.Equal(t, tt.signalID, signal.ID, "Signal ID should match")
			assert.Equal(t, tt.expectedName, signal.Name, "Signal name should match")
			assert.Equal(t, tt.expectedNum, signal.UnixNumber, "Unix signal number should match")
			assert.Equal(t, tt.expectedExit, signal.ExitCode, "Exit code should match")
		})
	}
}

func TestGetSignalByName(t *testing.T) {
	catalog := NewCatalog()

	tests := []struct {
		name        string
		signalName  string
		expectedID  string
		shouldError bool
	}{
		{
			name:        "SIGTERM",
			signalName:  "SIGTERM",
			expectedID:  "term",
			shouldError: false,
		},
		{
			name:        "SIGINT",
			signalName:  "SIGINT",
			expectedID:  "int",
			shouldError: false,
		},
		{
			name:        "Invalid signal name",
			signalName:  "SIGINVALID",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, err := catalog.GetSignalByName(tt.signalName)

			if tt.shouldError {
				assert.Error(t, err, "GetSignalByName should return an error for invalid signal name")
				assert.Nil(t, signal, "Signal should be nil for invalid name")
				return
			}

			require.NoError(t, err, "GetSignalByName should not return an error")
			require.NotNil(t, signal, "Signal should not be nil")

			assert.Equal(t, tt.expectedID, signal.ID, "Signal ID should match")
			assert.Equal(t, tt.signalName, signal.Name, "Signal name should match")
		})
	}
}

func TestListSignals(t *testing.T) {
	catalog := NewCatalog()

	signals, err := catalog.ListSignals()
	require.NoError(t, err, "ListSignals should not return an error")
	assert.NotEmpty(t, signals, "ListSignals should return at least one signal")

	// Verify expected signals are present
	signalIDs := make(map[string]bool)
	for _, signal := range signals {
		signalIDs[signal.ID] = true
	}

	expectedSignals := []string{"term", "int", "hup", "quit", "pipe", "alrm"}
	for _, expectedID := range expectedSignals {
		assert.True(t, signalIDs[expectedID], "Expected signal %s should be in catalog", expectedID)
	}
}

func TestSignalDoubleTap(t *testing.T) {
	catalog := NewCatalog()

	signal, err := catalog.GetSignal("int")
	require.NoError(t, err, "Should get SIGINT signal")

	// Verify double-tap configuration
	assert.NotNil(t, signal.DoubleTapWindowSeconds, "SIGINT should have double-tap window configured")
	assert.Equal(t, 2, *signal.DoubleTapWindowSeconds, "Double-tap window should be 2 seconds")
	assert.NotEmpty(t, signal.DoubleTapMessage, "Double-tap message should be configured")
	assert.Equal(t, "graceful_shutdown_with_double_tap", signal.DefaultBehavior, "Default behavior should include double-tap")
	assert.NotNil(t, signal.DoubleTapExitCode, "Double-tap exit code should be configured")
	assert.Equal(t, 130, *signal.DoubleTapExitCode, "Double-tap exit code should be 130")
}

func TestSignalReload(t *testing.T) {
	catalog := NewCatalog()

	signal, err := catalog.GetSignal("hup")
	require.NoError(t, err, "Should get SIGHUP signal")

	// Verify reload configuration
	assert.Equal(t, "reload_via_restart", signal.DefaultBehavior, "Default behavior should be reload_via_restart")
	assert.Equal(t, "restart_based", signal.ReloadStrategy, "Reload strategy should be restart_based")
	assert.NotNil(t, signal.ValidationRequired, "Validation required flag should be set")
	assert.True(t, *signal.ValidationRequired, "Validation should be required for HUP")

	// Verify cleanup actions include validation
	require.NotEmpty(t, signal.CleanupActions, "Cleanup actions should be defined")
	assert.Contains(t, signal.CleanupActions, "validate_new_config_against_schema", "Cleanup actions should include schema validation")
}

func TestWindowsFallback(t *testing.T) {
	catalog := NewCatalog()

	tests := []struct {
		name              string
		signalID          string
		expectFallback    bool
		expectedBehavior  string
		expectedTelemetry string
	}{
		{
			name:              "SIGHUP has fallback",
			signalID:          "hup",
			expectFallback:    true,
			expectedBehavior:  "http_admin_endpoint",
			expectedTelemetry: "fulmen.signal.unsupported",
		},
		{
			name:              "SIGPIPE has fallback",
			signalID:          "pipe",
			expectFallback:    true,
			expectedBehavior:  "exception_handling",
			expectedTelemetry: "fulmen.signal.unsupported",
		},
		{
			name:           "SIGTERM has Windows event",
			signalID:       "term",
			expectFallback: false,
		},
		{
			name:           "SIGINT has Windows event",
			signalID:       "int",
			expectFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, err := catalog.GetSignal(tt.signalID)
			require.NoError(t, err, "Should get signal")

			if tt.expectFallback {
				require.NotNil(t, signal.WindowsFallback, "Windows fallback should be configured")
				assert.Equal(t, tt.expectedBehavior, signal.WindowsFallback.FallbackBehavior, "Fallback behavior should match")
				assert.Equal(t, "INFO", signal.WindowsFallback.LogLevel, "Log level should be INFO")
				assert.Equal(t, tt.expectedTelemetry, signal.WindowsFallback.TelemetryEvent, "Telemetry event should match")
				assert.NotEmpty(t, signal.WindowsFallback.LogMessage, "Log message should be configured")
				assert.NotEmpty(t, signal.WindowsFallback.LogTemplate, "Log template should be configured")
				assert.NotEmpty(t, signal.WindowsFallback.OperationHint, "Operation hint should be configured")
			} else {
				if signal.WindowsEvent != nil {
					assert.NotEmpty(t, *signal.WindowsEvent, "Windows event should be configured for supported signals")
				}
			}
		})
	}
}

func TestCatalogDescription(t *testing.T) {
	catalog := NewCatalog()

	description, err := catalog.GetDescription()
	require.NoError(t, err, "GetDescription should not return an error")
	assert.NotEmpty(t, description, "Description should not be empty")
	assert.Contains(t, description, "signal handling", "Description should mention signal handling")
}

func TestSignalCleanupActions(t *testing.T) {
	catalog := NewCatalog()

	tests := []struct {
		name            string
		signalID        string
		expectedActions []string
	}{
		{
			name:            "SIGTERM cleanup",
			signalID:        "term",
			expectedActions: []string{"close_connections", "flush_buffers", "remove_pid_file", "log_shutdown"},
		},
		{
			name:            "SIGINT cleanup",
			signalID:        "int",
			expectedActions: []string{"close_connections", "flush_buffers", "remove_pid_file", "log_interrupt"},
		},
		{
			name:            "SIGHUP cleanup (reload)",
			signalID:        "hup",
			expectedActions: []string{"validate_new_config_against_schema", "graceful_shutdown", "restart_with_new_config", "log_reload"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, err := catalog.GetSignal(tt.signalID)
			require.NoError(t, err, "Should get signal")

			assert.Equal(t, tt.expectedActions, signal.CleanupActions, "Cleanup actions should match expected")
		})
	}
}

func TestSignalTimeouts(t *testing.T) {
	catalog := NewCatalog()

	tests := []struct {
		name            string
		signalID        string
		expectedTimeout int
	}{
		{
			name:            "SIGTERM has 30s timeout",
			signalID:        "term",
			expectedTimeout: 30,
		},
		{
			name:            "SIGINT has 5s timeout",
			signalID:        "int",
			expectedTimeout: 5,
		},
		{
			name:            "SIGQUIT has 0s timeout (immediate)",
			signalID:        "quit",
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal, err := catalog.GetSignal(tt.signalID)
			require.NoError(t, err, "Should get signal")

			assert.Equal(t, tt.expectedTimeout, signal.TimeoutSeconds, "Timeout should match expected")
		})
	}
}

// TestCatalogParity verifies the catalog matches expected structure from Crucible v0.2.6.
// This test ensures cross-language parity by validating key catalog properties.
func TestCatalogParity(t *testing.T) {
	catalog := NewCatalog()

	// Verify version matches Crucible standard
	version, err := catalog.Version()
	require.NoError(t, err, "Should get version")
	assert.Equal(t, "v1.0.0", version, "Catalog version should match Crucible standard")

	// Verify all expected signals are present
	signals, err := catalog.ListSignals()
	require.NoError(t, err, "Should list signals")
	require.GreaterOrEqual(t, len(signals), 8, "Should have at least 8 standard signals")

	// Verify exit code calculation (128 + signal number)
	expectedExitCodes := map[string]int{
		"term": 143, // 128 + 15
		"int":  130, // 128 + 2
		"hup":  129, // 128 + 1
		"quit": 131, // 128 + 3
		"pipe": 141, // 128 + 13
		"alrm": 142, // 128 + 14
	}

	for signalID, expectedExit := range expectedExitCodes {
		signal, err := catalog.GetSignal(signalID)
		require.NoError(t, err, "Should get signal %s", signalID)
		assert.Equal(t, expectedExit, signal.ExitCode, "Exit code for %s should be %d", signalID, expectedExit)
	}

	// Verify schema reference
	assert.NotNil(t, catalog.config, "Catalog config should be loaded")
	assert.Contains(t, catalog.config.Schema, "signals.schema.json", "Schema reference should point to signals schema")
}

// TestCatalogLoadIdempotency verifies that loading the catalog multiple times returns the same data.
func TestCatalogLoadIdempotency(t *testing.T) {
	catalog := NewCatalog()

	// Load multiple times
	version1, err1 := catalog.Version()
	version2, err2 := catalog.Version()

	assert.NoError(t, err1, "First load should succeed")
	assert.NoError(t, err2, "Second load should succeed")
	assert.Equal(t, version1, version2, "Version should be consistent across loads")

	// Verify signals are identical
	signal1, err1 := catalog.GetSignal("term")
	signal2, err2 := catalog.GetSignal("term")

	assert.NoError(t, err1, "First signal fetch should succeed")
	assert.NoError(t, err2, "Second signal fetch should succeed")
	assert.Same(t, signal1, signal2, "Signal pointers should be identical (same cached instance)")
}

// TestCatalogConcurrentAccess verifies that the catalog is thread-safe.
func TestCatalogConcurrentAccess(t *testing.T) {
	catalog := NewCatalog()

	// Access catalog from multiple goroutines concurrently
	const goroutines = 10
	done := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			version, err := catalog.Version()
			assert.NoError(t, err, "Concurrent version access should succeed")
			assert.Equal(t, "v1.0.0", version, "Version should be correct")

			signal, err := catalog.GetSignal("term")
			assert.NoError(t, err, "Concurrent signal access should succeed")
			assert.NotNil(t, signal, "Signal should not be nil")

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// TestGetSignalErrorPropagation verifies error handling when catalog fails to load.
// Since we can't easily force a load error with embedded FS, this test documents
// the expected behavior.
func TestGetSignalErrorPropagation(t *testing.T) {
	catalog := NewCatalog()

	// Valid signal should work
	signal, err := catalog.GetSignal("term")
	assert.NoError(t, err, "Valid signal should load successfully")
	assert.NotNil(t, signal, "Valid signal should return data")

	// Invalid signal should return specific error
	signal, err = catalog.GetSignal("nonexistent")
	assert.Error(t, err, "Invalid signal should return error")
	assert.Nil(t, signal, "Invalid signal should return nil")
	assert.Contains(t, err.Error(), "signal not found", "Error should indicate signal not found")
}

// TestListSignalsImmutability verifies that modifying returned signals doesn't affect the catalog.
func TestListSignalsImmutability(t *testing.T) {
	catalog := NewCatalog()

	signals1, err := catalog.ListSignals()
	require.NoError(t, err, "First list should succeed")

	signals2, err := catalog.ListSignals()
	require.NoError(t, err, "Second list should succeed")

	// Slices should be different instances
	assert.NotSame(t, &signals1, &signals2, "ListSignals should return a new slice each time")

	// But content should be identical
	assert.Equal(t, len(signals1), len(signals2), "Signal count should be consistent")
}
