package ascii

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fulmenhq/gofulmen/config"
	"gopkg.in/yaml.v3"
)

//go:embed assets/terminal-overrides-defaults.yaml
var terminalOverridesDefaults []byte

type TerminalOverrides struct {
	Version   string                    `yaml:"version" json:"version"`
	Terminals map[string]TerminalConfig `yaml:"terminals" json:"terminals"`
}

type TerminalConfig struct {
	Name      string         `yaml:"name" json:"name"`
	Overrides map[string]int `yaml:"overrides,omitempty" json:"overrides,omitempty"`
	Notes     string         `yaml:"notes,omitempty" json:"notes,omitempty"`
}

var (
	terminalCatalog       *TerminalOverrides
	currentTerminalConfig *TerminalConfig
)

func init() {
	if err := loadTerminalCatalog(); err != nil {
		return
	}
	detectCurrentTerminal()
}

func loadTerminalCatalog() error {
	// Layer 1: Load embedded defaults from crucible SSOT
	// Create a fresh instance to avoid modifying any existing config
	var catalog TerminalOverrides
	if err := yaml.Unmarshal(terminalOverridesDefaults, &catalog); err != nil {
		return fmt.Errorf("failed to load embedded terminal overrides: %w", err)
	}
	terminalCatalog = &catalog

	// Layer 2: Merge user overrides from GetFulmenConfigDir()
	fulmenConfigDir := config.GetFulmenConfigDir()
	userConfigPath := filepath.Join(fulmenConfigDir, "terminal-overrides.yaml")

	if _, err := os.Stat(userConfigPath); err == nil {
		if err := loadUserOverrides(userConfigPath); err != nil {
			return err
		}
	}

	return nil
}

func loadUserOverrides(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read user terminal overrides: %w", err)
	}

	var userConfig TerminalOverrides
	if err := yaml.Unmarshal(data, &userConfig); err != nil {
		return fmt.Errorf("failed to parse user terminal overrides: %w", err)
	}

	mergeTerminalConfigs(terminalCatalog, &userConfig)
	return nil
}

func mergeTerminalConfigs(base, override *TerminalOverrides) {
	if override.Terminals == nil {
		return
	}

	if base.Terminals == nil {
		base.Terminals = make(map[string]TerminalConfig)
	}

	for termID, userConfig := range override.Terminals {
		baseConfig, exists := base.Terminals[termID]
		if !exists {
			base.Terminals[termID] = userConfig
			continue
		}

		if baseConfig.Overrides == nil {
			baseConfig.Overrides = make(map[string]int)
		}

		for char, width := range userConfig.Overrides {
			baseConfig.Overrides[char] = width
		}

		if userConfig.Name != "" {
			baseConfig.Name = userConfig.Name
		}
		if userConfig.Notes != "" {
			baseConfig.Notes = userConfig.Notes
		}

		base.Terminals[termID] = baseConfig
	}
}

func detectCurrentTerminal() {
	if terminalCatalog == nil {
		return
	}

	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "" {
		term := os.Getenv("TERM")
		if strings.Contains(term, "ghostty") {
			termProgram = "ghostty"
		}
	}

	if termProgram != "" {
		if config, exists := terminalCatalog.Terminals[termProgram]; exists {
			currentTerminalConfig = &config
			return
		}
	}

	currentTerminalConfig = nil
}

func GetTerminalConfig() *TerminalConfig {
	return currentTerminalConfig
}

func GetAllTerminalConfigs() map[string]TerminalConfig {
	if terminalCatalog == nil {
		return nil
	}
	return terminalCatalog.Terminals
}

// SetTerminalOverrides allows external applications to provide their own terminal configuration
// This implements Layer 3 of the 3-layer config pattern (BYOC - Bring Your Own Config)
//
// Example usage for sophisticated TUI apps:
//
//	myConfig := &ascii.TerminalOverrides{
//	    Version: "1.0.0",
//	    Terminals: map[string]ascii.TerminalConfig{
//	        "myterm": {
//	            Name: "My Terminal",
//	            Overrides: map[string]int{"🔧": 2},
//	        },
//	    },
//	}
//	ascii.SetTerminalOverrides(myConfig)
func SetTerminalOverrides(overrides *TerminalOverrides) {
	terminalCatalog = overrides
	detectCurrentTerminal()
}

// SetTerminalConfig allows setting configuration for a specific terminal
// This is a convenience function for Layer 3 (BYOC) when you only need to override one terminal
//
// Example:
//
//	ascii.SetTerminalConfig("ghostty", ascii.TerminalConfig{
//	    Name: "Ghostty",
//	    Overrides: map[string]int{"🔧": 3},
//	})
func SetTerminalConfig(terminalName string, cfg TerminalConfig) {
	if terminalCatalog == nil {
		terminalCatalog = &TerminalOverrides{
			Version:   "1.0.0",
			Terminals: make(map[string]TerminalConfig),
		}
	}
	if terminalCatalog.Terminals == nil {
		terminalCatalog.Terminals = make(map[string]TerminalConfig)
	}
	terminalCatalog.Terminals[terminalName] = cfg
	detectCurrentTerminal()
}

// ReloadTerminalOverrides reloads the terminal configuration from defaults and user overrides
// This is useful if you want to reset after using SetTerminalOverrides or SetTerminalConfig
func ReloadTerminalOverrides() error {
	return loadTerminalCatalog()
}
