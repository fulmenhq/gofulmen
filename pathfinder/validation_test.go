package pathfinder

import (
	"testing"

	"github.com/fulmenhq/gofulmen/errors"
	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	testingutil "github.com/fulmenhq/gofulmen/telemetry/testing"
)

func TestValidatePathResultWithEnvelope_InvalidInput(t *testing.T) {
	invalidResult := PathResult{
		RelativePath: "",
		SourcePath:   "",
		LogicalPath:  "",
		LoaderType:   "",
	}

	err := ValidatePathResultWithEnvelope(invalidResult, "test-correlation-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	envelope, ok := err.(*errors.ErrorEnvelope)
	if !ok {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}

	if envelope.Code != "PATHFINDER_VALIDATION_ERROR" {
		t.Errorf("expected code %q, got %q", "PATHFINDER_VALIDATION_ERROR", envelope.Code)
	}

	if envelope.CorrelationID != "test-correlation-id" {
		t.Errorf("expected correlation ID %q, got %q", "test-correlation-id", envelope.CorrelationID)
	}

	if envelope.Context == nil {
		t.Error("expected non-nil context")
	}
}

func TestValidatePathResultsWithEnvelope_MixedInput(t *testing.T) {
	invalidResult := PathResult{
		RelativePath: "",
		SourcePath:   "",
		LogicalPath:  "",
		LoaderType:   "",
	}

	tests := []struct {
		name        string
		results     []PathResult
		expectError bool
	}{
		{
			name:        "contains invalid result",
			results:     []PathResult{invalidResult},
			expectError: true,
		},
		{
			name:        "empty results",
			results:     []PathResult{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathResultsWithEnvelope(tt.results, "test-correlation-id")

			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectError {
				envelope, ok := err.(*errors.ErrorEnvelope)
				if !ok {
					t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
				}

				if envelope.Code != "PATHFINDER_VALIDATION_ERROR" {
					t.Errorf("expected code %q, got %q", "PATHFINDER_VALIDATION_ERROR", envelope.Code)
				}
			}
		})
	}
}

func TestValidatePathWithinRootWithEnvelope(t *testing.T) {
	tests := []struct {
		name          string
		absPath       string
		absRoot       string
		expectError   bool
		expectedCode  string
		checkSentinel error
	}{
		{
			name:        "valid path within root",
			absPath:     "/home/user/project/file.go",
			absRoot:     "/home/user/project",
			expectError: false,
		},
		{
			name:          "path escapes root with ..",
			absPath:       "/home/user/etc/passwd",
			absRoot:       "/home/user/project",
			expectError:   true,
			expectedCode:  "PATHFINDER_SECURITY_ERROR",
			checkSentinel: ErrEscapesRoot,
		},
		{
			name:          "non-absolute path",
			absPath:       "relative/path",
			absRoot:       "/home/user/project",
			expectError:   true,
			expectedCode:  "PATHFINDER_VALIDATION_ERROR",
			checkSentinel: ErrInvalidPath,
		},
		{
			name:          "non-absolute root",
			absPath:       "/home/user/project/file.go",
			absRoot:       "relative/root",
			expectError:   true,
			expectedCode:  "PATHFINDER_VALIDATION_ERROR",
			checkSentinel: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathWithinRootWithEnvelope(tt.absPath, tt.absRoot, "test-correlation-id")

			if tt.expectError && err == nil {
				t.Fatal("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectError {
				envelope, ok := err.(*errors.ErrorEnvelope)
				if !ok {
					t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
				}

				if envelope.Code != tt.expectedCode {
					t.Errorf("expected code %q, got %q", tt.expectedCode, envelope.Code)
				}

				if envelope.CorrelationID != "test-correlation-id" {
					t.Errorf("expected correlation ID %q, got %q", "test-correlation-id", envelope.CorrelationID)
				}

				if envelope.Context == nil {
					t.Error("expected non-nil context")
				}

				if tt.checkSentinel != nil {
					if envelope.Original == nil {
						t.Error("expected non-nil Original field")
					}
				}
			}
		})
	}
}

func TestValidatePathResultWithTelemetry(t *testing.T) {
	collector := testingutil.NewFakeCollector()

	invalidResult := PathResult{
		RelativePath: "",
		SourcePath:   "",
		LogicalPath:  "",
		LoaderType:   "",
	}

	config := telemetry.DefaultConfig()
	config.Enabled = true
	config.Emitter = collector
	telSys, err := telemetry.NewSystem(config)
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	err = validatePathResultWithTelemetry(invalidResult, "test-correlation-id", telSys)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	envelope, ok := err.(*errors.ErrorEnvelope)
	if !ok {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}

	if envelope.Code != "PATHFINDER_VALIDATION_ERROR" {
		t.Errorf("expected code %q, got %q", "PATHFINDER_VALIDATION_ERROR", envelope.Code)
	}

	metricResults := collector.GetMetricsByName(metrics.PathfinderValidationErrors)
	if len(metricResults) == 0 {
		t.Fatalf("expected %s metric to be emitted", metrics.PathfinderValidationErrors)
	}

	counterValue, ok := metricResults[0].Value.(float64)
	if !ok {
		t.Fatalf("expected float64 value, got %T", metricResults[0].Value)
	}

	if counterValue != 1 {
		t.Errorf("expected counter value 1, got %f", counterValue)
	}

	if metricResults[0].Tags["error_type"] == "" {
		t.Error("expected error_type tag to be set")
	}
}
