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

	SetTelemetrySystem(telSys)
	defer SetTelemetrySystem(nil)

	tests := []struct {
		name      string
		data      []byte
		opts      []Option
		wantAlg   string
		wantCount int
	}{
		{
			name:      "xxh3-128 success",
			data:      []byte("test data"),
			opts:      []Option{WithAlgorithm(XXH3_128)},
			wantAlg:   "xxh3-128",
			wantCount: 1,
		},
		{
			name:      "sha256 success",
			data:      []byte("test data"),
			opts:      []Option{WithAlgorithm(SHA256)},
			wantAlg:   "sha256",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.Reset()

			_, err := Hash(tt.data, tt.opts...)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			metrics := collector.GetMetricsByName("fulhash_hash_count")
			if len(metrics) != tt.wantCount {
				t.Errorf("expected %d fulhash_hash_count metrics, got %d", tt.wantCount, len(metrics))
			}

			if len(metrics) > 0 {
				tags := metrics[0].Tags
				if tags["algorithm"] != tt.wantAlg {
					t.Errorf("expected algorithm=%s, got %s", tt.wantAlg, tags["algorithm"])
				}
				if tags["status"] != "success" {
					t.Errorf("expected status=success, got %s", tags["status"])
				}
			}

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

	SetTelemetrySystem(telSys)
	defer SetTelemetrySystem(nil)

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

	successMetrics := collector.GetMetricsByName("fulhash_hash_count")
	if len(successMetrics) != 0 {
		t.Errorf("expected no success metrics, got %d", len(successMetrics))
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

	SetTelemetrySystem(telSys)
	defer SetTelemetrySystem(nil)

	_, err = HashString("test string")
	if err != nil {
		t.Fatalf("HashString() error = %v", err)
	}

	metrics := collector.GetMetricsByName("fulhash_hash_count")
	if len(metrics) != 1 {
		t.Errorf("expected 1 fulhash_hash_count metric, got %d", len(metrics))
	}

	if len(metrics) > 0 {
		tags := metrics[0].Tags
		if tags["algorithm"] != "xxh3-128" {
			t.Errorf("expected algorithm=xxh3-128, got %s", tags["algorithm"])
		}
		if tags["status"] != "success" {
			t.Errorf("expected status=success, got %s", tags["status"])
		}
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

	SetTelemetrySystem(telSys)
	defer SetTelemetrySystem(nil)

	data := []byte("test data from reader")
	reader := bytes.NewReader(data)

	_, err = HashReader(reader)
	if err != nil {
		t.Fatalf("HashReader() error = %v", err)
	}

	metrics := collector.GetMetricsByName("fulhash_hash_count")
	if len(metrics) != 1 {
		t.Errorf("expected 1 fulhash_hash_count metric, got %d", len(metrics))
	}

	if len(metrics) > 0 {
		tags := metrics[0].Tags
		if tags["algorithm"] != "xxh3-128" {
			t.Errorf("expected algorithm=xxh3-128, got %s", tags["algorithm"])
		}
		if tags["status"] != "success" {
			t.Errorf("expected status=success, got %s", tags["status"])
		}
	}

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

	SetTelemetrySystem(telSys)
	defer SetTelemetrySystem(nil)

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

	successMetrics := collector.GetMetricsByName("fulhash_hash_count")
	if len(successMetrics) != 0 {
		t.Errorf("expected no success metrics, got %d", len(successMetrics))
	}
}

func TestHash_TelemetryDisabled(t *testing.T) {
	SetTelemetrySystem(nil)

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
