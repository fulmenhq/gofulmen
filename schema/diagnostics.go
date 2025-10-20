package schema

import (
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SeverityLevel represents the diagnostic severity.
type SeverityLevel string

const (
	// SeverityError indicates a validation failure.
	SeverityError SeverityLevel = "ERROR"
	// SeverityWarn indicates a non-fatal warning.
	SeverityWarn SeverityLevel = "WARN"

	sourceGoFulmen = "gofulmen"
	sourceGoneat   = "goneat"
)

// Diagnostic captures a validation or schema compilation diagnostic.
type Diagnostic struct {
	Pointer  string        `json:"pointer"`
	Keyword  string        `json:"keyword"`
	Message  string        `json:"message"`
	Severity SeverityLevel `json:"severity"`
	Source   string        `json:"source"`
}

// DiagnosticsToValidationErrors converts diagnostics into ValidationErrors (for legacy callers).
func DiagnosticsToValidationErrors(diags []Diagnostic) ValidationErrors {
	if len(diags) == 0 {
		return nil
	}

	errs := make(ValidationErrors, 0, len(diags))
	for _, d := range diags {
		if d.Severity != SeverityError {
			continue
		}
		field := d.Pointer
		if field == "" {
			field = d.Keyword
		}
		errs = append(errs, NewValidationError(field, d.Message, nil))
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

func diagnosticsFromValidationError(err *jsonschema.ValidationError, source string) []Diagnostic {
	if err == nil {
		return nil
	}

	var diags []Diagnostic
	stack := []*jsonschema.ValidationError{err}
	for len(stack) > 0 {
		current := stack[0]
		stack = stack[1:]

		diags = append(diags, Diagnostic{
			Pointer:  current.InstanceLocation,
			Keyword:  trimKeyword(current.KeywordLocation),
			Message:  current.Message,
			Severity: SeverityError,
			Source:   source,
		})

		stack = append(stack, current.Causes...)
	}
	return diags
}

func trimKeyword(keyword string) string {
	if idx := strings.IndexRune(keyword, '#'); idx >= 0 {
		return keyword[idx+1:]
	}
	return keyword
}
