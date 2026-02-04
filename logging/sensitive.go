package logging

import "strings"

// IsSensitiveKey reports whether a key name should be treated as sensitive and
// masked in diagnostic output (e.g. envinfo/doctor).
//
// This is a heuristic helper for *display masking* (not deep redaction). For
// log redaction, prefer the redaction middleware.
func IsSensitiveKey(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	// Fast path for common exact field names.
	lower := strings.ToLower(name)
	switch lower {
	case "password", "passwd", "pwd", "secret", "token", "apikey", "api_key", "privatekey", "private_key", "credential", "credentials", "authorization", "auth":
		return true
	}

	upper := strings.ToUpper(name)

	// Common env-var style signals.
	// Keep the list intentionally short to avoid false positives.
	for _, needle := range []string{
		"TOKEN",
		"SECRET",
		"PASSWORD",
		"PASSWD",
		"API_KEY",
		"PRIVATE_KEY",
		"CREDENTIAL",
		"AUTHORIZATION",
	} {
		if strings.Contains(upper, needle) {
			return true
		}
	}

	return false
}
