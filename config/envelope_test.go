package config

import (
	"os"
	"testing"

	"github.com/fulmenhq/gofulmen/errors"
	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	testingutil "github.com/fulmenhq/gofulmen/telemetry/testing"
)

func TestLoadEnvOverridesWithEnvelope_ParseError(t *testing.T) {
	specs := []EnvVarSpec{
		{
			Name: "TEST_INT_VAR",
			Path: []string{"test", "value"},
			Type: EnvInt,
		},
	}

	_ = os.Setenv("TEST_INT_VAR", "not-an-integer")
	defer func() { _ = os.Unsetenv("TEST_INT_VAR") }()

	_, err := LoadEnvOverridesWithEnvelope(specs, "test-correlation-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	envelope, ok := err.(*errors.ErrorEnvelope)
	if !ok {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}

	if envelope.Code != "CONFIG_ENV_PARSE_ERROR" {
		t.Errorf("expected code %q, got %q", "CONFIG_ENV_PARSE_ERROR", envelope.Code)
	}

	if envelope.CorrelationID != "test-correlation-id" {
		t.Errorf("expected correlation ID %q, got %q", "test-correlation-id", envelope.CorrelationID)
	}

	if envelope.Context == nil {
		t.Error("expected non-nil context")
	}

	if envelope.Original == nil {
		t.Error("expected non-nil original error")
	}
}

func TestLoadEnvOverridesWithEnvelope_Success(t *testing.T) {
	specs := []EnvVarSpec{
		{
			Name: "TEST_STRING_VAR",
			Path: []string{"test", "string"},
			Type: EnvString,
		},
		{
			Name: "TEST_INT_VAR",
			Path: []string{"test", "number"},
			Type: EnvInt,
		},
	}

	_ = os.Setenv("TEST_STRING_VAR", "hello")
	_ = os.Setenv("TEST_INT_VAR", "42")
	defer func() {
		_ = os.Unsetenv("TEST_STRING_VAR")
		_ = os.Unsetenv("TEST_INT_VAR")
	}()

	result, err := LoadEnvOverridesWithEnvelope(specs, "test-correlation-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	testMap, ok := result["test"].(map[string]any)
	if !ok {
		t.Fatal("expected test key to be map")
	}

	if testMap["string"] != "hello" {
		t.Errorf("expected string value %q, got %v", "hello", testMap["string"])
	}

	if testMap["number"] != 42 {
		t.Errorf("expected number value 42, got %v", testMap["number"])
	}
}

func TestGetXDGBaseDirsWithEnvelope_MissingHome(t *testing.T) {
	originalHome := os.Getenv("HOME")
	_ = os.Unsetenv("HOME")
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	_, err := GetXDGBaseDirsWithEnvelope("test-correlation-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	envelope, ok := err.(*errors.ErrorEnvelope)
	if !ok {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}

	if envelope.Code != "CONFIG_XDG_ERROR" {
		t.Errorf("expected code %q, got %q", "CONFIG_XDG_ERROR", envelope.Code)
	}

	if envelope.CorrelationID != "test-correlation-id" {
		t.Errorf("expected correlation ID %q, got %q", "test-correlation-id", envelope.CorrelationID)
	}

	if envelope.Context == nil {
		t.Error("expected non-nil context")
	}
}

func TestGetXDGBaseDirsWithEnvelope_Success(t *testing.T) {
	originalHome := os.Getenv("HOME")
	testHome := "/tmp/testhome"
	_ = os.Setenv("HOME", testHome)
	defer func() {
		if originalHome != "" {
			_ = os.Setenv("HOME", originalHome)
		}
	}()

	result, err := GetXDGBaseDirsWithEnvelope("test-correlation-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ConfigHome == "" {
		t.Error("expected non-empty ConfigHome")
	}

	if result.DataHome == "" {
		t.Error("expected non-empty DataHome")
	}

	if result.CacheHome == "" {
		t.Error("expected non-empty CacheHome")
	}
}

func TestLoadLayeredConfigWithEnvelope_MetricsEmission(t *testing.T) {
	collector := testingutil.NewFakeCollector()

	config := telemetry.DefaultConfig()
	config.Enabled = true
	config.Emitter = collector
	telSys, err := telemetry.NewSystem(config)
	if err != nil {
		t.Fatalf("failed to create telemetry system: %v", err)
	}

	telemetryOnce.Do(func() {})
	telemetrySystem = telSys

	opts := LayeredConfigOptions{
		Category:     "",
		Version:      "",
		DefaultsFile: "",
		SchemaID:     "",
	}

	_, _, err = LoadLayeredConfigWithEnvelope(opts, "test-correlation-id")
	if err == nil {
		t.Fatal("expected error for missing parameters")
	}

	envelope, ok := err.(*errors.ErrorEnvelope)
	if !ok {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}

	if envelope.Code != "CONFIG_LOAD_ERROR" {
		t.Errorf("expected code %q, got %q", "CONFIG_LOAD_ERROR", envelope.Code)
	}

	counterMetrics := collector.GetMetricsByName(metrics.ConfigLoadErrors)
	if len(counterMetrics) == 0 {
		t.Fatalf("expected %s metric to be emitted", metrics.ConfigLoadErrors)
	}

	counterValue, ok := counterMetrics[0].Value.(float64)
	if !ok {
		t.Fatalf("expected float64 value, got %T", counterMetrics[0].Value)
	}

	if counterValue != 1 {
		t.Errorf("expected counter value 1, got %f", counterValue)
	}

	if counterMetrics[0].Tags["error_type"] != "missing_parameters" {
		t.Errorf("expected error_type %q, got %q", "missing_parameters", counterMetrics[0].Tags["error_type"])
	}

	histogramMetrics := collector.GetMetricsByName(metrics.ConfigLoadMs)
	if len(histogramMetrics) == 0 {
		t.Fatalf("expected %s metric to be emitted", metrics.ConfigLoadMs)
	}

	if histogramMetrics[0].Tags[metrics.TagStatus] != metrics.StatusError {
		t.Errorf("expected status %q, got %q", metrics.StatusError, histogramMetrics[0].Tags[metrics.TagStatus])
	}
}
