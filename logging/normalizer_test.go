package logging

import (
	"testing"
)

func TestNormalizeLoggerConfig_NilConfig(t *testing.T) {
	_, err := NormalizeLoggerConfig(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if err.Error() != "config cannot be nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNormalizeProfile_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    LoggingProfile
		expected LoggingProfile
		warning  bool
	}{
		{"lowercase simple", "simple", ProfileSimple, true},
		{"mixed case structured", "StRuCtUrEd", ProfileStructured, true},
		{"uppercase enterprise", "ENTERPRISE", ProfileEnterprise, false},
		{"lowercase custom", "custom", ProfileCustom, true},
		{"already normalized", ProfileSimple, ProfileSimple, false},
		{"empty defaults to simple", "", ProfileSimple, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Profile:    tt.input,
				Sinks:      []SinkConfig{{Type: "console", Format: "console"}},
				Middleware: []MiddlewareConfig{{Name: "correlation", Enabled: true}},
				Throttling: &ThrottlingConfig{Enabled: true, MaxRate: 100, BurstSize: 200, DropPolicy: "drop-oldest"},
				PolicyFile: "test.yaml",
			}

			result, err := NormalizeLoggerConfig(config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if config.Profile != tt.expected {
				t.Errorf("expected profile %s, got %s", tt.expected, config.Profile)
			}

			if tt.warning && len(result.Warnings) == 0 {
				t.Error("expected warning for profile normalization")
			}
			if !tt.warning {
				hasProfileWarning := false
				for _, w := range result.Warnings {
					if contains(w, "normalized profile") {
						hasProfileWarning = true
						break
					}
				}
				if hasProfileWarning {
					t.Errorf("unexpected profile normalization warning: %v", result.Warnings)
				}
			}
		})
	}
}

func TestNormalizeProfile_InvalidProfile(t *testing.T) {
	config := &LoggerConfig{
		Profile: "INVALID_PROFILE",
		Sinks:   []SinkConfig{{Type: "console", Format: "console"}},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Profile != "INVALID_PROFILE" {
		t.Errorf("expected invalid profile to remain unchanged, got %s", config.Profile)
	}

	foundWarning := false
	for _, w := range result.Warnings {
		if contains(w, "invalid profile") && contains(w, "will fail schema validation") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("expected warning for invalid profile, got warnings: %v", result.Warnings)
	}
}

func TestNormalizeMiddleware_Deduplication(t *testing.T) {
	config := &LoggerConfig{
		Profile: ProfileStructured,
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100},
			{Name: "redaction", Enabled: true, Order: 200},
			{Name: "correlation", Enabled: false, Order: 150},
		},
		Sinks: []SinkConfig{{Type: "console", Format: "json"}},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Middleware) != 2 {
		t.Errorf("expected 2 middleware after deduplication, got %d", len(config.Middleware))
	}

	if config.Middleware[0].Name != "correlation" || config.Middleware[0].Order != 150 {
		t.Errorf("expected last definition of correlation to win, got Order=%d Enabled=%v", config.Middleware[0].Order, config.Middleware[0].Enabled)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warning for duplicate middleware")
	}
}

func TestNormalizeMiddleware_NilConfig(t *testing.T) {
	config := &LoggerConfig{
		Profile: ProfileStructured,
		Middleware: []MiddlewareConfig{
			{Name: "test", Enabled: true, Config: nil},
		},
		Sinks: []SinkConfig{{Type: "console", Format: "json"}},
	}

	_, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Middleware[0].Config == nil {
		t.Error("expected nil Config to be initialized to empty map")
	}
}

func TestNormalizeThrottling_Defaults(t *testing.T) {
	config := &LoggerConfig{
		Profile:    ProfileEnterprise,
		Throttling: &ThrottlingConfig{Enabled: true},
		Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
		Middleware: []MiddlewareConfig{{Name: "correlation", Enabled: true}},
		PolicyFile: "test.yaml",
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Throttling.MaxRate != 1000 {
		t.Errorf("expected default maxRate 1000, got %d", config.Throttling.MaxRate)
	}

	if config.Throttling.BurstSize != 2000 {
		t.Errorf("expected default burstSize 2000 (2x maxRate), got %d", config.Throttling.BurstSize)
	}

	if config.Throttling.DropPolicy != "drop-oldest" {
		t.Errorf("expected default dropPolicy 'drop-oldest', got %s", config.Throttling.DropPolicy)
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warnings for throttling defaults")
	}
}

func TestNormalizeThrottling_InvalidDropPolicy(t *testing.T) {
	config := &LoggerConfig{
		Profile: ProfileStructured,
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    500,
			BurstSize:  1000,
			DropPolicy: "invalid-policy",
		},
		Sinks: []SinkConfig{{Type: "console", Format: "json"}},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Throttling.DropPolicy != "drop-oldest" {
		t.Errorf("expected invalid dropPolicy to be reset to 'drop-oldest', got %s", config.Throttling.DropPolicy)
	}

	found := false
	for _, w := range result.Warnings {
		if contains(w, "invalid dropPolicy") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for invalid dropPolicy")
	}
}

func TestApplyProfileDefaults_Simple(t *testing.T) {
	config := &LoggerConfig{
		Profile:    ProfileSimple,
		Middleware: []MiddlewareConfig{{Name: "test", Enabled: true}},
		Throttling: &ThrottlingConfig{Enabled: true, MaxRate: 100},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Sinks) == 0 {
		t.Error("expected default console sink for SIMPLE profile")
	}

	if len(config.Middleware) != 0 {
		t.Error("expected middleware to be removed for SIMPLE profile")
	}

	if config.Throttling.Enabled {
		t.Error("expected throttling to be disabled for SIMPLE profile")
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warnings for SIMPLE profile adjustments")
	}
}

func TestApplyProfileDefaults_Structured(t *testing.T) {
	config := &LoggerConfig{
		Profile: ProfileStructured,
		Middleware: []MiddlewareConfig{
			{Name: "mw1", Enabled: true},
			{Name: "mw2", Enabled: true},
			{Name: "mw3", Enabled: true},
		},
		Sinks: []SinkConfig{{Type: "console", Format: "json"}},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Middleware) != 3 {
		t.Error("expected middleware to be preserved for STRUCTURED profile")
	}

	found := false
	for _, w := range result.Warnings {
		if contains(w, "max 2 recommended") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning for exceeding recommended middleware count")
	}
}

func TestApplyProfileDefaults_Enterprise(t *testing.T) {
	tests := []struct {
		name           string
		config         *LoggerConfig
		expectThrottle bool
		expectMW       bool
		expectPolicy   bool
	}{
		{
			name: "no throttling",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Middleware: []MiddlewareConfig{{Name: "correlation", Enabled: true}},
				Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
				PolicyFile: "test.yaml",
			},
			expectThrottle: true,
			expectMW:       false,
			expectPolicy:   false,
		},
		{
			name: "no middleware",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Throttling: &ThrottlingConfig{Enabled: true, MaxRate: 1000, BurstSize: 2000},
				Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
				PolicyFile: "test.yaml",
			},
			expectThrottle: false,
			expectMW:       true,
			expectPolicy:   false,
		},
		{
			name: "no policy",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Throttling: &ThrottlingConfig{Enabled: true, MaxRate: 1000, BurstSize: 2000},
				Middleware: []MiddlewareConfig{{Name: "correlation", Enabled: true}},
				Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
			},
			expectThrottle: false,
			expectMW:       false,
			expectPolicy:   true,
		},
		{
			name: "throttling disabled",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Throttling: &ThrottlingConfig{Enabled: false, MaxRate: 1000, BurstSize: 2000},
				Middleware: []MiddlewareConfig{{Name: "correlation", Enabled: true}},
				Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
				PolicyFile: "test.yaml",
			},
			expectThrottle: true,
			expectMW:       false,
			expectPolicy:   false,
		},
		{
			name: "all middleware disabled",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Throttling: &ThrottlingConfig{Enabled: true, MaxRate: 1000, BurstSize: 2000},
				Middleware: []MiddlewareConfig{
					{Name: "correlation", Enabled: false},
					{Name: "redaction", Enabled: false},
				},
				Sinks:      []SinkConfig{{Type: "console", Format: "json"}},
				PolicyFile: "test.yaml",
			},
			expectThrottle: false,
			expectMW:       true,
			expectPolicy:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeLoggerConfig(tt.config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectThrottle {
				if tt.config.Throttling == nil || !tt.config.Throttling.Enabled {
					t.Error("expected throttling to be enabled")
				}
				foundWarning := false
				for _, w := range result.Warnings {
					if contains(w, "throttling") {
						foundWarning = true
						break
					}
				}
				if !foundWarning {
					t.Error("expected throttling warning")
				}
			}

			if tt.expectMW {
				enabledCount := 0
				for _, mw := range tt.config.Middleware {
					if mw.Enabled {
						enabledCount++
					}
				}
				if enabledCount == 0 {
					t.Error("expected at least one enabled middleware")
				}
				foundWarning := false
				for _, w := range result.Warnings {
					if contains(w, "middleware") {
						foundWarning = true
						break
					}
				}
				if !foundWarning {
					t.Error("expected middleware warning")
				}
			}

			if tt.expectPolicy {
				foundWarning := false
				for _, w := range result.Warnings {
					if contains(w, "policyFile") {
						foundWarning = true
						break
					}
				}
				if !foundWarning {
					t.Error("expected policy warning")
				}
			}
		})
	}
}

func TestApplyProfileDefaults_Custom(t *testing.T) {
	config := &LoggerConfig{
		Profile: ProfileCustom,
		Sinks:   []SinkConfig{{Type: "file", Format: "custom"}},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("expected no warnings for CUSTOM profile, got %d", len(result.Warnings))
	}
}

func TestNormalizeLoggerConfig_RoundTrip(t *testing.T) {
	config := &LoggerConfig{
		Profile:      "structured",
		DefaultLevel: "INFO",
		Service:      "test-service",
		Environment:  "production",
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 100},
			{Name: "redaction", Enabled: true, Order: 200},
		},
		Throttling: &ThrottlingConfig{
			Enabled: true,
		},
		Sinks: []SinkConfig{
			{Type: "console", Format: "json"},
		},
	}

	result, err := NormalizeLoggerConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Profile != ProfileStructured {
		t.Errorf("expected profile STRUCTURED, got %s", config.Profile)
	}

	if config.Throttling.MaxRate != 1000 {
		t.Errorf("expected throttling defaults applied, maxRate=%d", config.Throttling.MaxRate)
	}

	if len(config.Middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(config.Middleware))
	}

	if len(result.Warnings) == 0 {
		t.Error("expected warnings for normalization actions")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
