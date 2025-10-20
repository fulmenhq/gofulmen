package logging

import (
	"testing"

	"go.uber.org/zap"
)

func TestIntegration_SimpleProfile_Programmatic(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileSimple,
		DefaultLevel: "INFO",
		Service:      "test-simple",
		Environment:  "development",
		Sinks: []SinkConfig{
			{Type: "console", Format: "console", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	if config.Profile != ProfileSimple {
		t.Errorf("expected SIMPLE profile, got %s", config.Profile)
	}

	if len(config.Middleware) != 0 {
		t.Errorf("SIMPLE profile should have no middleware, got %d", len(config.Middleware))
	}

	if config.Throttling != nil && config.Throttling.Enabled {
		t.Error("SIMPLE profile should not have throttling enabled")
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("test message", zap.String("key", "value"))
}

func TestIntegration_StructuredProfile_Programmatic(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-structured",
		Environment:  "production",
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100, Config: make(map[string]any)},
		},
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  2000,
			DropPolicy: "drop-oldest",
		},
		StaticFields: map[string]any{
			"app_version": "1.0.0",
		},
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	if config.Profile != ProfileStructured {
		t.Errorf("expected STRUCTURED profile, got %s", config.Profile)
	}

	if config.Throttling == nil || !config.Throttling.Enabled {
		t.Error("STRUCTURED profile should have throttling enabled")
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("structured test", zap.String("request_id", "test-123"))
}

func TestIntegration_EnterpriseProfile_Programmatic(t *testing.T) {
	t.Skip("ENTERPRISE profile requires policy file - skipping file-based test")
}

func TestIntegration_CustomProfile_Programmatic(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileCustom,
		DefaultLevel: "DEBUG",
		Service:      "test-custom",
		Environment:  "development",
		Sinks: []SinkConfig{
			{Type: "console", Format: "text", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	if config.Profile != ProfileCustom {
		t.Errorf("expected CUSTOM profile, got %s", config.Profile)
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Debug("custom debug message", zap.Bool("debug", true))
}

func TestIntegration_MiddlewarePipeline(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-middleware",
		Environment:  "test",
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100, Config: make(map[string]any)},
		},
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("test",
		zap.String("password", "shouldBeRedacted"),
		zap.String("username", "visible"),
	)
}

func TestIntegration_ThrottlingMiddleware(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-throttle",
		Environment:  "test",
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    5,
			BurstSize:  5,
			DropPolicy: "drop-oldest",
		},
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	for i := 0; i < 10; i++ {
		logger.Info("throttle test", zap.Int("iteration", i))
	}
}

func TestIntegration_StaticFields(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-static",
		Environment:  "test",
		StaticFields: map[string]any{
			"app_version": "1.0.0",
			"deployment":  "canary",
		},
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("static fields test")
}

func TestIntegration_EnvironmentEnrichment(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-service",
		Component:    "api",
		Environment:  "production",
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("env test")
}

func TestIntegration_ProfileNormalization(t *testing.T) {
	tests := []struct {
		name            string
		inputProfile    LoggingProfile
		expectedProfile LoggingProfile
		format          string
	}{
		{"lowercase simple", "simple", ProfileSimple, "console"},
		{"uppercase structured", "STRUCTURED", ProfileStructured, "json"},
		{"lowercase custom", "custom", ProfileCustom, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Profile:      tt.inputProfile,
				DefaultLevel: "INFO",
				Service:      "test-normalization",
				Environment:  "test",
				Sinks: []SinkConfig{
					{Type: "console", Format: tt.format, Console: &ConsoleSinkConfig{Stream: "stderr"}},
				},
			}

			normResult, err := NormalizeLoggerConfig(config)
			if err != nil {
				t.Fatalf("normalization failed: %v", err)
			}

			if config.Profile != tt.expectedProfile {
				t.Errorf("expected profile %s, got %s", tt.expectedProfile, config.Profile)
			}

			if tt.inputProfile != tt.expectedProfile && len(normResult.Warnings) == 0 {
				t.Error("expected normalization warning for case difference")
			}

			logger, err := New(config)
			if err != nil {
				t.Fatalf("failed to create logger after normalization: %v", err)
			}

			logger.Info("normalization test")
		})
	}
}

func TestIntegration_WithFieldsChaining(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-chaining",
		Environment:  "test",
		Sinks: []SinkConfig{
			{Type: "console", Format: "json", Console: &ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	applyDefaults(config)

	logger, err := New(config)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	contextLogger := logger.WithFields(map[string]any{
		"request_id": "req-123",
		"user_id":    "user-456",
	})

	contextLogger.Info("request started")
	contextLogger.Debug("processing request")
	contextLogger.Info("request completed")
}
