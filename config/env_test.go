package config

import (
	"testing"
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

func TestLoadEnvOverrides_Invalid(t *testing.T) {
	t.Setenv("APP_RETRIES", "not-a-number")
	_, err := LoadEnvOverrides([]EnvVarSpec{
		{Name: "APP_RETRIES", Path: []string{"settings", "retries"}, Type: EnvInt},
	})
	if err == nil {
		t.Fatalf("expected error for invalid integer override")
	}
}
