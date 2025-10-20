package logging

import (
	"testing"
)

func TestGetProfileRequirements(t *testing.T) {
	tests := []struct {
		name                   string
		profile                LoggingProfile
		wantRequiredSinks      int
		wantAllowedFormats     int
		wantMaxMiddleware      *int
		wantMinMiddleware      *int
		wantThrottlingAllowed  bool
		wantThrottlingRequired bool
		wantPolicyEnforcement  bool
	}{
		{
			name:                   "SIMPLE profile",
			profile:                ProfileSimple,
			wantRequiredSinks:      1,
			wantAllowedFormats:     2,
			wantMaxMiddleware:      intPtr(0),
			wantMinMiddleware:      intPtr(0),
			wantThrottlingAllowed:  false,
			wantThrottlingRequired: false,
			wantPolicyEnforcement:  false,
		},
		{
			name:                   "STRUCTURED profile",
			profile:                ProfileStructured,
			wantRequiredSinks:      0,
			wantAllowedFormats:     3,
			wantMaxMiddleware:      intPtr(2),
			wantMinMiddleware:      nil,
			wantThrottlingAllowed:  true,
			wantThrottlingRequired: false,
			wantPolicyEnforcement:  false,
		},
		{
			name:                   "ENTERPRISE profile",
			profile:                ProfileEnterprise,
			wantRequiredSinks:      0,
			wantAllowedFormats:     1,
			wantMaxMiddleware:      nil,
			wantMinMiddleware:      intPtr(1),
			wantThrottlingAllowed:  true,
			wantThrottlingRequired: true,
			wantPolicyEnforcement:  true,
		},
		{
			name:                   "CUSTOM profile",
			profile:                ProfileCustom,
			wantRequiredSinks:      0,
			wantAllowedFormats:     0,
			wantMaxMiddleware:      nil,
			wantMinMiddleware:      nil,
			wantThrottlingAllowed:  true,
			wantThrottlingRequired: false,
			wantPolicyEnforcement:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs := GetProfileRequirements(tt.profile)

			if len(reqs.RequiredSinks) != tt.wantRequiredSinks {
				t.Errorf("RequiredSinks count = %d, want %d", len(reqs.RequiredSinks), tt.wantRequiredSinks)
			}

			if len(reqs.AllowedFormats) != tt.wantAllowedFormats {
				t.Errorf("AllowedFormats count = %d, want %d", len(reqs.AllowedFormats), tt.wantAllowedFormats)
			}

			if !intPtrEqual(reqs.MaxMiddleware, tt.wantMaxMiddleware) {
				t.Errorf("MaxMiddleware = %v, want %v", formatIntPtr(reqs.MaxMiddleware), formatIntPtr(tt.wantMaxMiddleware))
			}

			if !intPtrEqual(reqs.MinMiddleware, tt.wantMinMiddleware) {
				t.Errorf("MinMiddleware = %v, want %v", formatIntPtr(reqs.MinMiddleware), formatIntPtr(tt.wantMinMiddleware))
			}

			if reqs.ThrottlingAllowed != tt.wantThrottlingAllowed {
				t.Errorf("ThrottlingAllowed = %v, want %v", reqs.ThrottlingAllowed, tt.wantThrottlingAllowed)
			}

			if reqs.ThrottlingRequired != tt.wantThrottlingRequired {
				t.Errorf("ThrottlingRequired = %v, want %v", reqs.ThrottlingRequired, tt.wantThrottlingRequired)
			}

			if reqs.PolicyEnforcement != tt.wantPolicyEnforcement {
				t.Errorf("PolicyEnforcement = %v, want %v", reqs.PolicyEnforcement, tt.wantPolicyEnforcement)
			}
		})
	}
}

func TestValidateProfileRequirements_SIMPLE(t *testing.T) {
	tests := []struct {
		name              string
		sinks             []SinkConfig
		middleware        []MiddlewareConfig
		format            string
		throttlingEnabled bool
		policyEnabled     bool
		wantErrorCount    int
	}{
		{
			name: "valid SIMPLE configuration",
			sinks: []SinkConfig{
				{Type: "console", Format: "console"},
			},
			middleware:        []MiddlewareConfig{},
			format:            "console",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    0,
		},
		{
			name: "SIMPLE with json format fails",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware:        []MiddlewareConfig{},
			format:            "json",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    2, // Both global format and sink format violations
		},
		{
			name:              "SIMPLE missing console sink",
			sinks:             []SinkConfig{{Type: "file"}},
			middleware:        []MiddlewareConfig{},
			format:            "console",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    1,
		},
		{
			name: "SIMPLE with middleware fails",
			sinks: []SinkConfig{
				{Type: "console", Format: "console"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "console",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    1,
		},
		{
			name: "SIMPLE with throttling fails",
			sinks: []SinkConfig{
				{Type: "console", Format: "console"},
			},
			middleware:        []MiddlewareConfig{},
			format:            "console",
			throttlingEnabled: true,
			policyEnabled:     false,
			wantErrorCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProfileRequirements(
				ProfileSimple,
				tt.sinks,
				tt.middleware,
				tt.format,
				tt.throttlingEnabled,
				tt.policyEnabled,
			)

			if len(errors) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d", len(errors), tt.wantErrorCount)
				for i, err := range errors {
					t.Logf("  error[%d]: %v", i, err)
				}
			}
		})
	}
}

func TestValidateProfileRequirements_STRUCTURED(t *testing.T) {
	tests := []struct {
		name              string
		sinks             []SinkConfig
		middleware        []MiddlewareConfig
		format            string
		throttlingEnabled bool
		policyEnabled     bool
		wantErrorCount    int
	}{
		{
			name: "valid STRUCTURED configuration",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    0,
		},
		{
			name: "STRUCTURED with 2 middleware ok",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
				{Name: "redact-secrets", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: true,
			policyEnabled:     false,
			wantErrorCount:    0,
		},
		{
			name: "STRUCTURED with 3 middleware fails",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
				{Name: "redact-secrets", Enabled: true},
				{Name: "redact-pii", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProfileRequirements(
				ProfileStructured,
				tt.sinks,
				tt.middleware,
				tt.format,
				tt.throttlingEnabled,
				tt.policyEnabled,
			)

			if len(errors) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d", len(errors), tt.wantErrorCount)
				for i, err := range errors {
					t.Logf("  error[%d]: %v", i, err)
				}
			}
		})
	}
}

func TestValidateProfileRequirements_ENTERPRISE(t *testing.T) {
	tests := []struct {
		name              string
		sinks             []SinkConfig
		middleware        []MiddlewareConfig
		format            string
		throttlingEnabled bool
		policyEnabled     bool
		wantErrorCount    int
	}{
		{
			name: "valid ENTERPRISE configuration",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: true,
			policyEnabled:     true,
			wantErrorCount:    0,
		},
		{
			name: "ENTERPRISE missing middleware",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware:        []MiddlewareConfig{},
			format:            "json",
			throttlingEnabled: true,
			policyEnabled:     true,
			wantErrorCount:    1,
		},
		{
			name: "ENTERPRISE missing throttling",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: false,
			policyEnabled:     true,
			wantErrorCount:    1,
		},
		{
			name: "ENTERPRISE missing policy",
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "json",
			throttlingEnabled: true,
			policyEnabled:     false,
			wantErrorCount:    1,
		},
		{
			name: "ENTERPRISE with text format fails",
			sinks: []SinkConfig{
				{Type: "console", Format: "text"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			format:            "text",
			throttlingEnabled: true,
			policyEnabled:     true,
			wantErrorCount:    2, // Both global format and sink format violations
		},
		{
			name: "ENTERPRISE multiple violations",
			sinks: []SinkConfig{
				{Type: "console", Format: "text"},
			},
			middleware:        []MiddlewareConfig{},
			format:            "text",
			throttlingEnabled: false,
			policyEnabled:     false,
			wantErrorCount:    5, // format (global) + format (sink) + middleware + throttling + policy
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProfileRequirements(
				ProfileEnterprise,
				tt.sinks,
				tt.middleware,
				tt.format,
				tt.throttlingEnabled,
				tt.policyEnabled,
			)

			if len(errors) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d", len(errors), tt.wantErrorCount)
				for i, err := range errors {
					t.Logf("  error[%d]: %v", i, err)
				}
			}
		})
	}
}

func TestValidateProfileRequirements_PerSinkFormats(t *testing.T) {
	tests := []struct {
		name              string
		profile           LoggingProfile
		sinks             []SinkConfig
		middleware        []MiddlewareConfig
		throttlingEnabled bool
		policyEnabled     bool
		wantErrorCount    int
	}{
		{
			name:    "SIMPLE with console format on sink - valid",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console", Format: "console"},
			},
			wantErrorCount: 0,
		},
		{
			name:    "SIMPLE with json format on sink - invalid",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
			},
			wantErrorCount: 1,
		},
		{
			name:    "SIMPLE with mixed formats - one invalid",
			profile: ProfileSimple,
			sinks: []SinkConfig{
				{Type: "console", Format: "console"},
				{Type: "file", Format: "json"},
			},
			wantErrorCount: 1,
		},
		{
			name:    "ENTERPRISE with json on all sinks - valid",
			profile: ProfileEnterprise,
			sinks: []SinkConfig{
				{Type: "console", Format: "json"},
				{Type: "file", Format: "json"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			throttlingEnabled: true,
			policyEnabled:     true,
			wantErrorCount:    0,
		},
		{
			name:    "ENTERPRISE with text format - invalid",
			profile: ProfileEnterprise,
			sinks: []SinkConfig{
				{Type: "console", Format: "text"},
			},
			middleware: []MiddlewareConfig{
				{Name: "correlation", Enabled: true},
			},
			throttlingEnabled: true,
			policyEnabled:     true,
			wantErrorCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateProfileRequirements(
				tt.profile,
				tt.sinks,
				tt.middleware,
				"",
				tt.throttlingEnabled,
				tt.policyEnabled,
			)

			if len(errors) != tt.wantErrorCount {
				t.Errorf("got %d errors, want %d", len(errors), tt.wantErrorCount)
				for i, err := range errors {
					t.Logf("  error[%d]: %v", i, err)
				}
			}
		})
	}
}

func TestValidateProfileRequirements_CUSTOM(t *testing.T) {
	sinks := []SinkConfig{
		{Type: "custom-sink", Format: "custom-format"},
	}
	middleware := []MiddlewareConfig{
		{Name: "custom1", Enabled: true},
		{Name: "custom2", Enabled: true},
		{Name: "custom3", Enabled: true},
	}

	errors := ValidateProfileRequirements(
		ProfileCustom,
		sinks,
		middleware,
		"custom-format",
		true,
		true,
	)

	if len(errors) != 0 {
		t.Errorf("CUSTOM profile should allow any configuration, got %d errors", len(errors))
		for i, err := range errors {
			t.Logf("  error[%d]: %v", i, err)
		}
	}
}

func intPtr(i int) *int {
	return &i
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func formatIntPtr(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p + '0'))
}
