package logging

import (
	"testing"

	"github.com/fulmenhq/gofulmen/telemetry"
	"github.com/fulmenhq/gofulmen/telemetry/metrics"
	telemetrytesting "github.com/fulmenhq/gofulmen/telemetry/testing"
)

func TestLoggingTelemetryDisabledByDefault(t *testing.T) {
	config := DefaultConfig("test-service")

	if config.EnableTelemetry {
		t.Error("Expected telemetry to be disabled by default")
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message")
	_ = logger.Sync()
}

func TestLoggingTelemetryEmitsMetrics(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := DefaultConfig("test-service")
	config.DefaultLevel = "DEBUG"
	config.EnableTelemetry = true
	config.TelemetrySystem = sys

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message 1")
	logger.Debug("test message 2")
	logger.Warn("test message 3")
	_ = logger.Sync()

	if !fc.HasMetric(metrics.LoggingEmitCount) {
		t.Error("Expected logging_emit_count metric to be emitted")
	}

	emitCount := fc.CountMetricsByName(metrics.LoggingEmitCount)
	if emitCount != 3 {
		t.Errorf("Expected 3 emit count metrics, got %d", emitCount)
	}

	if !fc.HasMetric(metrics.LoggingEmitLatencyMs) {
		t.Error("Expected logging_emit_latency_ms metric to be emitted")
	}

	latencyCount := fc.CountMetricsByName(metrics.LoggingEmitLatencyMs)
	if latencyCount != 3 {
		t.Errorf("Expected 3 latency metrics, got %d", latencyCount)
	}
}

func TestLoggingTelemetryIncludesSeverityTag(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := DefaultConfig("test-service")
	config.EnableTelemetry = true
	config.TelemetrySystem = sys

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("info message")
	logger.Error("error message")
	_ = logger.Sync()

	emitMetrics := fc.GetMetricsByName(metrics.LoggingEmitCount)
	if len(emitMetrics) != 2 {
		t.Fatalf("Expected 2 emit metrics, got %d", len(emitMetrics))
	}

	foundInfo := false
	foundError := false
	for _, m := range emitMetrics {
		severity := m.Tags[metrics.TagSeverity]
		if severity == "info" {
			foundInfo = true
		}
		if severity == "error" {
			foundError = true
		}
	}

	if !foundInfo {
		t.Error("Expected to find info severity tag")
	}
	if !foundError {
		t.Error("Expected to find error severity tag")
	}
}

func TestLoggingTelemetryWithMiddleware(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := &LoggerConfig{
		Profile:         ProfileStructured,
		DefaultLevel:    "INFO",
		Service:         "test-service",
		Environment:     "test",
		EnableTelemetry: true,
		TelemetrySystem: sys,
		Sinks: []SinkConfig{
			{
				Type:   "console",
				Format: "json",
				Console: &ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: false,
				},
			},
		},
		Middleware: []MiddlewareConfig{
			{
				Name:    "correlation",
				Enabled: true,
				Order:   5,
			},
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test with middleware")
	_ = logger.Sync()

	if !fc.HasMetric(metrics.LoggingEmitCount) {
		t.Error("Expected metrics to be emitted with middleware enabled")
	}

	emitCount := fc.CountMetricsByName(metrics.LoggingEmitCount)
	if emitCount != 1 {
		t.Errorf("Expected 1 emit count metric, got %d", emitCount)
	}
}

func TestLoggingTelemetryWithThrottling(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := &LoggerConfig{
		Profile:         ProfileStructured,
		DefaultLevel:    "INFO",
		Service:         "test-service",
		Environment:     "test",
		EnableTelemetry: true,
		TelemetrySystem: sys,
		Sinks: []SinkConfig{
			{
				Type:   "console",
				Format: "json",
				Console: &ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: false,
				},
			},
		},
		Middleware: []MiddlewareConfig{
			{
				Name:    "correlation",
				Enabled: true,
				Order:   5,
			},
		},
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  100,
			WindowSize: 1,
			DropPolicy: "drop-oldest",
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	for i := 0; i < 5; i++ {
		logger.Info("throttled message")
	}
	_ = logger.Sync()

	emitCount := fc.CountMetricsByName(metrics.LoggingEmitCount)
	if emitCount != 5 {
		t.Errorf("Expected 5 emit count metrics, got %d", emitCount)
	}
}

func TestLoggingTelemetryDisabled(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := DefaultConfig("test-service")
	config.EnableTelemetry = false
	config.TelemetrySystem = sys

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message")
	_ = logger.Sync()

	emitCount := fc.CountMetricsByName(metrics.LoggingEmitCount)
	if emitCount != 0 {
		t.Errorf("Expected 0 metrics when telemetry disabled, got %d", emitCount)
	}
}

func TestLoggingTelemetryWithChildLoggers(t *testing.T) {
	fc := telemetrytesting.NewFakeCollector()

	sys, err := telemetry.NewSystem(&telemetry.Config{
		Enabled: true,
		Emitter: fc,
	})
	if err != nil {
		t.Fatalf("Failed to create telemetry system: %v", err)
	}

	config := DefaultConfig("test-service")
	config.EnableTelemetry = true
	config.TelemetrySystem = sys

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	childLogger := logger.WithComponent("child")
	childLogger.Info("child message")

	namedLogger := logger.Named("named")
	namedLogger.Info("named message")

	fieldsLogger := logger.WithFields(map[string]any{"key": "value"})
	fieldsLogger.Info("fields message")

	_ = logger.Sync()

	emitCount := fc.CountMetricsByName(metrics.LoggingEmitCount)
	if emitCount != 3 {
		t.Errorf("Expected 3 emit count metrics from child loggers, got %d", emitCount)
	}
}
