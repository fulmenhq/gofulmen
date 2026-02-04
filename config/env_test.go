package config

import (
	stdErrors "errors"
	"testing"

	gferrors "github.com/fulmenhq/gofulmen/errors"
)

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("APP_RETRIES", "7")
	t.Setenv("APP_ENABLE", "true")

	overrides, err := LoadEnvOverrides([]EnvVarSpec{
		{Name: "APP_RETRIES", Path: []string{"settings", "retries"}, Type: EnvInt},
		{Name: "APP_ENABLE", Path: []string{"settings", "enabled"}, Type: EnvBool},
	})
	if err != nil {
		t.Fatalf("LoadEnvOverrides returned error: %v", err)
	}

	settings := overrides["settings"].(map[string]any)
	if val := settings["retries"].(int); val != 7 {
		t.Fatalf("expected retries=7, got %v", val)
	}
	if val := settings["enabled"].(bool); !val {
		t.Fatalf("expected enabled=true")
	}
}

func TestLoadEnvOverridesWithReport_AliasesAndConflicts(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "8080")
	t.Setenv("APP_PORT", "9000")

	report, err := LoadEnvOverridesWithReport([]EnvVarSpecWithAliases{
		{
			Name:    "APP_SERVER_PORT",
			Aliases: []string{"APP_PORT"},
			Path:    []string{"server", "port"},
			Type:    EnvInt,
		},
	})
	if err != nil {
		t.Fatalf("LoadEnvOverridesWithReport returned error: %v", err)
	}

	// Alias takes precedence.
	server, ok := report.Overrides["server"].(map[string]any)
	if !ok {
		t.Fatalf("expected server map, got %T", report.Overrides["server"])
	}
	if got := server["port"]; got != 9000 {
		t.Fatalf("server.port=%v, want 9000", got)
	}

	if len(report.Applied) != 1 {
		t.Fatalf("Applied len=%d, want 1", len(report.Applied))
	}
	if report.Applied[0].ChosenName != "APP_PORT" || report.Applied[0].Source != EnvVarSourceAlias {
		t.Fatalf("Applied=%+v, want alias APP_PORT", report.Applied[0])
	}

	if len(report.Conflicts) != 1 {
		t.Fatalf("Conflicts len=%d, want 1", len(report.Conflicts))
	}
	if report.Conflicts[0].ChosenName != "APP_PORT" {
		t.Fatalf("Conflict ChosenName=%q, want APP_PORT", report.Conflicts[0].ChosenName)
	}
	if report.Conflicts[0].Canonical != "8080" {
		t.Fatalf("Conflict Canonical=%q, want 8080", report.Conflicts[0].Canonical)
	}
	if report.Conflicts[0].Alias != "9000" {
		t.Fatalf("Conflict Alias=%q, want 9000", report.Conflicts[0].Alias)
	}
}

func TestLoadEnvOverridesWithReport_NoConflictWhenValuesMatch(t *testing.T) {
	t.Setenv("APP_SERVER_PORT", "8080")
	t.Setenv("APP_PORT", " 8080 ")

	report, err := LoadEnvOverridesWithReport([]EnvVarSpecWithAliases{
		{
			Name:    "APP_SERVER_PORT",
			Aliases: []string{"APP_PORT"},
			Path:    []string{"server", "port"},
			Type:    EnvInt,
		},
	})
	if err != nil {
		t.Fatalf("LoadEnvOverridesWithReport returned error: %v", err)
	}
	if len(report.Conflicts) != 0 {
		t.Fatalf("Conflicts len=%d, want 0", len(report.Conflicts))
	}
}

func TestLoadEnvOverridesWithReport_ConflictValuesMaskedWhenSensitive(t *testing.T) {
	t.Setenv("APP_SERVER_TOKEN", "canonical-secret")
	t.Setenv("APP_TOKEN", "alias-secret")

	report, err := LoadEnvOverridesWithReport([]EnvVarSpecWithAliases{
		{
			Name:    "APP_SERVER_TOKEN",
			Aliases: []string{"APP_TOKEN"},
			Path:    []string{"auth", "token"},
			Type:    EnvString,
		},
	})
	if err != nil {
		t.Fatalf("LoadEnvOverridesWithReport returned error: %v", err)
	}
	if len(report.Conflicts) != 1 {
		t.Fatalf("Conflicts len=%d, want 1", len(report.Conflicts))
	}
	if !report.Conflicts[0].Masked {
		t.Fatalf("Conflict Masked=false, want true")
	}
	if report.Conflicts[0].Canonical != "[set]" || report.Conflicts[0].Alias != "[set]" {
		t.Fatalf("Conflict values not masked: %+v", report.Conflicts[0])
	}
}

func TestLoadEnvOverridesWithEnvelopeAndReport_ParseErrorMasksSensitiveValue(t *testing.T) {
	t.Setenv("APP_TOKEN", "not-a-number")

	_, err := LoadEnvOverridesWithEnvelopeAndReport([]EnvVarSpecWithAliases{
		{
			Name: "APP_TOKEN",
			Path: []string{"auth", "token"},
			Type: EnvInt,
		},
	}, "cid")
	if err == nil {
		t.Fatalf("expected error")
	}

	var env *gferrors.ErrorEnvelope
	if !stdErrors.As(err, &env) {
		t.Fatalf("expected *errors.ErrorEnvelope, got %T", err)
	}
	if got := env.Context["env_value"]; got != "[set]" {
		t.Fatalf("env_value=%v, want [set]", got)
	}
	if got := env.Context["env_masked"]; got != true {
		t.Fatalf("env_masked=%v, want true", got)
	}
}

func TestLoadEnvOverrides_Invalid(t *testing.T) {
	t.Setenv("APP_RETRIES", "not-a-number")
	_, err := LoadEnvOverrides([]EnvVarSpec{
		{Name: "APP_RETRIES", Path: []string{"settings", "retries"}, Type: EnvInt},
	})
	if err == nil {
		t.Fatalf("expected error for invalid integer override")
	}
}
