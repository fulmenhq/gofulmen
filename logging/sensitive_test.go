package logging

import "testing"

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "empty", key: "", want: false},
		{name: "whitespace", key: "   ", want: false},
		{name: "token env", key: "GITHUB_TOKEN", want: true},
		{name: "control plane token", key: "GRONINGEN_CONTROL_PLANE_TOKEN", want: true},
		{name: "password env", key: "DATABASE_PASSWORD", want: true},
		{name: "api key env", key: "API_KEY", want: true},
		{name: "private key env", key: "MY_PRIVATE_KEY", want: true},
		{name: "authorization header", key: "Authorization", want: true},
		{name: "log level", key: "LOG_LEVEL", want: false},
		{name: "port", key: "PORT", want: false},
		{name: "exact token", key: "token", want: true},
		{name: "exact secret", key: "secret", want: true},
		{name: "trim", key: "  github_token  ", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSensitiveKey(tt.key); got != tt.want {
				t.Fatalf("IsSensitiveKey(%q)=%v, want %v", tt.key, got, tt.want)
			}
		})
	}
}
