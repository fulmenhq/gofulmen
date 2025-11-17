package logging

import (
	"fmt"
	"regexp"
	"strings"
)

// RedactionMiddleware redacts sensitive data from log events per Crucible spec.
// Follows middleware-config.schema.json#/$defs/redactionConfig
//
// Implements pattern-based and field-based redaction with:
// - Pre-compiled regex patterns for performance
// - Case-insensitive field matching
// - Generic "redacted" flag (future: pattern categorization)
// - Configurable replacement string
type RedactionMiddleware struct {
	priority       int
	patterns       []*regexp.Regexp
	patternStrings []string // For error reporting
	fields         map[string]bool
	replacement    string
	enabled        bool
}

// Default redaction patterns from Crucible spec (lines 252-257)
// These match the recommendations in logging.md but are opt-in
var DefaultRedactionPatterns = []string{
	`SECRET_[A-Z0-9_]+`,                              // Environment variables like SECRET_KEY
	`[A-Z0-9_]*TOKEN[A-Z0-9_]*`,                      // API tokens like GITHUB_TOKEN
	`[A-Z0-9_]*KEY[A-Z0-9_]*`,                        // Keys like API_KEY, DATABASE_KEY
	`[A-Za-z0-9+/]{40,}={0,2}`,                       // Base64 secrets (40+ characters)
	`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, // Email addresses
	`\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}`,         // Credit cards
}

// Default sensitive field names (case-insensitive)
var DefaultRedactionFields = []string{
	"password",
	"token",
	"apiKey",
	"authorization",
	"secret",
	"cardNumber",
	"cvv",
	"ssn",
}

// NewRedactionMiddleware creates a new redaction middleware from schema-compliant config
// Follows Crucible spec requirements:
// - MUST pre-compile patterns at initialization
// - MUST fail fast on invalid patterns
// - MUST use schema-defined replacement default
func NewRedactionMiddleware(config *RedactionConfig, priority int) (*RedactionMiddleware, error) {
	if config == nil {
		config = &RedactionConfig{}
	}

	// Set default replacement if not specified
	replacement := config.Replacement
	if replacement == "" {
		replacement = "[REDACTED]" // Schema default per middleware-config.schema.json
	}

	// Pre-compile all patterns (MUST per Crucible spec line 360)
	var compiledPatterns []*regexp.Regexp
	var patternStrings []string

	for _, pattern := range config.Patterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			// MUST fail at initialization per Crucible spec line 365
			return nil, fmt.Errorf("invalid redaction pattern %q: %w", pattern, err)
		}
		compiledPatterns = append(compiledPatterns, compiled)
		patternStrings = append(patternStrings, pattern)
	}

	// Build case-insensitive field map (SHOULD per Crucible spec line 264)
	fieldMap := make(map[string]bool)
	for _, field := range config.Fields {
		fieldMap[strings.ToLower(field)] = true
	}

	return &RedactionMiddleware{
		priority:       priority,
		patterns:       compiledPatterns,
		patternStrings: patternStrings,
		fields:         fieldMap,
		replacement:    replacement,
		enabled:        true,
	}, nil
}

// Process applies redaction to message and context fields
// Returns nil to drop event (though redaction doesn't drop)
// Sets redactionFlags per Crucible spec line 266
func (m *RedactionMiddleware) Process(event *LogEvent) *LogEvent {
	if event == nil || !m.enabled {
		return event
	}

	redacted := false

	// Redact message text
	event.Message = m.redactString(event.Message, &redacted)

	// Redact context fields
	if event.Context != nil {
		event.Context = m.redactContext(event.Context, &redacted)
	}

	// Set generic "redacted" flag per Crucible spec line 266
	// (future versions MAY categorize by pattern type)
	if redacted {
		event.RedactionFlags = appendUnique(event.RedactionFlags, "redacted")
	}

	return event
}

// redactString applies pattern-based redaction to a string
func (m *RedactionMiddleware) redactString(s string, redacted *bool) string {
	result := s
	for _, pattern := range m.patterns {
		if pattern.MatchString(result) {
			result = pattern.ReplaceAllString(result, m.replacement)
			*redacted = true
		}
	}
	return result
}

// redactContext applies field-based and pattern-based redaction to context map
// Field matching is case-insensitive per Crucible spec line 264
func (m *RedactionMiddleware) redactContext(ctx map[string]any, redacted *bool) map[string]any {
	result := make(map[string]any, len(ctx))
	for k, v := range ctx {
		// Check if field name matches (case-insensitive)
		if m.fields[strings.ToLower(k)] {
			result[k] = m.replacement
			*redacted = true
		} else {
			result[k] = m.redactValue(v, redacted)
		}
	}
	return result
}

// redactValue recursively redacts values (strings, maps, slices)
func (m *RedactionMiddleware) redactValue(v any, redacted *bool) any {
	switch val := v.(type) {
	case string:
		return m.redactString(val, redacted)
	case map[string]any:
		return m.redactContext(val, redacted)
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = m.redactValue(item, redacted)
		}
		return result
	case []string:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = m.redactString(item, redacted)
		}
		return result
	case []map[string]any:
		result := make([]map[string]any, len(val))
		for i, item := range val {
			result[i] = m.redactContext(item, redacted)
		}
		return result
	default:
		return v
	}
}

// Order returns middleware execution priority (lower runs first)
func (m *RedactionMiddleware) Order() int {
	return m.priority
}

// Name returns middleware name for logging/debugging
func (m *RedactionMiddleware) Name() string {
	return "redaction"
}

// DefaultRedactionConfig returns a sensible default redaction configuration
// using the patterns from Crucible spec
func DefaultRedactionConfig() *RedactionConfig {
	return &RedactionConfig{
		Patterns:    DefaultRedactionPatterns,
		Fields:      DefaultRedactionFields,
		Replacement: "[REDACTED]",
	}
}

// NewRedactionMiddlewareFactory creates a factory function for registry
// This adapts the new schema-based config to the existing factory interface
func NewRedactionMiddlewareFactory(cfg map[string]any) (Middleware, error) {
	// Extract priority (default: 10 per Crucible spec line 349)
	priority := 10
	if p, ok := cfg["priority"].(int); ok {
		priority = p
	} else if p, ok := cfg["priority"].(float64); ok {
		priority = int(p)
	}

	// Extract redaction config
	var redactionCfg *RedactionConfig
	if redactCfg, ok := cfg["redaction"].(map[string]any); ok {
		redactionCfg = &RedactionConfig{}

		// Extract patterns
		if patterns, ok := redactCfg["patterns"].([]any); ok {
			for _, p := range patterns {
				if str, ok := p.(string); ok {
					redactionCfg.Patterns = append(redactionCfg.Patterns, str)
				}
			}
		}

		// Extract fields
		if fields, ok := redactCfg["fields"].([]any); ok {
			for _, f := range fields {
				if str, ok := f.(string); ok {
					redactionCfg.Fields = append(redactionCfg.Fields, str)
				}
			}
		}

		// Extract replacement
		if repl, ok := redactCfg["replacement"].(string); ok {
			redactionCfg.Replacement = repl
		}
	}

	return NewRedactionMiddleware(redactionCfg, priority)
}
