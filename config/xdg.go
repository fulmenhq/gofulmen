package config

import (
	"os"
	"path/filepath"

	"github.com/fulmenhq/gofulmen/errors"
)

// XDGBaseDirs provides XDG Base Directory paths
type XDGBaseDirs struct {
	ConfigHome string
	DataHome   string
	CacheHome  string
}

// GetXDGBaseDirs returns the XDG Base Directory paths
func GetXDGBaseDirs() XDGBaseDirs {
	return XDGBaseDirs{
		ConfigHome: getXDGConfigHome(),
		DataHome:   getXDGDataHome(),
		CacheHome:  getXDGCacheHome(),
	}
}

// GetXDGBaseDirsWithEnvelope returns the XDG Base Directory paths with structured error reporting.
// Returns an error envelope if HOME environment variable is not set.
func GetXDGBaseDirsWithEnvelope(correlationID string) (XDGBaseDirs, error) {
	home := os.Getenv("HOME")
	if home == "" {
		envelope := errors.NewErrorEnvelope("CONFIG_XDG_ERROR", "HOME environment variable not set")
		envelope = errors.SafeWithSeverity(envelope, errors.SeverityHigh)
		envelope = envelope.WithCorrelationID(correlationID)
		envelope = errors.SafeWithContext(envelope, map[string]interface{}{
			"component":  "config",
			"operation":  "get_xdg_dirs",
			"error_type": "missing_home_env",
		})
		return XDGBaseDirs{}, envelope
	}

	return GetXDGBaseDirs(), nil
}

func getXDGConfigHome() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}
	return ""
}

func getXDGDataHome() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return dataHome
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "share")
	}
	return ""
}

func getXDGCacheHome() string {
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		return cacheHome
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".cache")
	}
	return ""
}

// GetAppConfigDir returns the config directory for a given app name
// Uses XDG Base Directory specification: $XDG_CONFIG_HOME/appName or ~/.config/appName
func GetAppConfigDir(appName string) string {
	xdg := GetXDGBaseDirs()
	return filepath.Join(xdg.ConfigHome, appName)
}

// GetAppDataDir returns the data directory for a given app name
// Uses XDG Base Directory specification: $XDG_DATA_HOME/appName or ~/.local/share/appName
func GetAppDataDir(appName string) string {
	xdg := GetXDGBaseDirs()
	return filepath.Join(xdg.DataHome, appName)
}

// GetAppCacheDir returns the cache directory for a given app name
// Uses XDG Base Directory specification: $XDG_CACHE_HOME/appName or ~/.cache/appName
func GetAppCacheDir(appName string) string {
	xdg := GetXDGBaseDirs()
	return filepath.Join(xdg.CacheHome, appName)
}

// GetFulmenConfigDir returns the Fulmen ecosystem config directory
// This is a convenience function for Fulmen ecosystem tools
// Returns: ~/.config/fulmen (or $XDG_CONFIG_HOME/fulmen)
func GetFulmenConfigDir() string {
	return GetAppConfigDir("fulmen")
}

// GetFulmenDataDir returns the Fulmen ecosystem data directory
// Returns: ~/.local/share/fulmen (or $XDG_DATA_HOME/fulmen)
func GetFulmenDataDir() string {
	return GetAppDataDir("fulmen")
}

// GetFulmenCacheDir returns the Fulmen ecosystem cache directory
// Returns: ~/.cache/fulmen (or $XDG_CACHE_HOME/fulmen)
func GetFulmenCacheDir() string {
	return GetAppCacheDir("fulmen")
}

// Deprecated: Use GetAppConfigDir("your-app") or GetFulmenConfigDir() for Fulmen ecosystem
func GetGofulmenConfigDir() string {
	return GetFulmenConfigDir()
}

// Deprecated: Use GetAppDataDir("your-app") or GetFulmenDataDir() for Fulmen ecosystem
func GetGofulmenDataDir() string {
	return GetFulmenDataDir()
}

// Deprecated: Use GetAppCacheDir("your-app") or GetFulmenCacheDir() for Fulmen ecosystem
func GetGofulmenCacheDir() string {
	return GetFulmenCacheDir()
}
