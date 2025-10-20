package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicy_FromPath(t *testing.T) {
	policyPath := filepath.Join("testdata", "policies", "permissive-policy.yaml")

	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadPolicy failed: %v", err)
	}

	if len(policy.AllowedProfiles) != 4 {
		t.Errorf("expected 4 allowed profiles, got %d", len(policy.AllowedProfiles))
	}

	if !policy.AuditSettings.LogPolicyViolations {
		t.Error("expected LogPolicyViolations to be true")
	}

	if policy.AuditSettings.EnforceStrictMode {
		t.Error("expected EnforceStrictMode to be false for permissive policy")
	}
}

func TestLoadPolicy_StrictPolicy(t *testing.T) {
	policyPath := filepath.Join("testdata", "policies", "strict-enterprise-policy.yaml")

	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadPolicy failed: %v", err)
	}

	if len(policy.AllowedProfiles) != 1 {
		t.Errorf("expected 1 allowed profile, got %d", len(policy.AllowedProfiles))
	}

	if policy.AllowedProfiles[0] != ProfileEnterprise {
		t.Errorf("expected ENTERPRISE profile, got %s", policy.AllowedProfiles[0])
	}

	if !policy.AuditSettings.EnforceStrictMode {
		t.Error("expected EnforceStrictMode to be true for strict policy")
	}

	enterpriseReqs, ok := policy.ProfileRequirements[ProfileEnterprise]
	if !ok {
		t.Fatal("expected ENTERPRISE profile requirements")
	}

	if enterpriseReqs.MinEnvironment != "staging" {
		t.Errorf("expected minEnvironment 'staging', got '%s'", enterpriseReqs.MinEnvironment)
	}

	if len(enterpriseReqs.RequiredFeatures) != 3 {
		t.Errorf("expected 3 required features, got %d", len(enterpriseReqs.RequiredFeatures))
	}
}

func TestLoadPolicy_SearchPaths(t *testing.T) {
	tempDir := t.TempDir()

	fulmenDir := filepath.Join(tempDir, ".fulmen")
	if err := os.MkdirAll(fulmenDir, 0755); err != nil {
		t.Fatalf("failed to create .fulmen dir: %v", err)
	}

	policyPath := filepath.Join(fulmenDir, "logging-policy.yaml")
	policyContent := `
allowedProfiles:
  - SIMPLE
auditSettings:
  logPolicyViolations: false
  enforceStrictMode: false
  requirePolicyFile: false
`
	if err := os.WriteFile(policyPath, []byte(policyContent), 0644); err != nil {
		t.Fatalf("failed to write policy: %v", err)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	})

	policy, err := LoadPolicy("nonexistent.yaml")
	if err != nil {
		t.Fatalf("LoadPolicy should find policy in .fulmen/: %v", err)
	}

	if len(policy.AllowedProfiles) != 1 || policy.AllowedProfiles[0] != ProfileSimple {
		t.Errorf("policy not loaded correctly from .fulmen/")
	}
}

func TestLoadPolicy_NotFound(t *testing.T) {
	_, err := LoadPolicy("nonexistent-file-that-does-not-exist.yaml")
	if err == nil {
		t.Error("expected error for nonexistent policy file")
	}
}

func TestLoadPolicy_OrgPath(t *testing.T) {
	tempDir := t.TempDir()

	orgPolicyPath := filepath.Join(tempDir, "logging-policy.yaml")
	policyContent := `
allowedProfiles:
  - STRUCTURED
  - ENTERPRISE
auditSettings:
  logPolicyViolations: true
  enforceStrictMode: false
  requirePolicyFile: false
`
	if err := os.WriteFile(orgPolicyPath, []byte(policyContent), 0644); err != nil {
		t.Fatalf("failed to write policy: %v", err)
	}

	originalEnv := os.Getenv("FULMEN_ORG_PATH")
	t.Cleanup(func() {
		if originalEnv != "" {
			if err := os.Setenv("FULMEN_ORG_PATH", originalEnv); err != nil {
				t.Errorf("failed to restore FULMEN_ORG_PATH: %v", err)
			}
		} else {
			if err := os.Unsetenv("FULMEN_ORG_PATH"); err != nil {
				t.Errorf("failed to unset FULMEN_ORG_PATH: %v", err)
			}
		}
	})

	if err := os.Setenv("FULMEN_ORG_PATH", tempDir); err != nil {
		t.Fatalf("failed to set FULMEN_ORG_PATH: %v", err)
	}

	policy, err := LoadPolicy("nonexistent.yaml")
	if err != nil {
		t.Fatalf("LoadPolicy should find policy via FULMEN_ORG_PATH: %v", err)
	}

	if len(policy.AllowedProfiles) != 2 {
		t.Errorf("expected 2 allowed profiles, got %d", len(policy.AllowedProfiles))
	}

	if policy.AllowedProfiles[0] != ProfileStructured || policy.AllowedProfiles[1] != ProfileEnterprise {
		t.Errorf("policy not loaded correctly from FULMEN_ORG_PATH")
	}
}

func TestValidateConfigAgainstPolicy_AllowedProfiles(t *testing.T) {
	policy := &LoggingPolicy{
		AllowedProfiles: []LoggingProfile{ProfileStructured, ProfileEnterprise},
	}

	tests := []struct {
		name           string
		profile        LoggingProfile
		wantViolations int
	}{
		{
			name:           "STRUCTURED allowed",
			profile:        ProfileStructured,
			wantViolations: 0,
		},
		{
			name:           "ENTERPRISE allowed",
			profile:        ProfileEnterprise,
			wantViolations: 0,
		},
		{
			name:           "SIMPLE not allowed",
			profile:        ProfileSimple,
			wantViolations: 1,
		},
		{
			name:           "CUSTOM not allowed",
			profile:        ProfileCustom,
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Profile: tt.profile,
				Service: "test",
			}

			violations := ValidateConfigAgainstPolicy(config, policy, "", "")

			if len(violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantViolations)
				for i, v := range violations {
					t.Logf("  violation[%d]: %v", i, v)
				}
			}
		})
	}
}

func TestValidateConfigAgainstPolicy_RequiredProfiles(t *testing.T) {
	policy := &LoggingPolicy{
		RequiredProfiles: map[string][]LoggingProfile{
			"api-service": {ProfileEnterprise},
			"cli-tool":    {ProfileSimple, ProfileStructured},
		},
	}

	tests := []struct {
		name           string
		appType        string
		profile        LoggingProfile
		wantViolations int
	}{
		{
			name:           "api-service with ENTERPRISE - valid",
			appType:        "api-service",
			profile:        ProfileEnterprise,
			wantViolations: 0,
		},
		{
			name:           "api-service with SIMPLE - invalid",
			appType:        "api-service",
			profile:        ProfileSimple,
			wantViolations: 1,
		},
		{
			name:           "cli-tool with SIMPLE - valid",
			appType:        "cli-tool",
			profile:        ProfileSimple,
			wantViolations: 0,
		},
		{
			name:           "cli-tool with STRUCTURED - valid",
			appType:        "cli-tool",
			profile:        ProfileStructured,
			wantViolations: 0,
		},
		{
			name:           "cli-tool with ENTERPRISE - invalid",
			appType:        "cli-tool",
			profile:        ProfileEnterprise,
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Profile: tt.profile,
				Service: "test",
			}

			violations := ValidateConfigAgainstPolicy(config, policy, "", tt.appType)

			if len(violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantViolations)
				for i, v := range violations {
					t.Logf("  violation[%d]: %v", i, v)
				}
			}
		})
	}
}

func TestValidateConfigAgainstPolicy_EnvironmentRules(t *testing.T) {
	policy := &LoggingPolicy{
		EnvironmentRules: map[string][]LoggingProfile{
			"production":  {ProfileEnterprise},
			"staging":     {ProfileStructured, ProfileEnterprise},
			"development": {ProfileSimple, ProfileStructured},
		},
	}

	tests := []struct {
		name           string
		environment    string
		profile        LoggingProfile
		wantViolations int
	}{
		{
			name:           "production with ENTERPRISE - valid",
			environment:    "production",
			profile:        ProfileEnterprise,
			wantViolations: 0,
		},
		{
			name:           "production with SIMPLE - invalid",
			environment:    "production",
			profile:        ProfileSimple,
			wantViolations: 1,
		},
		{
			name:           "staging with STRUCTURED - valid",
			environment:    "staging",
			profile:        ProfileStructured,
			wantViolations: 0,
		},
		{
			name:           "staging with SIMPLE - invalid",
			environment:    "staging",
			profile:        ProfileSimple,
			wantViolations: 1,
		},
		{
			name:           "development with SIMPLE - valid",
			environment:    "development",
			profile:        ProfileSimple,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{
				Profile: tt.profile,
				Service: "test",
			}

			violations := ValidateConfigAgainstPolicy(config, policy, tt.environment, "")

			if len(violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantViolations)
				for i, v := range violations {
					t.Logf("  violation[%d]: %v", i, v)
				}
			}
		})
	}
}

func TestValidateConfigAgainstPolicy_ProfileRequirements(t *testing.T) {
	policy := &LoggingPolicy{
		ProfileRequirements: map[LoggingProfile]ProfileConstraints{
			ProfileEnterprise: {
				MinEnvironment:   "staging",
				RequiredFeatures: []string{"middleware", "throttling"},
			},
		},
	}

	tests := []struct {
		name           string
		config         *LoggerConfig
		environment    string
		wantViolations int
	}{
		{
			name: "ENTERPRISE with all requirements - valid",
			config: &LoggerConfig{
				Profile: ProfileEnterprise,
				Service: "test",
				Middleware: []MiddlewareConfig{
					{Name: "correlation", Enabled: true},
				},
				Throttling: &ThrottlingConfig{Enabled: true},
			},
			environment:    "production",
			wantViolations: 0,
		},
		{
			name: "ENTERPRISE missing middleware - invalid",
			config: &LoggerConfig{
				Profile:    ProfileEnterprise,
				Service:    "test",
				Throttling: &ThrottlingConfig{Enabled: true},
			},
			environment:    "production",
			wantViolations: 1,
		},
		{
			name: "ENTERPRISE missing throttling - invalid",
			config: &LoggerConfig{
				Profile: ProfileEnterprise,
				Service: "test",
				Middleware: []MiddlewareConfig{
					{Name: "correlation", Enabled: true},
				},
			},
			environment:    "production",
			wantViolations: 1,
		},
		{
			name: "ENTERPRISE in development environment - invalid",
			config: &LoggerConfig{
				Profile: ProfileEnterprise,
				Service: "test",
				Middleware: []MiddlewareConfig{
					{Name: "correlation", Enabled: true},
				},
				Throttling: &ThrottlingConfig{Enabled: true},
			},
			environment:    "development",
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := ValidateConfigAgainstPolicy(tt.config, policy, tt.environment, "")

			if len(violations) != tt.wantViolations {
				t.Errorf("got %d violations, want %d", len(violations), tt.wantViolations)
				for i, v := range violations {
					t.Logf("  violation[%d]: %v", i, v)
				}
			}
		})
	}
}

func TestEnforcePolicy_NonStrictMode(t *testing.T) {
	policy := &LoggingPolicy{
		AllowedProfiles: []LoggingProfile{ProfileEnterprise},
		AuditSettings: PolicyAuditSettings{
			LogPolicyViolations: true,
			EnforceStrictMode:   false,
		},
	}

	config := &LoggerConfig{
		Profile: ProfileSimple,
		Service: "test",
	}

	err := EnforcePolicy(config, policy, "", "")

	if err != nil {
		t.Errorf("non-strict mode should not return error, got: %v", err)
	}
}

func TestEnforcePolicy_StrictMode(t *testing.T) {
	policy := &LoggingPolicy{
		AllowedProfiles: []LoggingProfile{ProfileEnterprise},
		AuditSettings: PolicyAuditSettings{
			LogPolicyViolations: true,
			EnforceStrictMode:   true,
		},
	}

	config := &LoggerConfig{
		Profile: ProfileSimple,
		Service: "test",
	}

	err := EnforcePolicy(config, policy, "", "")

	if err == nil {
		t.Error("strict mode should return error for policy violations")
	}
}

func TestEnforcePolicy_NoViolations(t *testing.T) {
	policy := &LoggingPolicy{
		AllowedProfiles: []LoggingProfile{ProfileSimple, ProfileStructured},
		AuditSettings: PolicyAuditSettings{
			EnforceStrictMode: true,
		},
	}

	config := &LoggerConfig{
		Profile: ProfileSimple,
		Service: "test",
	}

	err := EnforcePolicy(config, policy, "", "")

	if err != nil {
		t.Errorf("no violations should not return error, got: %v", err)
	}
}

func TestHasFeature(t *testing.T) {
	tests := []struct {
		name    string
		config  *LoggerConfig
		feature string
		want    bool
	}{
		{
			name: "has middleware",
			config: &LoggerConfig{
				Middleware: []MiddlewareConfig{{Name: "test"}},
			},
			feature: "middleware",
			want:    true,
		},
		{
			name:    "no middleware",
			config:  &LoggerConfig{},
			feature: "middleware",
			want:    false,
		},
		{
			name: "has throttling",
			config: &LoggerConfig{
				Throttling: &ThrottlingConfig{Enabled: true},
			},
			feature: "throttling",
			want:    true,
		},
		{
			name: "throttling disabled",
			config: &LoggerConfig{
				Throttling: &ThrottlingConfig{Enabled: false},
			},
			feature: "throttling",
			want:    false,
		},
		{
			name: "has policy",
			config: &LoggerConfig{
				PolicyFile: "policy.yaml",
			},
			feature: "policy",
			want:    true,
		},
		{
			name: "has caller",
			config: &LoggerConfig{
				EnableCaller: true,
			},
			feature: "caller",
			want:    true,
		},
		{
			name: "has stacktrace",
			config: &LoggerConfig{
				EnableStacktrace: true,
			},
			feature: "stacktrace",
			want:    true,
		},
		{
			name:    "unknown feature",
			config:  &LoggerConfig{},
			feature: "unknown",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFeature(tt.config, tt.feature)
			if got != tt.want {
				t.Errorf("hasFeature() = %v, want %v", got, tt.want)
			}
		})
	}
}
