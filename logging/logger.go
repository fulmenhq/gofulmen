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
}

// New creates a new logger from configuration
func New(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
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
	}
}
