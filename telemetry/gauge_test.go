package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGauge verifies gauge metric emission
func TestGauge(t *testing.T) {
	sys, err := NewSystem(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, sys)

	// Test basic gauge
	err = sys.Gauge("cpu_usage_percent", 75.5, map[string]string{"host": "server1"})
	assert.NoError(t, err)

	// Test gauge with zero value
	err = sys.Gauge("memory_usage_bytes", 0, map[string]string{"host": "server1"})
	assert.NoError(t, err)

	// Test gauge with negative value (valid for things like temperature differential)
	err = sys.Gauge("temperature_delta", -5.2, map[string]string{"sensor": "outdoor"})
	assert.NoError(t, err)

	// Test gauge with no tags
	err = sys.Gauge("simple_gauge", 42.0, nil)
	assert.NoError(t, err)
}

// TestGaugeDisabled verifies gauge behavior when telemetry is disabled
func TestGaugeDisabled(t *testing.T) {
	config := &Config{Enabled: false}
	sys, err := NewSystem(config)
	require.NoError(t, err)
	require.NotNil(t, sys)

	// Should return nil when disabled
	err = sys.Gauge("cpu_usage_percent", 75.5, nil)
	assert.NoError(t, err)
}

// TestGaugeTypes tests different gauge value types
func TestGaugeTypes(t *testing.T) {
	sys, err := NewSystem(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, sys)

	// Test integer values
	err = sys.Gauge("integer_gauge", 100, nil)
	assert.NoError(t, err)

	// Test float values
	err = sys.Gauge("float_gauge", 123.456, nil)
	assert.NoError(t, err)

	// Test very small values
	err = sys.Gauge("small_gauge", 0.001, nil)
	assert.NoError(t, err)

	// Test very large values
	err = sys.Gauge("large_gauge", 1e9, nil)
	assert.NoError(t, err)
}

// TestGaugeRealWorldScenarios tests common gauge use cases
func TestGaugeRealWorldScenarios(t *testing.T) {
	sys, err := NewSystem(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, sys)

	// CPU usage percentage
	err = sys.Gauge("system_cpu_usage_percent", 45.7, map[string]string{
		"host": "web-server-01",
		"cpu":  "cpu0",
	})
	assert.NoError(t, err)

	// Memory usage in bytes
	err = sys.Gauge("system_memory_usage_bytes", 8589934592, map[string]string{
		"host": "web-server-01",
		"type": "used",
	})
	assert.NoError(t, err)

	// Temperature in Celsius
	err = sys.Gauge("environment_temperature_celsius", 23.5, map[string]string{
		"sensor": "indoor",
		"room":   "office",
	})
	assert.NoError(t, err)

	// Disk space percentage
	err = sys.Gauge("disk_usage_percent", 67.3, map[string]string{
		"host":  "web-server-01",
		"disk":  "/dev/sda1",
		"mount": "/",
	})
	assert.NoError(t, err)

	// Network connections
	err = sys.Gauge("network_connections_active", 142, map[string]string{
		"host":     "web-server-01",
		"protocol": "tcp",
	})
	assert.NoError(t, err)
}

// TestGaugeWithBatches tests gauge emission with batching enabled
func TestGaugeWithBatches(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     2,
		BatchInterval: 100 * time.Millisecond,
	}
	sys, err := NewSystem(config)
	require.NoError(t, err)
	require.NotNil(t, sys)

	// Emit multiple gauges - should be batched
	err = sys.Gauge("gauge1", 1.0, nil)
	assert.NoError(t, err)

	err = sys.Gauge("gauge2", 2.0, nil)
	assert.NoError(t, err)

	// Third gauge should trigger batch flush
	err = sys.Gauge("gauge3", 3.0, nil)
	assert.NoError(t, err)

	// Manually flush any remaining metrics
	err = sys.Flush()
	assert.NoError(t, err)
}
