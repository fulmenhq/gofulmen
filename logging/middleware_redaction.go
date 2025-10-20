package logging

import (
	"regexp"
	"strings"
)

// RedactSecretsMiddleware redacts sensitive credentials from log events.
//
// Detects and redacts common secret patterns in messages and context:
//   - API keys (sk_live_*, api_key=...)
//   - Bearer tokens
//   - Passwords (password=...)
//   - JWTs (eyJ...)
//   - Sensitive field names (password, token, apiKey, secret, etc.)
type RedactSecretsMiddleware struct {
	order    int
	patterns []*regexp.Regexp
}

// RedactPIIMiddleware redacts personally identifiable information from log events.
//
// Detects and redacts PII patterns in messages and context:
//   - Email addresses
//   - Phone numbers (US format)
//   - Social Security Numbers (SSN)
type RedactPIIMiddleware struct {
	order    int
	patterns []*regexp.Regexp
}

var (
	// Secret patterns
	apiKeyPattern      = regexp.MustCompile(`\b(sk_live_[a-zA-Z0-9]{24,}|api_?key\s*[=:]\s*['"]?[a-zA-Z0-9_\-]{16,}['"]?)`)
	bearerTokenPattern = regexp.MustCompile(`[Bb]earer\s+[a-zA-Z0-9_\-\.]{15,}`)
	passwordPattern    = regexp.MustCompile(`\bpassword\s*[=:]\s*['"]?[^\s'"]{8,}['"]?`)
	jwtPattern         = regexp.MustCompile(`\beyJ[a-zA-Z0-9_\-]+\.eyJ[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+`)

	// PII patterns
	emailPattern = regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Z|a-z]{2,}\b`)
	phonePattern = regexp.MustCompile(`(\+?1[-.\s]?)?(\()?\d{3}(\))?[-.\s]?\d{3}[-.\s]?\d{4}`)
	ssnPattern   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)

	// Sensitive field names
	sensitiveFields = map[string]bool{
		"password":      true,
		"passwd":        true,
		"pwd":           true,
		"secret":        true,
		"token":         true,
		"apikey":        true,
		"api_key":       true,
		"bearer":        true,
		"authorization": true,
		"auth":          true,
		"privatekey":    true,
		"private_key":   true,
	}
)

const redactedPlaceholder = "[REDACTED]"

// NewRedactSecretsMiddleware creates a new secret redaction middleware instance.
func NewRedactSecretsMiddleware(config map[string]any) (Middleware, error) {
	order := 10
	if configOrder, ok := config["order"].(int); ok {
		order = configOrder
	} else if configOrder, ok := config["order"].(float64); ok {
		order = int(configOrder)
	}

	return &RedactSecretsMiddleware{
		order: order,
		patterns: []*regexp.Regexp{
			apiKeyPattern,
			bearerTokenPattern,
			passwordPattern,
			jwtPattern,
		},
	}, nil
}

// NewRedactPIIMiddleware creates a new PII redaction middleware instance.
func NewRedactPIIMiddleware(config map[string]any) (Middleware, error) {
	order := 15
	if configOrder, ok := config["order"].(int); ok {
		order = configOrder
	} else if configOrder, ok := config["order"].(float64); ok {
		order = int(configOrder)
	}

	return &RedactPIIMiddleware{
		order: order,
		patterns: []*regexp.Regexp{
			emailPattern,
			phonePattern,
			ssnPattern,
		},
	}, nil
}

// Process redacts secrets from message and context fields.
func (m *RedactSecretsMiddleware) Process(event *LogEvent) *LogEvent {
	if event == nil {
		return nil
	}

	redacted := false

	event.Message = m.redactString(event.Message, &redacted)

	if event.Context != nil {
		event.Context = m.redactContext(event.Context, &redacted)
	}

	if redacted {
		event.RedactionFlags = appendUnique(event.RedactionFlags, "secrets")
	}

	return event
}

// Process redacts PII from message and context fields.
func (m *RedactPIIMiddleware) Process(event *LogEvent) *LogEvent {
	if event == nil {
		return nil
	}

	redacted := false

	event.Message = m.redactString(event.Message, &redacted)

	if event.Context != nil {
		event.Context = m.redactContext(event.Context, &redacted)
	}

	if redacted {
		event.RedactionFlags = appendUnique(event.RedactionFlags, "pii")
	}

	return event
}

func (m *RedactSecretsMiddleware) redactString(s string, redacted *bool) string {
	result := s
	for _, pattern := range m.patterns {
		if pattern.MatchString(result) {
			result = pattern.ReplaceAllString(result, redactedPlaceholder)
			*redacted = true
		}
	}
	return result
}

func (m *RedactPIIMiddleware) redactString(s string, redacted *bool) string {
	result := s
	for _, pattern := range m.patterns {
		if pattern.MatchString(result) {
			result = pattern.ReplaceAllString(result, redactedPlaceholder)
			*redacted = true
		}
	}
	return result
}

func (m *RedactSecretsMiddleware) redactContext(ctx map[string]any, redacted *bool) map[string]any {
	result := make(map[string]any, len(ctx))
	for k, v := range ctx {
		if isSensitiveField(k) {
			result[k] = redactedPlaceholder
			*redacted = true
		} else {
			result[k] = m.redactValue(v, redacted)
		}
	}
	return result
}

func (m *RedactPIIMiddleware) redactContext(ctx map[string]any, redacted *bool) map[string]any {
	result := make(map[string]any, len(ctx))
	for k, v := range ctx {
		result[k] = m.redactValue(v, redacted)
	}
	return result
}

func (m *RedactSecretsMiddleware) redactValue(v any, redacted *bool) any {
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

func (m *RedactPIIMiddleware) redactValue(v any, redacted *bool) any {
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

func (m *RedactSecretsMiddleware) Order() int   { return m.order }
func (m *RedactSecretsMiddleware) Name() string { return "redact-secrets" }

func (m *RedactPIIMiddleware) Order() int   { return m.order }
func (m *RedactPIIMiddleware) Name() string { return "redact-pii" }

func isSensitiveField(fieldName string) bool {
	return sensitiveFields[strings.ToLower(fieldName)]
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func init() {
	DefaultRegistry().Register("redact-secrets", NewRedactSecretsMiddleware)
	DefaultRegistry().Register("redact-pii", NewRedactPIIMiddleware)
}
