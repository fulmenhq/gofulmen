package config

import (
	"os"
	"path/filepath"
)

type Config struct {
}

func LoadConfig() (*Config, error) {
	return &Config{}, nil
}

// GetConfigPaths returns default config search paths for fulmen ecosystem
// Deprecated: Use GetAppConfigPaths() with your app name for non-Fulmen tools
func GetConfigPaths() []string {
	return GetAppConfigPaths("fulmen", "gofulmen")
}

// GetAppConfigPaths returns config search paths for a given app name
// Searches in order:
//  1. XDG config dir (e.g., ~/.config/appName/config.yaml)
//  2. Dot-directory in home (e.g., ~/.appName/config.yaml)
//  3. Dot-file in home (e.g., ~/.appName.yaml)
//  4. Current directory (e.g., ./appName.yaml)
//
// If legacyNames are provided, also searches those locations for backward compatibility
func GetAppConfigPaths(appName string, legacyNames ...string) []string {
	xdg := GetXDGBaseDirs()
	home := os.Getenv("HOME")

	var paths []string

	// 1. XDG config directory (preferred)
	paths = append(paths,
		filepath.Join(xdg.ConfigHome, appName, "config.yaml"),
		filepath.Join(xdg.ConfigHome, appName, "config.json"),
	)

	// 2. Dot-directory in home
	if home != "" {
		paths = append(paths,
			filepath.Join(home, "."+appName, "config.yaml"),
			filepath.Join(home, "."+appName, "config.json"),
		)
	}

	// 3. Dot-file in home (single file)
	if home != "" {
		paths = append(paths,
			filepath.Join(home, "."+appName+".yaml"),
			filepath.Join(home, "."+appName+".json"),
		)
	}

	// 4. Current directory
	paths = append(paths,
		"./"+appName+".yaml",
		"./"+appName+".json",
		"./."+appName+".yaml",
		"./."+appName+".json",
	)

	// 5. Legacy locations (if provided)
	for _, legacyName := range legacyNames {
		if legacyName != appName {
			paths = append(paths,
				filepath.Join(xdg.ConfigHome, legacyName, "config.json"),
			)
			if home != "" {
				paths = append(paths,
					filepath.Join(home, "."+legacyName+".json"),
				)
			}
		}
	}

	return paths
}

// SaveConfig saves configuration to the specified path
func SaveConfig(config *Config, path string) error {
	dir := filepath.Dir(path)
	// #nosec G301 -- config directories use 0755 for multi-user access compatibility
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// #nosec G304 -- intentional user-controlled file creation for saving configuration to user-specified path
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
