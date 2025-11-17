package logging

// Helper functions for common logging configurations
// Following Crucible best practices and v0.1.15 plan goals

// WithRedaction creates a redaction middleware config with default or custom patterns
// This is a convenience helper for the common CLI/server pattern mentioned in v0.1.15 plan
func WithRedaction(patterns []string, fields []string) MiddlewareConfig {
	cfg := &RedactionConfig{
		Patterns:    patterns,
		Fields:      fields,
		Replacement: "[REDACTED]",
	}

	// Use defaults if none provided
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = DefaultRedactionPatterns
	}
	if len(cfg.Fields) == 0 {
		cfg.Fields = DefaultRedactionFields
	}

	return MiddlewareConfig{
		Type:      "redaction",
		Enabled:   true,
		Priority:  10, // Crucible recommended default (line 349)
		Redaction: cfg,
	}
}

// WithDefaultRedaction creates a redaction middleware config with all default patterns
// Convenience for: WithRedaction(nil, nil)
func WithDefaultRedaction() MiddlewareConfig {
	return WithRedaction(nil, nil)
}

// WithMinimalRedaction creates a redaction config for SIMPLE profile
// Targets known env vars and tokens per Crucible recommendations (line 280)
func WithMinimalRedaction() MiddlewareConfig {
	return WithRedaction(
		[]string{
			`SECRET_[A-Z0-9_]+`,         // SECRET_KEY, etc.
			`[A-Z0-9_]*TOKEN[A-Z0-9_]*`, // GITHUB_TOKEN, etc.
		},
		[]string{
			"password",
			"token",
			"secret",
		},
	)
}

// WithCorrelation creates a correlation middleware config
// TODO: Implement correlation middleware (for now, placeholder)
func WithCorrelation() MiddlewareConfig {
	return MiddlewareConfig{
		Type:     "correlation",
		Enabled:  true,
		Priority: 20, // Crucible recommended (line 351)
	}
}

// BundleSimpleWithRedaction creates a middleware bundle for SIMPLE profile with redaction
// Common pattern for CLI tools handling secrets (per v0.1.15 plan goal #4)
func BundleSimpleWithRedaction() []MiddlewareConfig {
	return []MiddlewareConfig{
		WithMinimalRedaction(),
	}
}

// BundleStructuredWithRedaction creates a middleware bundle for STRUCTURED profile
// Common pattern for services with correlation + redaction (per v0.1.15 plan goal #4)
func BundleStructuredWithRedaction() []MiddlewareConfig {
	return []MiddlewareConfig{
		WithDefaultRedaction(),
		WithCorrelation(),
	}
}
