package httpauth

import (
	"crypto/subtle"
	"strings"
)

// BearerTokenMatches checks whether authHeader contains a Bearer token that matches
// expectedToken using a constant-time comparison on the token portion.
//
// The header scheme is matched case-insensitively ("Bearer", "bearer", etc.).
func BearerTokenMatches(authHeader, expectedToken string) bool {
	if expectedToken == "" {
		return false
	}

	fields := strings.Fields(authHeader)
	if len(fields) != 2 {
		return false
	}

	if !strings.EqualFold(fields[0], "Bearer") {
		return false
	}

	token := fields[1]
	if len(token) != len(expectedToken) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) == 1
}
