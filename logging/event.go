package logging

import "time"

// LogEvent represents a structured log event matching crucible schema
type LogEvent struct {
	Timestamp     time.Time      `json:"timestamp"`
	Severity      Severity       `json:"severity"`
	SeverityLevel int            `json:"severityLevel,omitempty"`
	Message       string         `json:"message"`
	Logger        string         `json:"logger,omitempty"`
	Service       string         `json:"service"`
	Component     string         `json:"component,omitempty"`
	Environment   string         `json:"environment,omitempty"`
	Context       map[string]any `json:"context,omitempty"`
	Error         *LogError      `json:"error,omitempty"`
	TraceID       string         `json:"traceId,omitempty"`
	SpanID        string         `json:"spanId,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	EventID       string         `json:"eventId,omitempty"`
}

// LogError represents error information in log events
type LogError struct {
	Message string         `json:"message"`
	Type    string         `json:"type"`
	Stack   string         `json:"stack,omitempty"`
	Code    string         `json:"code,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}
