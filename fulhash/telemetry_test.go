package fulhash

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry"
	teltesting "github.com/fulmenhq/gofulmen/telemetry/testing"
)

func TestHash_TelemetryEmission(t *testing.T) {
	collector := teltesting.NewFakeCollector()
	telSys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetry.SetGlobalSystem(telSys)
	defer telemetry.SetGlobalSystem(nil)

	tests := []struct {
		name           string
		data           []byte
		opts           []Option
		wantAlg        string
		wantAlgMetric  string
		wantCount      int
		wantBytesCount int
	}{
		{
			name:           "xxh3-128 success",
			data:           []byte("test data"),
			opts:           []Option{WithAlgorithm(XXH3_128)},
			wantAlg:        "xxh3-128",
			wantAlgMetric:  "fulhash_operations_total_xxh3_128",
			wantCount:      1,
			wantBytesCount: 9, // len("test data")
		},
		{
			name:           "sha256 success",
			data:           []byte("test data"),
			opts:           []Option{WithAlgorithm(SHA256)},
			wantAlg:        "sha256",
			wantAlgMetric:  "fulhash_operations_total_sha256",
			wantCount:      1,
			wantBytesCount: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.Reset()

			_, err := Hash(tt.data, tt.opts...)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			// Check algorithm-specific operation counter
			algMetrics := collector.GetMetricsByName(tt.wantAlgMetric)
			if len(algMetrics) != tt.wantCount {
				t.Errorf("expected %d %s metrics, got %d", tt.wantCount, tt.wantAlgMetric, len(algMetrics))
			}

			if len(algMetrics) > 0 {
				tags := algMetrics[0].Tags
				if tags["algorithm"] != tt.wantAlg {
					t.Errorf("expected algorithm=%s, got %s", tt.wantAlg, tags["algorithm"])
				}
			}

			// Check bytes hashed counter
			bytesMetrics := collector.GetMetricsByName("fulhash_bytes_hashed_total")
			if len(bytesMetrics) != 1 {
				t.Errorf("expected 1 bytes_hashed_total metric, got %d", len(bytesMetrics))
			}

			// Check operation latency histogram
			latencyMetrics := collector.GetMetricsByName("fulhash_operation_ms")
			if len(latencyMetrics) != 1 {
				t.Errorf("expected 1 operation_ms metric, got %d", len(latencyMetrics))
			}

			// Ensure no error metrics
			errorMetrics := collector.GetMetricsByName("fulhash_errors_count")
			if len(errorMetrics) != 0 {
				t.Errorf("expected no error metrics, got %d", len(errorMetrics))
			}
		})
	}
}

func TestHash_UnsupportedAlgorithm_TelemetryEmission(t *testing.T) {
	collector := teltesting.NewFakeCollector()
	telSys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetry.SetGlobalSystem(telSys)
	defer telemetry.SetGlobalSystem(nil)

	_, err = Hash([]byte("data"), WithAlgorithm("invalid"))
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}

	errorMetrics := collector.GetMetricsByName("fulhash_errors_count")
	if len(errorMetrics) != 1 {
		t.Fatalf("expected 1 error metric, got %d", len(errorMetrics))
	}

	tags := errorMetrics[0].Tags
	if tags["status"] != "error" {
		t.Errorf("expected status=error, got %s", tags["status"])
	}
	if tags["error_type"] != "unsupported_algorithm" {
		t.Errorf("expected error_type=unsupported_algorithm, got %s", tags["error_type"])
	}

	// Ensure no success metrics were emitted (check all algorithm-specific counters)
	xxh3Metrics := collector.GetMetricsByName("fulhash_operations_total_xxh3_128")
	if len(xxh3Metrics) != 0 {
		t.Errorf("expected no xxh3_128 operation metrics, got %d", len(xxh3Metrics))
	}

	sha256Metrics := collector.GetMetricsByName("fulhash_operations_total_sha256")
	if len(sha256Metrics) != 0 {
		t.Errorf("expected no sha256 operation metrics, got %d", len(sha256Metrics))
	}
}

func TestHashString_TelemetryEmission(t *testing.T) {
	collector := teltesting.NewFakeCollector()
	telSys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetry.SetGlobalSystem(telSys)
	defer telemetry.SetGlobalSystem(nil)

	_, err = HashString("test string")
	if err != nil {
		t.Fatalf("HashString() error = %v", err)
	}

	// Check hash_string_total counter
	stringMetrics := collector.GetMetricsByName("fulhash_hash_string_total")
	if len(stringMetrics) != 1 {
		t.Errorf("expected 1 hash_string_total metric, got %d", len(stringMetrics))
	}

	// Check algorithm-specific counter (default is xxh3-128)
	xxh3Metrics := collector.GetMetricsByName("fulhash_operations_total_xxh3_128")
	if len(xxh3Metrics) != 1 {
		t.Errorf("expected 1 xxh3_128 operation metric, got %d", len(xxh3Metrics))
	}

	if len(xxh3Metrics) > 0 {
		tags := xxh3Metrics[0].Tags
		if tags["algorithm"] != "xxh3-128" {
			t.Errorf("expected algorithm=xxh3-128, got %s", tags["algorithm"])
		}
	}

	// Check bytes hashed counter
	bytesMetrics := collector.GetMetricsByName("fulhash_bytes_hashed_total")
	if len(bytesMetrics) != 1 {
		t.Errorf("expected 1 bytes_hashed_total metric, got %d", len(bytesMetrics))
	}

	// Check operation latency histogram
	latencyMetrics := collector.GetMetricsByName("fulhash_operation_ms")
	if len(latencyMetrics) != 1 {
		t.Errorf("expected 1 operation_ms metric, got %d", len(latencyMetrics))
	}
}

func TestHashReader_TelemetryEmission(t *testing.T) {
	collector := teltesting.NewFakeCollector()
	telSys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetry.SetGlobalSystem(telSys)
	defer telemetry.SetGlobalSystem(nil)

	data := []byte("test data from reader")
	reader := bytes.NewReader(data)

	_, err = HashReader(reader)
	if err != nil {
		t.Fatalf("HashReader() error = %v", err)
	}

	// Check algorithm-specific counter (default is xxh3-128)
	xxh3Metrics := collector.GetMetricsByName("fulhash_operations_total_xxh3_128")
	if len(xxh3Metrics) != 1 {
		t.Errorf("expected 1 xxh3_128 operation metric, got %d", len(xxh3Metrics))
	}

	if len(xxh3Metrics) > 0 {
		tags := xxh3Metrics[0].Tags
		if tags["algorithm"] != "xxh3-128" {
			t.Errorf("expected algorithm=xxh3-128, got %s", tags["algorithm"])
		}
	}

	// Check bytes hashed counter
	bytesMetrics := collector.GetMetricsByName("fulhash_bytes_hashed_total")
	if len(bytesMetrics) != 1 {
		t.Errorf("expected 1 bytes_hashed_total metric, got %d", len(bytesMetrics))
	}

	// Check operation latency histogram
	latencyMetrics := collector.GetMetricsByName("fulhash_operation_ms")
	if len(latencyMetrics) != 1 {
		t.Errorf("expected 1 operation_ms metric, got %d", len(latencyMetrics))
	}

	// Ensure no error metrics
	errorMetrics := collector.GetMetricsByName("fulhash_errors_count")
	if len(errorMetrics) != 0 {
		t.Errorf("expected no error metrics, got %d", len(errorMetrics))
	}
}

func TestHashReader_IOError_TelemetryEmission(t *testing.T) {
	collector := teltesting.NewFakeCollector()
	telSys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: collector,
	})
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetry.SetGlobalSystem(telSys)
	defer telemetry.SetGlobalSystem(nil)

	reader := &errorReader{}

	_, err = HashReader(reader)
	if err == nil {
		t.Fatal("expected I/O error from reader")
	}

	errorMetrics := collector.GetMetricsByName("fulhash_errors_count")
	if len(errorMetrics) != 1 {
		t.Fatalf("expected 1 error metric, got %d", len(errorMetrics))
	}

	tags := errorMetrics[0].Tags
	if tags["status"] != "error" {
		t.Errorf("expected status=error, got %s", tags["status"])
	}
	if tags["error_type"] != "io_error" {
		t.Errorf("expected error_type=io_error, got %s", tags["error_type"])
	}

	// Ensure no success metrics were emitted
	xxh3Metrics := collector.GetMetricsByName("fulhash_operations_total_xxh3_128")
	if len(xxh3Metrics) != 0 {
		t.Errorf("expected no xxh3_128 operation metrics, got %d", len(xxh3Metrics))
	}

	sha256Metrics := collector.GetMetricsByName("fulhash_operations_total_sha256")
	if len(sha256Metrics) != 0 {
		t.Errorf("expected no sha256 operation metrics, got %d", len(sha256Metrics))
	}
}

func TestHash_TelemetryDisabled(t *testing.T) {
	telemetry.SetGlobalSystem(nil)

	_, err := Hash([]byte("test"))
	if err != nil {
		t.Fatalf("Hash() should work without telemetry: %v", err)
	}

	_, err = HashString("test")
	if err != nil {
		t.Fatalf("HashString() should work without telemetry: %v", err)
	}

	_, err = HashReader(strings.NewReader("test"))
	if err != nil {
		t.Fatalf("HashReader() should work without telemetry: %v", err)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}
