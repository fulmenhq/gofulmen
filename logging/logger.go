package logging

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger wraps zap with Fulmen configuration and middleware
type Logger struct {
	zap          *zap.Logger
	config       *LoggerConfig
	atomicLevel  zap.AtomicLevel
	staticFields map[string]any
	pipeline     *MiddlewarePipeline
}

// New creates a new logger from configuration
func New(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize profile-specific defaults FIRST
	if err := initializeProfileDefaults(config); err != nil {
		return nil, fmt.Errorf("profile initialization failed: %w", err)
	}

	// Load and enforce policy if specified
	if config.PolicyFile != "" {
		policy, err := LoadPolicy(config.PolicyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load policy: %w", err)
		}
		if err := EnforcePolicy(config, policy, config.Environment, ""); err != nil {
			return nil, fmt.Errorf("policy enforcement failed: %w", err)
		}
	}

	// Validate profile requirements AFTER defaults applied
	// Filter to only enabled middleware for validation
	enabledMiddleware := make([]MiddlewareConfig, 0, len(config.Middleware))
	for _, mw := range config.Middleware {
		if mw.Enabled {
			enabledMiddleware = append(enabledMiddleware, mw)
		}
	}

	throttlingEnabled := config.Throttling != nil && config.Throttling.Enabled
	policyEnabled := config.PolicyFile != ""
	errs := ValidateProfileRequirements(
		config.Profile,
		config.Sinks,
		enabledMiddleware, // Only validate enabled middleware
		"",                // format not used in current implementation
		throttlingEnabled,
		policyEnabled,
	)
	if len(errs) > 0 {
		return nil, fmt.Errorf("profile requirements validation failed: %v", errs[0])
	}

	// Build middleware pipeline
	pipeline, err := buildMiddlewarePipeline(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build middleware pipeline: %w", err)
	}

	// Parse default level
	level := ParseSeverity(config.DefaultLevel).ToZapLevel()
	atomicLevel := zap.NewAtomicLevelAt(level)

	// Build encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "severity",
		NameKey:        "logger",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    severityEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Build cores for each sink
	var cores []zapcore.Core
	for _, sinkConfig := range config.Sinks {
		core, err := buildCore(sinkConfig, encoderConfig, atomicLevel)
		if err != nil {
			return nil, fmt.Errorf("failed to build sink %s: %w", sinkConfig.Type, err)
		}
		cores = append(cores, core)
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Build logger options
	opts := []zap.Option{
		zap.AddCaller(),
	}

	if config.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Add static fields
	if len(config.StaticFields) > 0 {
		fields := make([]zap.Field, 0, len(config.StaticFields))
		for k, v := range config.StaticFields {
			fields = append(fields, zap.Any(k, v))
		}
		opts = append(opts, zap.Fields(fields...))
	}

	// Always add service field
	opts = append(opts, zap.Fields(
		zap.String("service", config.Service),
	))

	// Add environment if specified
	if config.Environment != "" {
		opts = append(opts, zap.Fields(
			zap.String("environment", config.Environment),
		))
	}

	// Create zap logger
	zapLogger := zap.New(core, opts...)

	return &Logger{
		zap:          zapLogger,
		config:       config,
		atomicLevel:  atomicLevel,
		staticFields: config.StaticFields,
		pipeline:     pipeline,
	}, nil
}

// NewCLI creates a logger configured for CLI applications (stderr only)
func NewCLI(serviceName string) (*Logger, error) {
	config := &LoggerConfig{
		DefaultLevel: "INFO",
		Service:      serviceName,
		Environment:  "cli",
		Sinks: []SinkConfig{
			{
				Type:   "console",
				Format: "console",
				Console: &ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: true,
				},
			},
		},
		EnableCaller:     false,
		EnableStacktrace: true,
	}
	return New(config)
}

// buildCore creates a zapcore.Core for a sink configuration
func buildCore(sinkConfig SinkConfig, encoderConfig zapcore.EncoderConfig, defaultLevel zap.AtomicLevel) (zapcore.Core, error) {
	// Determine encoder
	var encoder zapcore.Encoder
	switch sinkConfig.Format {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console", "text":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Determine writer
	var writer zapcore.WriteSyncer
	switch sinkConfig.Type {
	case "console":
		writer = buildConsoleWriter(sinkConfig)
	case "file":
		w, err := buildFileWriter(sinkConfig)
		if err != nil {
			return nil, err
		}
		writer = w
	default:
		return nil, fmt.Errorf("unsupported sink type: %s", sinkConfig.Type)
	}

	// Determine level for this sink
	level := defaultLevel
	if sinkConfig.Level != "" {
		sinkLevel := ParseSeverity(sinkConfig.Level).ToZapLevel()
		level = zap.NewAtomicLevelAt(sinkLevel)
	}

	return zapcore.NewCore(encoder, writer, level), nil
}

// buildConsoleWriter creates a writer for console sink
func buildConsoleWriter(sinkConfig SinkConfig) zapcore.WriteSyncer {
	// Always stderr (enforced by schema validation)
	return zapcore.AddSync(os.Stderr)
}

// buildFileWriter creates a writer for file sink with rotation
func buildFileWriter(sinkConfig SinkConfig) (zapcore.WriteSyncer, error) {
	if sinkConfig.File == nil {
		return nil, fmt.Errorf("file sink requires file configuration")
	}

	lumber := &lumberjack.Logger{
		Filename:   sinkConfig.File.Path,
		MaxSize:    sinkConfig.File.MaxSize,    // MB
		MaxAge:     sinkConfig.File.MaxAge,     // days
		MaxBackups: sinkConfig.File.MaxBackups, // number of backups
		Compress:   sinkConfig.File.Compress,
	}

	return zapcore.AddSync(lumber), nil
}

// severityEncoder encodes levels as Fulmen severity strings
func severityEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var severity string
	switch l {
	case zapcore.DebugLevel:
		severity = "DEBUG"
	case zapcore.InfoLevel:
		severity = "INFO"
	case zapcore.WarnLevel:
		severity = "WARN"
	case zapcore.ErrorLevel:
		severity = "ERROR"
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		severity = "FATAL"
	default:
		severity = "INFO"
	}
	enc.AppendString(severity)
}

// Core logging methods

// Trace logs at TRACE level
func (l *Logger) Trace(msg string, fields ...zap.Field) {
	// Zap doesn't have TRACE, use DEBUG
	l.zap.Debug(msg, fields...)
}

// Debug logs at DEBUG level
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Info logs at INFO level
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Warn logs at WARN level
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Error logs at ERROR level
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal logs at FATAL level and terminates
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	return &Logger{
		zap:          l.zap.With(zapFields...),
		config:       l.config,
		atomicLevel:  l.atomicLevel,
		staticFields: l.staticFields,
		pipeline:     l.pipeline,
	}
}

// WithError returns a logger with error information
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		zap: l.zap.With(
			zap.Error(err),
		),
		config:       l.config,
		atomicLevel:  l.atomicLevel,
		staticFields: l.staticFields,
		pipeline:     l.pipeline,
	}
}

// WithComponent returns a logger with component field
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		zap: l.zap.With(
			zap.String("component", component),
		),
		config:       l.config,
		atomicLevel:  l.atomicLevel,
		staticFields: l.staticFields,
		pipeline:     l.pipeline,
	}
}

// WithContext extracts trace information from context
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace/span IDs if available (placeholder for future tracing integration)
	// For now, just return the logger
	return l
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// SetLevel dynamically changes the log level
func (l *Logger) SetLevel(severity Severity) {
	l.atomicLevel.SetLevel(severity.ToZapLevel())
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() Severity {
	zapLevel := l.atomicLevel.Level()

	switch zapLevel {
	case zapcore.DebugLevel:
		return DEBUG
	case zapcore.InfoLevel:
		return INFO
	case zapcore.WarnLevel:
		return WARN
	case zapcore.ErrorLevel:
		return ERROR
	case zapcore.FatalLevel:
		return FATAL
	default:
		return INFO
	}
}

// Named returns a logger with the specified name
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		zap:          l.zap.Named(name),
		config:       l.config,
		atomicLevel:  l.atomicLevel,
		staticFields: l.staticFields,
		pipeline:     l.pipeline,
	}
}

// initializeProfileDefaults applies profile-specific defaults and validations
func initializeProfileDefaults(config *LoggerConfig) error {
	switch config.Profile {
	case ProfileSimple:
		return initializeSimpleProfile(config)
	case ProfileStructured:
		return initializeStructuredProfile(config)
	case ProfileEnterprise:
		return initializeEnterpriseProfile(config)
	case ProfileCustom:
		return initializeCustomProfile(config)
	default:
		// Default to SIMPLE if not specified
		config.Profile = ProfileSimple
		return initializeSimpleProfile(config)
	}
}

// initializeSimpleProfile ensures console sink is present
func initializeSimpleProfile(config *LoggerConfig) error {
	// Ensure at least one console sink exists
	hasConsole := false
	for _, sink := range config.Sinks {
		if sink.Type == "console" {
			hasConsole = true
			break
		}
	}

	if !hasConsole {
		// Add default console sink
		config.Sinks = append(config.Sinks, SinkConfig{
			Type:   "console",
			Format: "console",
			Console: &ConsoleSinkConfig{
				Stream:   "stderr",
				Colorize: false,
			},
		})
	}

	return nil
}

// initializeStructuredProfile validates sinks are present and adds correlation
func initializeStructuredProfile(config *LoggerConfig) error {
	if len(config.Sinks) == 0 {
		return fmt.Errorf("STRUCTURED profile requires at least one sink")
	}

	// Ensure correlation middleware is present
	hasCorrelation := false
	for _, mw := range config.Middleware {
		if mw.Name == "correlation" {
			hasCorrelation = true
			break
		}
	}

	if !hasCorrelation {
		config.Middleware = append(config.Middleware, MiddlewareConfig{
			Name:    "correlation",
			Enabled: true,
			Order:   5,
		})
	}

	return nil
}

// initializeEnterpriseProfile validates middleware and throttling are configured
func initializeEnterpriseProfile(config *LoggerConfig) error {
	if len(config.Sinks) == 0 {
		return fmt.Errorf("ENTERPRISE profile requires at least one sink")
	}

	if len(config.Middleware) == 0 {
		return fmt.Errorf("ENTERPRISE profile requires middleware configuration")
	}

	if config.Throttling == nil || !config.Throttling.Enabled {
		return fmt.Errorf("ENTERPRISE profile requires throttling to be enabled")
	}

	return nil
}

// initializeCustomProfile is a no-op - accepts any configuration
func initializeCustomProfile(config *LoggerConfig) error {
	return nil
}

// buildMiddlewarePipeline constructs the middleware pipeline from config
func buildMiddlewarePipeline(config *LoggerConfig) (*MiddlewarePipeline, error) {
	var middleware []Middleware

	// Add configured middleware
	for _, mwConfig := range config.Middleware {
		if !mwConfig.Enabled {
			continue
		}

		// Create middleware from registry
		mw, err := DefaultRegistry().Create(mwConfig.Name, mwConfig.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create middleware %s: %w", mwConfig.Name, err)
		}

		// Override order if specified in config
		if mwConfig.Order > 0 {
			mw = &middlewareOrderOverride{
				Middleware: mw,
				order:      mwConfig.Order,
			}
		}

		middleware = append(middleware, mw)
	}

	// Add throttling middleware if enabled
	if config.Throttling != nil && config.Throttling.Enabled {
		throttleConfig := map[string]any{
			"maxRate":    config.Throttling.MaxRate,
			"burstSize":  config.Throttling.BurstSize,
			"dropPolicy": config.Throttling.DropPolicy,
		}
		mw, err := DefaultRegistry().Create("throttle", throttleConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create throttling middleware: %w", err)
		}
		middleware = append(middleware, mw)
	}

	return NewMiddlewarePipeline(middleware), nil
}

// middlewareOrderOverride wraps a middleware to override its Order() method
type middlewareOrderOverride struct {
	Middleware
	order int
}

func (m *middlewareOrderOverride) Order() int {
	return m.order
}
