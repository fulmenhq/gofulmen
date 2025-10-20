package logging

import (
	"testing"
)

func TestRedactSecretsMiddleware_APIKeys(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "stripe API key",
			message:  "Using API key sk_live_abcdefghijklmnopqrstuvwxyz",
			expected: "Using API key [REDACTED]",
		},
		{
			name:     "api_key parameter",
			message:  "Request with api_key=sk_test_1234567890abcdef",
			expected: "Request with [REDACTED]",
		},
		{
			name:     "no secrets",
			message:  "Normal log message without secrets",
			expected: "Normal log message without secrets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &LogEvent{
				Service: "test",
				Message: tt.message,
			}

			result := middleware.Process(event)

			if result.Message != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, result.Message)
			}
		})
	}
}

func TestRedactSecretsMiddleware_BearerTokens(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Authorization: Bearer abc123def456ghi789jkl012mno345pqr678stu901vwx234",
	}

	originalMessage := event.Message
	result := middleware.Process(event)

	if result.Message == originalMessage {
		t.Error("Expected bearer token to be redacted")
	}

	if result.Message != "Authorization: [REDACTED]" {
		t.Errorf("Expected 'Authorization: [REDACTED]', got %q", result.Message)
	}
}

func TestRedactSecretsMiddleware_Passwords(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Login failed with password=SuperSecret123",
	}

	result := middleware.Process(event)

	if result.Message != "Login failed with [REDACTED]" {
		t.Errorf("Expected password redacted, got %q", result.Message)
	}
}

func TestRedactSecretsMiddleware_JWTs(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
	}

	result := middleware.Process(event)

	if result.Message != "Token: [REDACTED]" {
		t.Errorf("Expected JWT redacted, got %q", result.Message)
	}
}

func TestRedactSecretsMiddleware_SensitiveFields(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Login attempt",
		Context: map[string]any{
			"username": "alice",
			"password": "SuperSecret123",
			"apiKey":   "sk_live_abcdef",
			"email":    "alice@example.com",
		},
	}

	result := middleware.Process(event)

	if result.Context["username"] != "alice" {
		t.Error("Non-sensitive field should not be redacted")
	}

	if result.Context["password"] != "[REDACTED]" {
		t.Errorf("Password field should be redacted, got %v", result.Context["password"])
	}

	if result.Context["apiKey"] != "[REDACTED]" {
		t.Errorf("apiKey field should be redacted, got %v", result.Context["apiKey"])
	}

	if result.Context["email"] == "[REDACTED]" {
		t.Error("Email should not be redacted by secrets middleware (only PII middleware)")
	}
}

func TestRedactSecretsMiddleware_NestedContext(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Request",
		Context: map[string]any{
			"request": map[string]any{
				"headers": map[string]any{
					"authorization": "Bearer sk_live_secret_token",
				},
				"body": "api_key=my_secret_key_12345",
			},
		},
	}

	result := middleware.Process(event)

	headers := result.Context["request"].(map[string]any)["headers"].(map[string]any)
	if headers["authorization"] != "[REDACTED]" {
		t.Error("Nested authorization header should be redacted")
	}

	body := result.Context["request"].(map[string]any)["body"]
	if body != "[REDACTED]" {
		t.Errorf("Body with api_key should be redacted, got %v", body)
	}
}

func TestRedactSecretsMiddleware_RedactionFlags(t *testing.T) {
	middleware, err := NewRedactSecretsMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactSecretsMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "password=secret123",
	}

	result := middleware.Process(event)

	if len(result.RedactionFlags) != 1 || result.RedactionFlags[0] != "secrets" {
		t.Errorf("Expected RedactionFlags [secrets], got %v", result.RedactionFlags)
	}
}

func TestRedactPIIMiddleware_Emails(t *testing.T) {
	middleware, err := NewRedactPIIMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactPIIMiddleware failed: %v", err)
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "standard email",
			message:  "User alice@example.com logged in",
			expected: "User [REDACTED] logged in",
		},
		{
			name:     "email with subdomain",
			message:  "Contact support@help.company.io",
			expected: "Contact [REDACTED]",
		},
		{
			name:     "multiple emails",
			message:  "CC: alice@foo.com, bob@bar.com",
			expected: "CC: [REDACTED], [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &LogEvent{
				Service: "test",
				Message: tt.message,
			}

			result := middleware.Process(event)

			if result.Message != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Message)
			}
		})
	}
}

func TestRedactPIIMiddleware_PhoneNumbers(t *testing.T) {
	middleware, err := NewRedactPIIMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactPIIMiddleware failed: %v", err)
	}

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "phone with dashes",
			message:  "Call 555-123-4567",
			expected: "Call [REDACTED]",
		},
		{
			name:     "phone with parentheses",
			message:  "Mobile: (555) 123-4567",
			expected: "Mobile: [REDACTED]",
		},
		{
			name:     "phone with spaces",
			message:  "Contact 555 123 4567",
			expected: "Contact [REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &LogEvent{
				Service: "test",
				Message: tt.message,
			}

			result := middleware.Process(event)

			if result.Message != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Message)
			}
		})
	}
}

func TestRedactPIIMiddleware_SSN(t *testing.T) {
	middleware, err := NewRedactPIIMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactPIIMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "SSN: 123-45-6789",
	}

	result := middleware.Process(event)

	if result.Message != "SSN: [REDACTED]" {
		t.Errorf("Expected SSN redacted, got %q", result.Message)
	}
}

func TestRedactPIIMiddleware_RedactionFlags(t *testing.T) {
	middleware, err := NewRedactPIIMiddleware(map[string]any{})
	if err != nil {
		t.Fatalf("NewRedactPIIMiddleware failed: %v", err)
	}

	event := &LogEvent{
		Service: "test",
		Message: "Email: user@example.com",
	}

	result := middleware.Process(event)

	if len(result.RedactionFlags) != 1 || result.RedactionFlags[0] != "pii" {
		t.Errorf("Expected RedactionFlags [pii], got %v", result.RedactionFlags)
	}
}

func TestRedactMiddleware_NilEvent(t *testing.T) {
	secretsMW, _ := NewRedactSecretsMiddleware(map[string]any{})
	piiMW, _ := NewRedactPIIMiddleware(map[string]any{})

	if secretsMW.Process(nil) != nil {
		t.Error("RedactSecretsMiddleware should return nil for nil event")
	}

	if piiMW.Process(nil) != nil {
		t.Error("RedactPIIMiddleware should return nil for nil event")
	}
}

func TestRedactMiddleware_Order(t *testing.T) {
	secretsMW, _ := NewRedactSecretsMiddleware(map[string]any{"order": 20})
	piiMW, _ := NewRedactPIIMiddleware(map[string]any{"order": 25})

	if secretsMW.Order() != 20 {
		t.Errorf("Expected secrets middleware order 20, got %d", secretsMW.Order())
	}

	if piiMW.Order() != 25 {
		t.Errorf("Expected PII middleware order 25, got %d", piiMW.Order())
	}
}

func TestRedactMiddleware_Names(t *testing.T) {
	secretsMW, _ := NewRedactSecretsMiddleware(map[string]any{})
	piiMW, _ := NewRedactPIIMiddleware(map[string]any{})

	if secretsMW.Name() != "redact-secrets" {
		t.Errorf("Expected name 'redact-secrets', got %q", secretsMW.Name())
	}

	if piiMW.Name() != "redact-pii" {
		t.Errorf("Expected name 'redact-pii', got %q", piiMW.Name())
	}
}

func TestRedactMiddleware_Registration(t *testing.T) {
	secretsFactory := DefaultRegistry().factories["redact-secrets"]
	piiFactory := DefaultRegistry().factories["redact-pii"]

	if secretsFactory == nil {
		t.Fatal("redact-secrets should be registered")
	}

	if piiFactory == nil {
		t.Fatal("redact-pii should be registered")
	}

	secretsMW, _ := secretsFactory(map[string]any{})
	if secretsMW.Name() != "redact-secrets" {
		t.Error("Factory should create redact-secrets middleware")
	}

	piiMW, _ := piiFactory(map[string]any{})
	if piiMW.Name() != "redact-pii" {
		t.Error("Factory should create redact-pii middleware")
	}
}

func TestRedactMiddleware_ArrayRedaction(t *testing.T) {
	middleware, _ := NewRedactSecretsMiddleware(map[string]any{})

	event := &LogEvent{
		Service: "test",
		Message: "Request",
		Context: map[string]any{
			"tokens": []any{
				"Bearer valid_token_abc123",
				"public_value",
				"api_key=secret_key_xyz789",
			},
		},
	}

	result := middleware.Process(event)

	tokens := result.Context["tokens"].([]any)
	if tokens[0] != "[REDACTED]" {
		t.Error("Bearer token in array should be redacted")
	}
	if tokens[1] != "public_value" {
		t.Error("Non-secret value should not be redacted")
	}
	if tokens[2] != "[REDACTED]" {
		t.Error("API key in array should be redacted")
	}
}

func TestRedactSecrets_TypedStringSlice(t *testing.T) {
	middleware, _ := NewRedactSecretsMiddleware(map[string]any{})

	event := &LogEvent{
		Service: "test",
		Message: "Request",
		Context: map[string]any{
			"headers": []string{
				"Content-Type: application/json",
				"Authorization: Bearer secret_token_abc123xyz789",
				"User-Agent: TestClient/1.0",
			},
		},
	}

	result := middleware.Process(event)

	headers := result.Context["headers"].([]string)
	if headers[0] != "Content-Type: application/json" {
		t.Error("Non-secret header should not be redacted")
	}
	if headers[1] != "Authorization: [REDACTED]" {
		t.Errorf("Bearer token in []string should be redacted, got %q", headers[1])
	}
	if headers[2] != "User-Agent: TestClient/1.0" {
		t.Error("Non-secret header should not be redacted")
	}
}

func TestRedactSecrets_TypedMapSlice(t *testing.T) {
	middleware, _ := NewRedactSecretsMiddleware(map[string]any{})

	event := &LogEvent{
		Service: "test",
		Message: "Batch request",
		Context: map[string]any{
			"requests": []map[string]any{
				{"url": "/api/users", "method": "GET"},
				{"url": "/api/auth", "token": "secret_abc123"},
			},
		},
	}

	result := middleware.Process(event)

	requests := result.Context["requests"].([]map[string]any)
	if requests[0]["url"] != "/api/users" {
		t.Error("Non-secret field should not be redacted")
	}
	if requests[1]["token"] != "[REDACTED]" {
		t.Errorf("Token in []map[string]any should be redacted, got %v", requests[1]["token"])
	}
}

func TestRedactPII_TypedStringSlice(t *testing.T) {
	middleware, _ := NewRedactPIIMiddleware(map[string]any{})

	event := &LogEvent{
		Service: "test",
		Message: "User data",
		Context: map[string]any{
			"emails": []string{
				"alice@example.com",
				"bob@test.org",
			},
		},
	}

	result := middleware.Process(event)

	emails := result.Context["emails"].([]string)
	if emails[0] != "[REDACTED]" {
		t.Errorf("Email in []string should be redacted, got %q", emails[0])
	}
	if emails[1] != "[REDACTED]" {
		t.Errorf("Email in []string should be redacted, got %q", emails[1])
	}
}

func TestRedactPII_TypedMapSlice(t *testing.T) {
	middleware, _ := NewRedactPIIMiddleware(map[string]any{})

	event := &LogEvent{
		Service: "test",
		Message: "Contact list",
		Context: map[string]any{
			"contacts": []map[string]any{
				{"name": "Alice", "email": "alice@example.com"},
				{"name": "Bob", "phone": "555-123-4567"},
			},
		},
	}

	result := middleware.Process(event)

	contacts := result.Context["contacts"].([]map[string]any)
	if contacts[0]["name"] != "Alice" {
		t.Error("Non-PII field should not be redacted")
	}
	if contacts[0]["email"] != "[REDACTED]" {
		t.Errorf("Email in []map[string]any should be redacted, got %v", contacts[0]["email"])
	}
	if contacts[1]["phone"] != "[REDACTED]" {
		t.Errorf("Phone in []map[string]any should be redacted, got %v", contacts[1]["phone"])
	}
}
