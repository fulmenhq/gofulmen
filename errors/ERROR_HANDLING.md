# Error Handling Guidelines

This document provides guidelines for handling errors when using the Fulmen error envelope system, particularly when working with `WithSeverity()` and `WithContext()` methods that return errors.

## Overview

The error envelope system provides structured error handling with validation for severity levels and context data. Both `WithSeverity()` and `WithContext()` methods return errors when validation fails, ensuring data integrity and schema compliance.

## Error Handling Strategies

The `errors` package provides several strategies for handling validation errors:

### 1. StrategyLogWarning (Default)

- Logs validation errors as warnings
- Continues execution with the original envelope
- Best for: Development and debugging

### 2. StrategyAppendToMessage

- Appends validation errors to the envelope message
- Preserves error information in a user-visible way
- Best for: Production environments where error visibility is important

### 3. StrategyFailFast

- Logs validation errors and continues
- Provides clear error logging for monitoring
- Best for: Critical systems where errors need visibility

### 4. StrategySilent

- Completely ignores validation errors
- Use with caution - errors are lost
- Best for: High-performance scenarios where validation is optional

## Helper Functions

### SafeWithSeverity

```go
envelope := errors.SafeWithSeverity(envelope, errors.SeverityHigh)
```

- Uses default error handling (StrategyLogWarning)
- Safe for production use
- Logs warnings if validation fails

### SafeWithContext

```go
envelope := errors.SafeWithContext(envelope, map[string]interface{}{
    "component": "my-service",
    "operation": "process",
})
```

- Uses default error handling (StrategyAppendToMessage)
- Safe for production use
- Appends errors to message if validation fails

### Custom Error Handling

```go
config := &errors.ErrorHandlingConfig{
    SeverityStrategy: errors.StrategyAppendToMessage,
    ContextStrategy:  errors.StrategyLogWarning,
    Logger:           myCustomLogger,
}

envelope := errors.ApplySeverityWithHandling(envelope, errors.SeverityHigh, config)
envelope := errors.ApplyContextWithHandling(envelope, context, config)
```

## Best Practices

### 1. Use Safe Helpers for Common Cases

```go
// Recommended: Use safe helpers
envelope := errors.NewErrorEnvelope("ERROR_CODE", "Error message")
envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
envelope = errors.SafeWithContext(envelope, context)

// Not recommended: Ignoring errors
envelope, _ := envelope.WithSeverity(errors.SeverityHigh) // Error lost!
envelope, _ := envelope.WithContext(context)                // Error lost!
```

### 2. Handle Errors Explicitly in Critical Paths

```go
envelope, err := envelope.WithSeverity(severity)
if err != nil {
    // Log the error for monitoring
    logger.Printf("Failed to set severity: %v", err)
    // Decide how to proceed - maybe use default severity
    envelope, _ = envelope.WithSeverity(errors.SeverityInfo)
}
```

### 3. Use Custom Config for Specific Needs

```go
// For high-volume services where performance matters
config := &errors.ErrorHandlingConfig{
    SeverityStrategy: errors.StrategySilent,
    ContextStrategy:  errors.StrategySilent,
}

// For development/debugging
config := &errors.ErrorHandlingConfig{
    SeverityStrategy: errors.StrategyLogWarning,
    ContextStrategy:  errors.StrategyLogWarning,
    Logger:           developmentLogger,
}
```

### 4. Document Error Handling Choices

```go
// This service uses silent error handling for performance reasons.
// Validation errors are logged but don't block error creation.
// Monitor logs for validation issues.
func createErrorEnvelope(code, message string) *errors.ErrorEnvelope {
    envelope := errors.NewErrorEnvelope(code, message)
    return errors.SafeWithSeverity(envelope, errors.SeverityMedium)
}
```

## Common Patterns

### Pattern 1: Validation Wrapper (Current Implementation)

```go
func (v *ErrorEnvelope) ValidateDataWithEnvelope(data interface{}, correlationID string) (*errors.ErrorEnvelope, error) {
    envelope := errors.NewErrorEnvelope("VALIDATION_ERROR", "Validation failed")
    envelope, severityErr := envelope.WithSeverity(errors.SeverityMedium)
    if severityErr != nil {
        // Log and continue with original envelope
        fmt.Printf("Warning: failed to set severity: %v\n", severityErr)
    }
    // Continue with envelope creation...
}
```

### Pattern 2: Safe Wrapper Function

```go
func CreateValidationError(correlationID string) *errors.ErrorEnvelope {
    envelope := errors.NewErrorEnvelope("VALIDATION_ERROR", "Validation failed")
    envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
    envelope = envelope.WithCorrelationID(correlationID)
    envelope = errors.SafeWithContext(envelope, validationContext)
    return envelope
}
```

### Pattern 3: Custom Error Handling

```go
func CreateErrorWithCustomHandling(code, message string, severity errors.Severity) *errors.ErrorEnvelope {
    envelope := errors.NewErrorEnvelope(code, message)

    config := &errors.ErrorHandlingConfig{
        SeverityStrategy: errors.StrategyAppendToMessage,
        ContextStrategy:  errors.StrategyLogWarning,
    }

    envelope = errors.ApplySeverityWithHandling(envelope, severity, config)
    envelope = errors.ApplyContextWithHandling(envelope, context, config)

    return envelope
}
```

## Migration Guide

### From Ignoring Errors

```go
// Before: Errors ignored
envelope, _ := envelope.WithSeverity(errors.SeverityHigh)
envelope, _ := envelope.WithContext(context)

// After: Use safe helpers
envelope := errors.SafeWithSeverity(envelope, errors.SeverityHigh)
envelope := errors.SafeWithContext(envelope, context)
```

### From Custom Error Handling

```go
// Before: Custom error handling scattered throughout code
envelope, err := envelope.WithSeverity(severity)
if err != nil {
    log.Printf("Severity error: %v", err)
    envelope.Message = fmt.Sprintf("%s (severity: %v)", envelope.Message, err)
}

// After: Centralized error handling
config := &errors.ErrorHandlingConfig{
    SeverityStrategy: errors.StrategyAppendToMessage,
    ContextStrategy:  errors.StrategyLogWarning,
}
envelope := errors.ApplySeverityWithHandling(envelope, severity, config)
```

## Performance Considerations

- **SafeWithSeverity/SafeWithContext**: Minimal overhead, suitable for high-frequency operations
- **Custom Config with StrategySilent**: Zero overhead when validation errors are not critical
- **StrategyAppendToMessage**: Slight overhead due to string concatenation
- **StrategyLogWarning**: I/O overhead from logging (use sparingly in hot paths)

## Monitoring and Observability

### Key Metrics to Monitor:

1. **Severity validation failures** - Indicates misconfigured severity levels
2. **Context validation failures** - Indicates malformed context data
3. **Error handling strategy usage** - Helps optimize error handling patterns

### Logging Best Practices:

```go
// Include correlation IDs in error logs
logger.Printf("[%s] Severity validation failed: %v", correlationID, err)

// Use structured logging for better analysis
logger.WithFields(logrus.Fields{
    "correlation_id": correlationID,
    "error_type":   "severity_validation",
    "severity":     requestedSeverity,
}).Errorf("Severity validation failed: %v", err)
```

## Summary

1. **Never ignore errors** from `WithSeverity()` or `WithContext()`
2. **Use safe helpers** (`SafeWithSeverity`, `SafeWithContext`) for most cases
3. **Choose appropriate strategies** based on your use case
4. **Document your error handling choices** for maintainability
5. **Monitor validation failures** to catch configuration issues early
6. **Test error handling paths** to ensure they work as expected

The error handling system provides flexibility while ensuring data integrity. Choose the approach that best fits your specific needs and performance requirements.
