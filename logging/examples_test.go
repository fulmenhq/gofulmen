package logging_test

import (
	"fmt"

	"github.com/fulmenhq/gofulmen/logging"
	"go.uber.org/zap"
)

func ExampleNew_simpleProfile() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileSimple,
		DefaultLevel: "INFO",
		Service:      "my-service",
		Environment:  "development",
		Sinks: []logging.SinkConfig{
			{
				Type:   "console",
				Format: "console",
				Console: &logging.ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: false,
				},
			},
		},
	}

	logger, err := logging.New(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger.Info("application started", zap.String("version", "1.0.0"))
}

func ExampleNew_structuredProfile() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "api-service",
		Environment:  "production",
		Middleware: []logging.MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100, Config: make(map[string]any)},
		},
		Throttling: &logging.ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  2000,
			DropPolicy: "drop-oldest",
		},
		StaticFields: map[string]any{
			"app_version": "2.0.0",
			"region":      "us-east-1",
		},
		Sinks: []logging.SinkConfig{
			{
				Type:   "console",
				Format: "json",
				Console: &logging.ConsoleSinkConfig{
					Stream: "stderr",
				},
			},
		},
	}

	logger, err := logging.New(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger.Info("processing request",
		zap.String("request_id", "req-123"),
		zap.Int("user_id", 456),
	)
}

func ExampleNew_customProfile() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileCustom,
		DefaultLevel: "DEBUG",
		Service:      "debug-service",
		Environment:  "development",
		Sinks: []logging.SinkConfig{
			{
				Type:   "console",
				Format: "text",
				Console: &logging.ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: true,
				},
			},
		},
		EnableCaller: true,
	}

	logger, err := logging.New(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger.Debug("debugging information", zap.String("trace_id", "abc-123"))
}

func ExampleLoadConfig() {
	config, err := logging.LoadConfig("testdata/configs/structured-profile.yaml")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger, err := logging.New(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger.Info("logger initialized from config file")
}

func ExampleLogger_WithFields() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "api-service",
		Environment:  "production",
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "json", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	logger, _ := logging.New(config)

	requestLogger := logger.WithFields(map[string]any{
		"request_id": "req-789",
		"user_id":    "user-456",
		"ip_address": "192.168.1.1",
	})

	requestLogger.Info("request received")
	requestLogger.Info("processing request")
	requestLogger.Info("request completed")
}

func ExampleLogger_WithComponent() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileStructured,
		DefaultLevel: "INFO",
		Service:      "api-service",
		Environment:  "production",
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "json", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	logger, _ := logging.New(config)

	dbLogger := logger.WithComponent("database")
	dbLogger.Info("connecting to database")
	dbLogger.Info("query executed")

	cacheLogger := logger.WithComponent("cache")
	cacheLogger.Info("cache hit")
}

func ExampleNormalizeLoggerConfig() {
	config := &logging.LoggerConfig{
		Profile:      "structured",
		DefaultLevel: "INFO",
		Service:      "my-service",
		Environment:  "production",
		Throttling: &logging.ThrottlingConfig{
			Enabled: true,
		},
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "json", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	result, err := logging.NormalizeLoggerConfig(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	fmt.Printf("Profile normalized: %s\n", config.Profile)
	fmt.Printf("Throttling maxRate: %d\n", config.Throttling.MaxRate)
	fmt.Printf("Warnings: %d\n", len(result.Warnings))
}

func Example_policyEnforcement() {
	config := &logging.LoggerConfig{
		Profile:      logging.ProfileEnterprise,
		DefaultLevel: "INFO",
		Service:      "secure-service",
		Environment:  "production",
		PolicyFile:   "testdata/policies/strict-enterprise-policy.yaml",
		Middleware: []logging.MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100, Config: make(map[string]any)},
		},
		Throttling: &logging.ThrottlingConfig{
			Enabled:    true,
			MaxRate:    5000,
			BurstSize:  10000,
			DropPolicy: "drop-oldest",
		},
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "json", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}

	logger, err := logging.New(config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	logger.Info("policy-compliant logging enabled")
}

func Example_progressiveLogging() {
	simple := &logging.LoggerConfig{
		Profile:     logging.ProfileSimple,
		Service:     "simple-app",
		Environment: "development",
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "console", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}
	simpleLogger, _ := logging.New(simple)
	simpleLogger.Info("simple logging - no middleware")

	structured := &logging.LoggerConfig{
		Profile:     logging.ProfileStructured,
		Service:     "structured-app",
		Environment: "staging",
		Middleware: []logging.MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100, Config: make(map[string]any)},
		},
		Throttling: &logging.ThrottlingConfig{Enabled: true, MaxRate: 1000, BurstSize: 2000},
		Sinks: []logging.SinkConfig{
			{Type: "console", Format: "json", Console: &logging.ConsoleSinkConfig{Stream: "stderr"}},
		},
	}
	structuredLogger, _ := logging.New(structured)
	structuredLogger.Info("structured logging with correlation")
}
