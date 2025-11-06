package metrics_test

import (
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry/metrics"
)

// TestPrometheusExporterMetricNames ensures exporter metric names follow taxonomy conventions
func TestPrometheusExporterMetricNames(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantUnit string
	}{
		{"refresh duration", metrics.PrometheusExporterRefreshDurationSeconds, metrics.UnitSeconds},
		{"refresh total", metrics.PrometheusExporterRefreshTotal, metrics.UnitCount},
		{"refresh errors", metrics.PrometheusExporterRefreshErrorsTotal, metrics.UnitCount},
		{"refresh inflight", metrics.PrometheusExporterRefreshInflight, metrics.UnitCount},
		{"http requests", metrics.PrometheusExporterHTTPRequestsTotal, metrics.UnitCount},
		{"http errors", metrics.PrometheusExporterHTTPErrorsTotal, metrics.UnitCount},
		{"restarts", metrics.PrometheusExporterRestartsTotal, metrics.UnitCount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify metric name follows snake_case convention
			if strings.ToLower(tt.metric) != tt.metric {
				t.Errorf("metric %q should be lowercase snake_case", tt.metric)
			}

			// Verify metric name doesn't contain spaces or hyphens
			if strings.Contains(tt.metric, " ") || strings.Contains(tt.metric, "-") {
				t.Errorf("metric %q should not contain spaces or hyphens", tt.metric)
			}

			// Verify counter metrics end with _total
			if tt.wantUnit == metrics.UnitCount && tt.metric != metrics.PrometheusExporterRefreshInflight {
				if !strings.HasSuffix(tt.metric, "_total") && !strings.HasSuffix(tt.metric, "_inflight") {
					t.Errorf("counter metric %q should end with _total or _inflight", tt.metric)
				}
			}

			// Verify histogram metrics end with appropriate suffix
			if tt.wantUnit == metrics.UnitSeconds && !strings.HasSuffix(tt.metric, "_seconds") {
				t.Errorf("histogram metric %q should end with _seconds", tt.metric)
			}
		})
	}
}

// TestFoundryModuleMetricNames ensures Foundry metric names follow taxonomy conventions
func TestFoundryModuleMetricNames(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantUnit string
	}{
		{"json detections", metrics.FoundryMimeDetectionsTotalJSON, metrics.UnitCount},
		{"xml detections", metrics.FoundryMimeDetectionsTotalXML, metrics.UnitCount},
		{"yaml detections", metrics.FoundryMimeDetectionsTotalYAML, metrics.UnitCount},
		{"csv detections", metrics.FoundryMimeDetectionsTotalCSV, metrics.UnitCount},
		{"plain text detections", metrics.FoundryMimeDetectionsTotalPlainText, metrics.UnitCount},
		{"unknown detections", metrics.FoundryMimeDetectionsTotalUnknown, metrics.UnitCount},
		{"detection latency", metrics.FoundryMimeDetectionMs, metrics.UnitMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify metric name follows snake_case convention
			if strings.ToLower(tt.metric) != tt.metric {
				t.Errorf("metric %q should be lowercase snake_case", tt.metric)
			}

			// Verify metric prefix
			if !strings.HasPrefix(tt.metric, "foundry_") {
				t.Errorf("metric %q should start with foundry_ prefix", tt.metric)
			}
		})
	}
}

// TestErrorHandlingMetricNames ensures error handling metric names follow conventions
func TestErrorHandlingMetricNames(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantUnit string
	}{
		{"wraps total", metrics.ErrorHandlingWrapsTotal, metrics.UnitCount},
		{"wrap latency", metrics.ErrorHandlingWrapMs, metrics.UnitMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.metric, "error_handling_") {
				t.Errorf("metric %q should start with error_handling_ prefix", tt.metric)
			}
		})
	}
}

// TestFulHashMetricNames ensures FulHash metric names follow conventions
func TestFulHashMetricNames(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		wantUnit string
	}{
		{"xxh3_128 operations", metrics.FulHashOperationsTotalXXH3128, metrics.UnitCount},
		{"sha256 operations", metrics.FulHashOperationsTotalSHA256, metrics.UnitCount},
		{"hash string total", metrics.FulHashHashStringTotal, metrics.UnitCount},
		{"bytes hashed", metrics.FulHashBytesHashedTotal, metrics.UnitBytes},
		{"operation latency", metrics.FulHashOperationMs, metrics.UnitMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.metric, "fulhash_") {
				t.Errorf("metric %q should start with fulhash_ prefix", tt.metric)
			}
		})
	}
}

// TestLabelConstants verifies label key constants
func TestLabelConstants(t *testing.T) {
	labels := map[string]string{
		"status":     metrics.TagStatus,
		"component":  metrics.TagComponent,
		"operation":  metrics.TagOperation,
		"phase":      metrics.TagPhase,
		"result":     metrics.TagResult,
		"error_type": metrics.TagErrorType,
		"reason":     metrics.TagReason,
		"path":       metrics.TagPath,
		"client":     metrics.TagClient,
		"mime_type":  metrics.TagMimeType,
	}

	for expected, actual := range labels {
		if actual != expected {
			t.Errorf("label constant mismatch: expected %q, got %q", expected, actual)
		}
	}
}

// TestPhaseValues verifies phase enumeration values
func TestPhaseValues(t *testing.T) {
	phases := []string{
		metrics.PhaseCollect,
		metrics.PhaseConvert,
		metrics.PhaseExport,
	}

	expected := []string{"collect", "convert", "export"}

	for i, phase := range phases {
		if phase != expected[i] {
			t.Errorf("phase value mismatch at index %d: expected %q, got %q", i, expected[i], phase)
		}
	}
}

// TestResultValues verifies result enumeration values
func TestResultValues(t *testing.T) {
	if metrics.ResultSuccess != "success" {
		t.Errorf("ResultSuccess should be %q, got %q", "success", metrics.ResultSuccess)
	}
	if metrics.ResultError != "error" {
		t.Errorf("ResultError should be %q, got %q", "error", metrics.ResultError)
	}
}

// TestErrorTypeValues verifies error type enumeration values
func TestErrorTypeValues(t *testing.T) {
	errorTypes := map[string]string{
		"validation": metrics.ErrorTypeValidation,
		"io":         metrics.ErrorTypeIO,
		"timeout":    metrics.ErrorTypeTimeout,
		"other":      metrics.ErrorTypeOther,
	}

	for expected, actual := range errorTypes {
		if actual != expected {
			t.Errorf("error type mismatch: expected %q, got %q", expected, actual)
		}
	}
}

// TestRestartReasonValues verifies restart reason enumeration values
func TestRestartReasonValues(t *testing.T) {
	reasons := map[string]string{
		"config":        metrics.RestartReasonConfig,
		"panic_recover": metrics.RestartReasonPanicRecover,
		"manual":        metrics.RestartReasonManual,
		"dependency":    metrics.RestartReasonDependency,
	}

	for expected, actual := range reasons {
		if actual != expected {
			t.Errorf("restart reason mismatch: expected %q, got %q", expected, actual)
		}
	}
}
