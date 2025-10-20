package logging

import (
	"encoding/json"
	"testing"
)

func TestLoggerConfig_ProfileField(t *testing.T) {
	tests := []struct {
		name        string
		config      LoggerConfig
		wantProfile LoggingProfile
	}{
		{
			name: "profile set to SIMPLE",
			config: LoggerConfig{
				Profile: ProfileSimple,
				Service: "test",
			},
			wantProfile: ProfileSimple,
		},
		{
			name: "profile set to STRUCTURED",
			config: LoggerConfig{
				Profile: ProfileStructured,
				Service: "test",
			},
			wantProfile: ProfileStructured,
		},
		{
			name: "profile set to ENTERPRISE",
			config: LoggerConfig{
				Profile: ProfileEnterprise,
				Service: "test",
			},
			wantProfile: ProfileEnterprise,
		},
		{
			name: "profile set to CUSTOM",
			config: LoggerConfig{
				Profile: ProfileCustom,
				Service: "test",
			},
			wantProfile: ProfileCustom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Profile != tt.wantProfile {
				t.Errorf("Profile = %v, want %v", tt.config.Profile, tt.wantProfile)
			}
		})
	}
}

func TestLoggerConfig_MiddlewareField(t *testing.T) {
	config := LoggerConfig{
		Service: "test",
		Middleware: []MiddlewareConfig{
			{Name: "correlation", Enabled: true, Order: 5},
			{Name: "redact-secrets", Enabled: true, Order: 10},
		},
	}

	if len(config.Middleware) != 2 {
		t.Errorf("expected 2 middleware, got %d", len(config.Middleware))
	}

	if config.Middleware[0].Name != "correlation" {
		t.Errorf("expected first middleware name 'correlation', got %s", config.Middleware[0].Name)
	}

	if config.Middleware[0].Order != 5 {
		t.Errorf("expected first middleware order 5, got %d", config.Middleware[0].Order)
	}
}

func TestLoggerConfig_ThrottlingField(t *testing.T) {
	config := LoggerConfig{
		Service: "test",
		Throttling: &ThrottlingConfig{
			Enabled:    true,
			MaxRate:    1000,
			BurstSize:  100,
			WindowSize: 60,
			DropPolicy: "drop-oldest",
		},
	}

	if config.Throttling == nil {
		t.Fatal("expected throttling config, got nil")
	}

	if !config.Throttling.Enabled {
		t.Error("expected throttling enabled")
	}

	if config.Throttling.MaxRate != 1000 {
		t.Errorf("expected maxRate 1000, got %d", config.Throttling.MaxRate)
	}

	if config.Throttling.DropPolicy != "drop-oldest" {
		t.Errorf("expected dropPolicy 'drop-oldest', got %s", config.Throttling.DropPolicy)
	}
}

func TestLoggerConfig_PolicyFileField(t *testing.T) {
	config := LoggerConfig{
		Service:    "test",
		PolicyFile: "/etc/fulmen/logging-policy.yaml",
	}

	if config.PolicyFile != "/etc/fulmen/logging-policy.yaml" {
		t.Errorf("expected policy file path, got %s", config.PolicyFile)
	}
}

func TestLoadConfig_WithProfile(t *testing.T) {
	t.Skip("Schema validation for profile field pending Crucible schema update (Phase 6)")
}

func TestLoadConfig_DefaultProfile(t *testing.T) {
	t.Skip("Schema validation for profile field pending Crucible schema update (Phase 6)")
}

func TestApplyDefaults_SinkFormat(t *testing.T) {
	tests := []struct {
		name        string
		profile     LoggingProfile
		sinks       []SinkConfig
		wantFormats []string
	}{
		{
			name:    "SIMPLE profile defaults to console format",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console"},
			},
			wantFormats: []string{"console"},
		},
		{
			name:    "STRUCTURED profile defaults to json format",
			profile: ProfileStructured,
			sinks: []SinkConfig{
				{Type: "console"},
			},
			wantFormats: []string{"json"},
		},
		{
			name:    "ENTERPRISE profile defaults to json format",
			profile: ProfileEnterprise,
			sinks: []SinkConfig{
				{Type: "console"},
			},
			wantFormats: []string{"json"},
		},
		{
			name:    "Explicit format preserved",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console", Format: "text"},
			},
			wantFormats: []string{"text"},
		},
		{
			name:    "Multiple sinks get profile-appropriate defaults",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console"},
				{Type: "file"},
			},
			wantFormats: []string{"console", "console"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggerConfig{
				Profile: tt.profile,
				Service: "test",
				Sinks:   tt.sinks,
			}
			applyDefaults(&config)

			if len(config.Sinks) != len(tt.wantFormats) {
				t.Fatalf("expected %d sinks, got %d", len(tt.wantFormats), len(config.Sinks))
			}

			for i, wantFormat := range tt.wantFormats {
				if config.Sinks[i].Format != wantFormat {
					t.Errorf("sink[%d] format = %s, want %s", i, config.Sinks[i].Format, wantFormat)
				}
			}
		})
	}
}

func TestApplyDefaults_Profile(t *testing.T) {
	tests := []struct {
		name        string
		input       LoggerConfig
		wantProfile LoggingProfile
	}{
		{
			name:        "empty profile gets SIMPLE default",
			input:       LoggerConfig{Service: "test"},
			wantProfile: ProfileSimple,
		},
		{
			name: "explicit profile preserved",
			input: LoggerConfig{
				Profile: ProfileEnterprise,
				Service: "test",
			},
			wantProfile: ProfileEnterprise,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			applyDefaults(&config)

			if config.Profile != tt.wantProfile {
				t.Errorf("Profile = %v, want %v", config.Profile, tt.wantProfile)
			}
		})
	}
}

func TestDefaultConfig_AlignedWithSimpleProfile(t *testing.T) {
	config := DefaultConfig("test-service")

	// Should have SIMPLE profile
	if config.Profile != ProfileSimple {
		t.Errorf("Profile = %v, want %v", config.Profile, ProfileSimple)
	}

	// Apply defaults (simulating what New() does)
	applyDefaults(config)

	// Should have console sink with console format (SIMPLE default)
	if len(config.Sinks) != 1 {
		t.Fatalf("expected 1 sink, got %d", len(config.Sinks))
	}

	sink := config.Sinks[0]
	if sink.Type != "console" {
		t.Errorf("sink type = %s, want console", sink.Type)
	}

	if sink.Format != "console" {
		t.Errorf("sink format = %s, want console (SIMPLE profile default)", sink.Format)
	}

	// Validate against SIMPLE profile requirements
	errors := ValidateProfileRequirements(
		config.Profile,
		config.Sinks,
		config.Middleware,
		"",
		false,
		false,
	)

	if len(errors) > 0 {
		t.Errorf("DefaultConfig should pass SIMPLE profile validation, got %d errors:", len(errors))
		for i, err := range errors {
			t.Logf("  error[%d]: %v", i, err)
		}
	}
}

func TestMiddlewareConfig_JSON(t *testing.T) {
	mw := MiddlewareConfig{
		Name:    "correlation",
		Enabled: true,
		Order:   5,
		Config: map[string]any{
			"generator": "uuid-v7",
		},
	}

	data, err := json.Marshal(mw)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded MiddlewareConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != mw.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, mw.Name)
	}

	if decoded.Enabled != mw.Enabled {
		t.Errorf("Enabled = %v, want %v", decoded.Enabled, mw.Enabled)
	}

	if decoded.Order != mw.Order {
		t.Errorf("Order = %d, want %d", decoded.Order, mw.Order)
	}
}

func TestThrottlingConfig_JSON(t *testing.T) {
	throttle := ThrottlingConfig{
		Enabled:    true,
		MaxRate:    1000,
		BurstSize:  100,
		WindowSize: 60,
		DropPolicy: "drop-oldest",
	}

	data, err := json.Marshal(throttle)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded ThrottlingConfig
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Enabled != throttle.Enabled {
		t.Errorf("Enabled = %v, want %v", decoded.Enabled, throttle.Enabled)
	}

	if decoded.MaxRate != throttle.MaxRate {
		t.Errorf("MaxRate = %d, want %d", decoded.MaxRate, throttle.MaxRate)
	}

	if decoded.DropPolicy != throttle.DropPolicy {
		t.Errorf("DropPolicy = %s, want %s", decoded.DropPolicy, throttle.DropPolicy)
	}
}
