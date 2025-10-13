package logging

import "go.uber.org/zap/zapcore"

// Severity represents log severity levels matching crucible schema
type Severity string

const (
	TRACE Severity = "TRACE" // Level 0
	DEBUG Severity = "DEBUG" // Level 10
	INFO  Severity = "INFO"  // Level 20
	WARN  Severity = "WARN"  // Level 30
	ERROR Severity = "ERROR" // Level 40
	FATAL Severity = "FATAL" // Level 50
	NONE  Severity = "NONE"  // Level 60 (disables logging)
)

// Level returns the numeric level for this severity
func (s Severity) Level() int {
	switch s {
	case TRACE:
		return 0
	case DEBUG:
		return 10
	case INFO:
		return 20
	case WARN:
		return 30
	case ERROR:
		return 40
	case FATAL:
		return 50
	case NONE:
		return 60
	default:
		return 20 // Default to INFO
	}
}

// ToZapLevel converts Fulmen severity to zap level
func (s Severity) ToZapLevel() zapcore.Level {
	switch s {
	case TRACE, DEBUG:
		return zapcore.DebugLevel
	case INFO:
		return zapcore.InfoLevel
	case WARN:
		return zapcore.WarnLevel
	case ERROR:
		return zapcore.ErrorLevel
	case FATAL:
		return zapcore.FatalLevel
	case NONE:
		return zapcore.InvalidLevel // Will filter out all logs
	default:
		return zapcore.InfoLevel
	}
}

// String returns the string representation
func (s Severity) String() string {
	return string(s)
}

// ParseSeverity parses a severity string
func ParseSeverity(s string) Severity {
	switch s {
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "NONE":
		return NONE
	default:
		return INFO
	}
}

// IsEnabled checks if this severity level should log given a minimum level
func (s Severity) IsEnabled(minLevel Severity) bool {
	return s.Level() >= minLevel.Level()
}
