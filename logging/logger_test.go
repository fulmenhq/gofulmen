package logging

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	config := DefaultConfig("test-service")

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger is nil")
	}

	// Test basic logging
	logger.Info("test message")
	_ = logger.Sync()
}

func TestNewCLI(t *testing.T) {
	logger, err := NewCLI("test-cli")
	if err != nil {
		t.Fatalf("Failed to create CLI logger: %v", err)
	}

	if logger == nil {
		t.Fatal("Logger is nil")
	}

	logger.Info("CLI test message")
	_ = logger.Sync()
}

func TestWithFields(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	contextLogger := logger.WithFields(map[string]any{
		"userId": "user123",
		"action": "test",
	})

	contextLogger.Info("test with fields")
	_ = logger.Sync()
}

func TestWithError(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	err := os.ErrNotExist
	errorLogger := logger.WithError(err)

	errorLogger.Error("test with error")
	_ = logger.Sync()
}

func TestWithComponent(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	componentLogger := logger.WithComponent("pathfinder")
	componentLogger.Info("test with component")
	_ = logger.Sync()
}

func TestSetLevel(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	// Start at INFO
	if logger.GetLevel() != INFO {
		t.Errorf("Expected INFO level, got %s", logger.GetLevel())
	}

	// Change to DEBUG
	logger.SetLevel(DEBUG)
	if logger.GetLevel() != DEBUG {
		t.Errorf("Expected DEBUG level, got %s", logger.GetLevel())
	}

	// Change to ERROR
	logger.SetLevel(ERROR)
	if logger.GetLevel() != ERROR {
		t.Errorf("Expected ERROR level, got %s", logger.GetLevel())
	}
}

func TestFileOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-service",
		Sinks: []SinkConfig{
			{
				Type:   "file",
				Format: "json",
				File: &FileSinkConfig{
					Path:       logPath,
					MaxSize:    10,
					MaxBackups: 3,
					MaxAge:     7,
					Compress:   false,
				},
			},
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger with file sink: %v", err)
	}

	logger.Info("test file output")
	_ = logger.Sync()

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestMultipleSinks(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi.log")

	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "DEBUG",
		Service:      "test-service",
		Sinks: []SinkConfig{
			{
				Type:   "console",
				Format: "json",
				Console: &ConsoleSinkConfig{
					Stream: "stderr",
				},
			},
			{
				Type:   "file",
				Format: "json",
				File: &FileSinkConfig{
					Path:    logPath,
					MaxSize: 10,
				},
			},
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger with multiple sinks: %v", err)
	}

	logger.Info("multi-sink test")
	_ = logger.Sync()

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created for multi-sink logger")
	}
}

func TestStaticFields(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test-service",
		Sinks: []SinkConfig{
			{
				Type:   "console",
				Format: "json",
				Console: &ConsoleSinkConfig{
					Stream: "stderr",
				},
			},
		},
		StaticFields: map[string]any{
			"version": "1.0.0",
			"region":  "us-east-1",
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger with static fields: %v", err)
	}

	logger.Info("test static fields")
	_ = logger.Sync()
}

func TestLoggingMethods(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Test all logging level methods
	logger.Trace("trace message")
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warning message")
	logger.Error("error message")

	_ = logger.Sync()
}

func TestNamed(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	namedLogger := logger.Named("subsystem")
	if namedLogger == nil {
		t.Fatal("Named() returned nil logger")
	}

	namedLogger.Info("message from named logger")
	_ = logger.Sync()
}

func TestWithContext(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)

	if contextLogger == nil {
		t.Fatal("WithContext() returned nil logger")
	}

	contextLogger.Info("message with context")
	_ = logger.Sync()
}

func TestNew_ProfileSimple(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileSimple,
		DefaultLevel: "INFO",
		Service:      "test",
		Environment:  "test",
		Sinks:        []SinkConfig{}, // Empty - should add console
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = logger.Sync() }()

	if len(config.Sinks) == 0 {
		t.Error("SIMPLE profile should add console sink")
	}
	if len(config.Sinks) > 0 && config.Sinks[0].Type != "console" {
		t.Errorf("SIMPLE profile added sink type = %s, want console", config.Sinks[0].Type)
	}
}

func TestNew_ProfileStructured(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "test",
		Environment:  "test",
		Sinks: []SinkConfig{
			{Type: "console", Format: "json"},
		},
		Middleware: []MiddlewareConfig{}, // Empty - should add correlation
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = logger.Sync() }()

	hasCorrelation := false
	for _, mw := range config.Middleware {
		if mw.Name == "correlation" {
			hasCorrelation = true
			break
		}
	}
	if !hasCorrelation {
		t.Error("STRUCTURED profile should add correlation middleware")
	}
}

func TestNew_ProfileEnterprise(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileEnterprise,
		DefaultLevel: "INFO",
		Service:      "test",
		Environment:  "production",
		PolicyFile:   "testdata/policies/permissive-policy.yaml",
		Sinks: []SinkConfig{
			{Type: "console", Format: "json"},
		},
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 5},
			{Name: "redact-secrets", Enabled: true, Order: 10},
		},
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  100,
			DropPolicy: "drop-oldest",
		},
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = logger.Sync() }()

	if logger.pipeline == nil {
		t.Error("ENTERPRISE profile should have pipeline")
	}
}

func TestNew_ProfileEnterpriseValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *LoggerConfig
		wantErr bool
	}{
		{
			name: "missing_sinks",
			config: &LoggerConfig{
				Profile:      ProfileEnterprise,
				DefaultLevel: "INFO",
				Service:      "test",
			},
			wantErr: true,
		},
		{
			name: "missing_middleware",
			config: &LoggerConfig{
				Profile:      ProfileEnterprise,
				DefaultLevel: "INFO",
				Service:      "test",
				Sinks: []SinkConfig{
					{Type: "console", Format: "json"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing_throttling",
			config: &LoggerConfig{
				Profile:      ProfileEnterprise,
				DefaultLevel: "INFO",
				Service:      "test",
				Sinks: []SinkConfig{
					{Type: "console", Format: "json"},
				},
				Middleware: []MiddlewareConfig{
					{Name: "correlation", Enabled: true},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if logger != nil {
				_ = logger.Sync()
			}
		})
	}
}

func TestNew_ProfileCustom(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileCustom,
		DefaultLevel: "INFO",
		Service:      "test",
		Environment:  "test",
		Sinks:        []SinkConfig{}, // Custom allows anything
	}

	logger, err := New(config)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// CUSTOM profile should not modify config
	if len(config.Sinks) != 0 {
		t.Error("CUSTOM profile should not add sinks")
	}
}

func TestBuildMiddlewarePipeline(t *testing.T) {
	config := &LoggerConfig{
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 5},
			{Name: "redact-secrets", Enabled: true, Order: 10},
			{Name: "redact-pii", Enabled: false, Order: 11}, // Disabled
		},
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    500,
			BurstSize:  50,
			DropPolicy: "drop-newest",
		},
	}

	pipeline, err := buildMiddlewarePipeline(config)
	if err != nil {
		t.Fatalf("buildMiddlewarePipeline() error = %v", err)
	}

	if pipeline == nil {
		t.Fatal("buildMiddlewarePipeline() returned nil pipeline")
	}

	// Should have 3 middleware (correlation, redact-secrets, throttle)
	// redact-pii is disabled
	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)
	if result == nil {
		t.Error("pipeline.Process() returned nil for valid event")
	}
}

func TestBuildMiddlewarePipeline_Empty(t *testing.T) {
	config := &LoggerConfig{
		Middleware: []MiddlewareConfig{},
	}

	pipeline, err := buildMiddlewarePipeline(config)
	if err != nil {
		t.Fatalf("buildMiddlewarePipeline() error = %v", err)
	}

	if pipeline == nil {
		t.Fatal("buildMiddlewarePipeline() returned nil pipeline")
	}

	// Empty pipeline should pass through events
	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)
	if result == nil {
		t.Error("empty pipeline should pass through events")
	}
}

func TestNew_ProfileEnterpriseValidation_DisabledMiddleware(t *testing.T) {
	config := &LoggerConfig{
		Profile:      ProfileEnterprise,
		DefaultLevel: "INFO",
		Service:      "test",
		Environment:  "production",
		PolicyFile:   "testdata/policies/permissive-policy.yaml",
		Sinks: []SinkConfig{
			{Type: "console", Format: "json"},
		},
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: false, Order: 5},
			{Name: "redact-secrets", Enabled: false, Order: 10},
		},
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  100,
			DropPolicy: "drop-oldest",
		},
	}

	_, err := New(config)
	if err == nil {
		t.Error("ENTERPRISE profile should reject config with all middleware disabled")
	}
}

func TestBuildMiddlewarePipeline_OrderOverride(t *testing.T) {
	config := &LoggerConfig{
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100},
			{Name: "redact-secrets", Enabled: true, Order: 5},
		},
	}

	pipeline, err := buildMiddlewarePipeline(config)
	if err != nil {
		t.Fatalf("buildMiddlewarePipeline() error = %v", err)
	}

	if len(pipeline.middleware) != 2 {
		t.Fatalf("Expected 2 middleware, got %d", len(pipeline.middleware))
	}

	// First should be redact-secrets (order 5 overridden from default 10)
	if pipeline.middleware[0].Name() != "redact-secrets" {
		t.Errorf("First middleware = %v, want redact-secrets", pipeline.middleware[0].Name())
	}
	if pipeline.middleware[0].Order() != 5 {
		t.Errorf("First middleware order = %d, want 5 (overridden)", pipeline.middleware[0].Order())
	}

	// Second should be correlation (order 100 overridden from default 5)
	if pipeline.middleware[1].Name() != "correlation" {
		t.Errorf("Second middleware = %v, want correlation", pipeline.middleware[1].Name())
	}
	if pipeline.middleware[1].Order() != 100 {
		t.Errorf("Second middleware order = %d, want 100 (overridden)", pipeline.middleware[1].Order())
	}
}
