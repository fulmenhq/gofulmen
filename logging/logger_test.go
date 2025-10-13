package logging

import (
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
	logger.Sync()
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
	logger.Sync()
}

func TestWithFields(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	contextLogger := logger.WithFields(map[string]any{
		"userId": "user123",
		"action": "test",
	})

	contextLogger.Info("test with fields")
	logger.Sync()
}

func TestWithError(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	err := os.ErrNotExist
	errorLogger := logger.WithError(err)

	errorLogger.Error("test with error")
	logger.Sync()
}

func TestWithComponent(t *testing.T) {
	config := DefaultConfig("test-service")
	logger, _ := New(config)

	componentLogger := logger.WithComponent("pathfinder")
	componentLogger.Info("test with component")
	logger.Sync()
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
	// Create temp directory for logs
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	config := &LoggerConfig{
		DefaultLevel: "INFO",
		Service:      "test-service",
		Environment:  "test",
		Sinks: []SinkConfig{
			{
				Type:   "file",
				Format: "json",
				File: &FileSinkConfig{
					Path:       logPath,
					MaxSize:    10,
					MaxAge:     7,
					MaxBackups: 3,
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
	logger.Sync()

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

func TestMultipleSinks(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "multi.log")

	config := &LoggerConfig{
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
	logger.Sync()

	// Verify file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created for multi-sink logger")
	}
}

func TestStaticFields(t *testing.T) {
	config := &LoggerConfig{
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
	logger.Sync()
}
