package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigPaths(t *testing.T) {
	paths := GetConfigPaths()
	if len(paths) == 0 {
		t.Error("Should return at least one config path")
	}

	foundFulmen := false
	for _, path := range paths {
		if strings.Contains(path, "fulmen") {
			foundFulmen = true
			break
		}
	}
	if !foundFulmen {
		t.Error("Config paths should include fulmen locations")
	}
}

func TestGetAppConfigPaths(t *testing.T) {
	paths := GetAppConfigPaths("myapp")
	if len(paths) == 0 {
		t.Error("Should return at least one config path")
	}

	foundMyapp := false
	for _, path := range paths {
		if strings.Contains(path, "myapp") {
			foundMyapp = true
			break
		}
	}
	if !foundMyapp {
		t.Error("Config paths should include myapp locations")
	}

	t.Logf("App config paths for 'myapp': %v", paths)
}

func TestGetAppConfigPathsWithLegacy(t *testing.T) {
	paths := GetAppConfigPaths("newapp", "oldapp", "legacy")
	if len(paths) == 0 {
		t.Error("Should return at least one config path")
	}

	foundNew := false
	foundLegacy := false
	for _, path := range paths {
		if strings.Contains(path, "newapp") {
			foundNew = true
		}
		if strings.Contains(path, "oldapp") || strings.Contains(path, "legacy") {
			foundLegacy = true
		}
	}

	if !foundNew {
		t.Error("Config paths should include newapp locations")
	}
	if !foundLegacy {
		t.Error("Config paths should include legacy locations")
	}

	t.Logf("App config paths with legacy: %v", paths)
}

func TestGetXDGBaseDirs(t *testing.T) {
	xdg := GetXDGBaseDirs()
	if xdg.ConfigHome == "" {
		t.Error("ConfigHome should not be empty")
	}
	t.Logf("XDG Config Home: %s", xdg.ConfigHome)
	t.Logf("XDG Data Home: %s", xdg.DataHome)
	t.Logf("XDG Cache Home: %s", xdg.CacheHome)
}

func TestGetAppConfigDir(t *testing.T) {
	dir := GetAppConfigDir("testapp")
	if !strings.Contains(dir, "testapp") {
		t.Errorf("App config dir should contain app name, got: %s", dir)
	}
	t.Logf("App config dir: %s", dir)
}

func TestGetAppDataDir(t *testing.T) {
	dir := GetAppDataDir("testapp")
	if !strings.Contains(dir, "testapp") {
		t.Errorf("App data dir should contain app name, got: %s", dir)
	}
	t.Logf("App data dir: %s", dir)
}

func TestGetAppCacheDir(t *testing.T) {
	dir := GetAppCacheDir("testapp")
	if !strings.Contains(dir, "testapp") {
		t.Errorf("App cache dir should contain app name, got: %s", dir)
	}
	t.Logf("App cache dir: %s", dir)
}

func TestGetFulmenConfigDir(t *testing.T) {
	dir := GetFulmenConfigDir()
	expected := filepath.Join(os.Getenv("HOME"), ".config", "fulmen")
	if dir != expected {
		t.Errorf("Expected %s, got %s", expected, dir)
	}
	t.Logf("Fulmen config dir: %s", dir)
}

func TestGetFulmenDataDir(t *testing.T) {
	dir := GetFulmenDataDir()
	if !strings.Contains(dir, "fulmen") {
		t.Errorf("Fulmen data dir should contain 'fulmen', got: %s", dir)
	}
	t.Logf("Fulmen data dir: %s", dir)
}

func TestGetFulmenCacheDir(t *testing.T) {
	dir := GetFulmenCacheDir()
	if !strings.Contains(dir, "fulmen") {
		t.Errorf("Fulmen cache dir should contain 'fulmen', got: %s", dir)
	}
	t.Logf("Fulmen cache dir: %s", dir)
}

func TestGetGofulmenConfigDirDeprecated(t *testing.T) {
	dir := GetGofulmenConfigDir()
	fulmenDir := GetFulmenConfigDir()
	if dir != fulmenDir {
		t.Error("Deprecated GetGofulmenConfigDir should return same as GetFulmenConfigDir")
	}
}

func TestGetGofulmenDataDir(t *testing.T) {
	dir := GetGofulmenDataDir()
	fulmenDir := GetFulmenDataDir()
	if dir != fulmenDir {
		t.Error("Deprecated GetGofulmenDataDir should return same as GetFulmenDataDir")
	}
	if !strings.Contains(dir, "fulmen") {
		t.Errorf("Gofulmen data dir should contain 'fulmen', got: %s", dir)
	}
	t.Logf("Gofulmen data dir: %s", dir)
}

func TestGetGofulmenCacheDir(t *testing.T) {
	dir := GetGofulmenCacheDir()
	fulmenDir := GetFulmenCacheDir()
	if dir != fulmenDir {
		t.Error("Deprecated GetGofulmenCacheDir should return same as GetFulmenCacheDir")
	}
	if !strings.Contains(dir, "fulmen") {
		t.Errorf("Gofulmen cache dir should contain 'fulmen', got: %s", dir)
	}
	t.Logf("Gofulmen cache dir: %s", dir)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	config := &Config{}
	err := SaveConfig(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should have been created")
	}

	// Verify the parent directory was created
	parentDir := filepath.Dir(configPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Error("Parent directory should have been created")
	}

	t.Logf("Config saved to: %s", configPath)
}

func TestSaveConfig_InvalidPath(t *testing.T) {
	// Test with an invalid path (no permissions)
	config := &Config{}
	err := SaveConfig(config, "/root/impossible/config.yaml")
	if err == nil {
		t.Error("Should fail to save config to restricted directory")
	}
	t.Logf("Expected error: %v", err)
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config == nil {
		t.Error("Config should not be nil")
	}
}
