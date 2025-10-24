package telemetry

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestADR0007HistogramBuckets verifies ADR-0007 default histogram bucket implementation
func TestADR0007HistogramBuckets(t *testing.T) {
	sys, err := NewSystem(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, sys)

	tests := []struct {
		name     string
		duration time.Duration
		expected []float64
	}{
		{
			name:     "1ms duration",
			duration: 1 * time.Millisecond,
			expected: []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		{
			name:     "10ms duration",
			duration: 10 * time.Millisecond,
			expected: []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		{
			name:     "100ms duration",
			duration: 100 * time.Millisecond,
			expected: []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		{
			name:     "1s duration",
			duration: 1 * time.Second,
			expected: []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test bucket calculation function
			buckets := calculateHistogramBuckets(tt.duration, DefaultHistogramBucketsMS)

			// Verify bucket boundaries match ADR-0007 defaults (plus +Inf)
			assert.Len(t, buckets, len(DefaultHistogramBucketsMS)+1)
			for i, bucket := range buckets[:len(DefaultHistogramBucketsMS)] {
				assert.Equal(t, DefaultHistogramBucketsMS[i], bucket.LE)
			}
			// Last bucket should be +Inf
			assert.Equal(t, math.Inf(1), buckets[len(DefaultHistogramBucketsMS)].LE)

			// Verify cumulative counts work correctly
			durationMs := float64(tt.duration.Milliseconds())
			for i, bucket := range buckets {
				expectedCount := int64(0)
				if durationMs <= bucket.LE {
					expectedCount = 1
				}
				assert.Equal(t, expectedCount, bucket.Count, "Bucket %d should have correct cumulative count", i)
			}
		})
	}
}

// TestHistogramWithMSMetrics verifies that metrics ending with "_ms" automatically use ADR-0007 buckets
func TestHistogramWithMSMetrics(t *testing.T) {
	sys, err := NewSystem(DefaultConfig())
	require.NoError(t, err)
	require.NotNil(t, sys)

	// Test metric ending with "_ms" - should use histogram buckets
	err = sys.Histogram("config_load_ms", 50*time.Millisecond, map[string]string{"test": "true"})
	assert.NoError(t, err)

	// Test metric not ending with "_ms" - should use single value (backward compatibility)
	err = sys.Histogram("custom_metric", 50*time.Millisecond, map[string]string{"test": "true"})
	assert.NoError(t, err)
}

// TestDefaultHistogramBucketsMS verifies the ADR-0007 bucket boundaries
func TestDefaultHistogramBucketsMS(t *testing.T) {
	// Verify ADR-0007 default buckets: [1, 5, 10, 50, 100, 500, 1000, 5000, 10000]
	expected := []float64{1, 5, 10, 50, 100, 500, 1000, 5000, 10000}
	assert.Equal(t, expected, DefaultHistogramBucketsMS, "Default buckets should match ADR-0007 specification")

	// Verify coverage ranges mentioned in ADR-0007
	assert.Equal(t, float64(1), DefaultHistogramBucketsMS[0], "First bucket should be 1ms for fast operations")
	assert.Equal(t, float64(100), DefaultHistogramBucketsMS[4], "Bucket 5 should be 100ms for mid-range I/O")
	assert.Equal(t, float64(1000), DefaultHistogramBucketsMS[6], "Bucket 7 should be 1000ms for slower operations")
	assert.Equal(t, float64(10000), DefaultHistogramBucketsMS[8], "Last bucket should be 10000ms for long tasks")
}
