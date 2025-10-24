package errors

import (
	"fmt"
	"log"
)

// ErrorHandlingStrategy defines how to handle validation errors
type ErrorHandlingStrategy int

const (
	// StrategyLogWarning logs validation errors as warnings but continues execution
	StrategyLogWarning ErrorHandlingStrategy = iota

	// StrategyAppendToMessage appends validation errors to the envelope message
	StrategyAppendToMessage

	// StrategyFailFast returns the validation error immediately
	StrategyFailFast

	// StrategySilent ignores validation errors completely (use with caution)
	StrategySilent
)

// ErrorHandlingConfig configures error handling behavior
type ErrorHandlingConfig struct {
	SeverityStrategy ErrorHandlingStrategy
	ContextStrategy  ErrorHandlingStrategy
	Logger           *log.Logger
}

// DefaultErrorHandlingConfig returns sensible defaults for error handling
func DefaultErrorHandlingConfig() *ErrorHandlingConfig {
	return &ErrorHandlingConfig{
		SeverityStrategy: StrategyLogWarning,
		ContextStrategy:  StrategyAppendToMessage,
		Logger:           log.Default(),
	}
}

// ApplySeverityWithHandling applies severity with consistent error handling
func ApplySeverityWithHandling(envelope *ErrorEnvelope, severity Severity, config *ErrorHandlingConfig) *ErrorEnvelope {
	if config == nil {
		config = DefaultErrorHandlingConfig()
	}

	result, err := envelope.WithSeverity(severity)
	if err != nil {
		switch config.SeverityStrategy {
		case StrategyLogWarning:
			if config.Logger != nil {
				config.Logger.Printf("Warning: failed to set severity %q: %v", severity, err)
			}
			return envelope // Return original envelope

		case StrategyAppendToMessage:
			envelope.Message = fmt.Sprintf("%s (severity error: %v)", envelope.Message, err)
			return envelope

		case StrategyFailFast:
			// For fail-fast, we still return the envelope but log the issue
			if config.Logger != nil {
				config.Logger.Printf("Error: severity validation failed: %v", err)
			}
			return envelope

		case StrategySilent:
			return envelope
		}
	}
	return result
}

// ApplyContextWithHandling applies context with consistent error handling
func ApplyContextWithHandling(envelope *ErrorEnvelope, context map[string]interface{}, config *ErrorHandlingConfig) *ErrorEnvelope {
	if config == nil {
		config = DefaultErrorHandlingConfig()
	}

	result, err := envelope.WithContext(context)
	if err != nil {
		switch config.ContextStrategy {
		case StrategyLogWarning:
			if config.Logger != nil {
				config.Logger.Printf("Warning: failed to set context: %v", err)
			}
			return envelope // Return original envelope

		case StrategyAppendToMessage:
			envelope.Message = fmt.Sprintf("%s (context error: %v)", envelope.Message, err)
			return envelope

		case StrategyFailFast:
			// For fail-fast, we still return the envelope but log the issue
			if config.Logger != nil {
				config.Logger.Printf("Error: context validation failed: %v", err)
			}
			return envelope

		case StrategySilent:
			return envelope
		}
	}
	return result
}

// SafeWithSeverity is a convenience wrapper that uses default error handling
func SafeWithSeverity(envelope *ErrorEnvelope, severity Severity) *ErrorEnvelope {
	return ApplySeverityWithHandling(envelope, severity, nil)
}

// SafeWithContext is a convenience wrapper that uses default error handling
func SafeWithContext(envelope *ErrorEnvelope, context map[string]interface{}) *ErrorEnvelope {
	return ApplyContextWithHandling(envelope, context, nil)
}
