package signals

import (
	"fmt"
	"sync"

	"github.com/fulmenhq/gofulmen/crucible"
	"gopkg.in/yaml.v3"
)

// Catalog provides immutable access to the signal handling configuration.
//
// The catalog loads signal definitions, behaviors, and fallback metadata from
// Crucible's embedded configuration using lazy loading for performance.
// All data is cached after first access and works offline in compiled binaries.
//
// Example:
//
//	catalog := GetDefaultCatalog()
//	signal, _ := catalog.GetSignal("term")
//	fmt.Printf("%s: %s\n", signal.Name, signal.Description)
type Catalog struct {
	// Lazy-loaded data with mutex protection
	config     *SignalCatalog
	configOnce sync.Once
	configErr  error
}

// SignalCatalog represents the top-level signal catalog structure.
type SignalCatalog struct {
	Schema      string              `yaml:"$schema"`
	Description string              `yaml:"description"`
	Version     string              `yaml:"version"`
	Signals     []*SignalDefinition `yaml:"signals"`

	// Internal index for fast lookups
	signalsByID   map[string]*SignalDefinition
	signalsByName map[string]*SignalDefinition
}

// SignalDefinition represents a single signal with its behavior and metadata.
type SignalDefinition struct {
	ID                     string           `yaml:"id"`
	Name                   string           `yaml:"name"`
	UnixNumber             int              `yaml:"unix_number"`
	PlatformOverrides      map[string]int   `yaml:"platform_overrides,omitempty"`
	WindowsEvent           *string          `yaml:"windows_event"`
	Description            string           `yaml:"description"`
	DefaultBehavior        string           `yaml:"default_behavior"`
	ExitCode               int              `yaml:"exit_code"`
	TimeoutSeconds         int              `yaml:"timeout_seconds"`
	DoubleTapWindowSeconds *int             `yaml:"double_tap_window_seconds,omitempty"`
	DoubleTapMessage       string           `yaml:"double_tap_message,omitempty"`
	DoubleTapBehavior      string           `yaml:"double_tap_behavior,omitempty"`
	DoubleTapExitCode      *int             `yaml:"double_tap_exit_code,omitempty"`
	ReloadStrategy         string           `yaml:"reload_strategy,omitempty"`
	ValidationRequired     *bool            `yaml:"validation_required,omitempty"`
	CleanupActions         []string         `yaml:"cleanup_actions,omitempty"`
	UsageNotes             string           `yaml:"usage_notes,omitempty"`
	WindowsFallback        *WindowsFallback `yaml:"windows_fallback,omitempty"`
}

// WindowsFallback describes fallback behavior when a signal is unavailable on Windows.
type WindowsFallback struct {
	FallbackBehavior string            `yaml:"fallback_behavior"`
	LogLevel         string            `yaml:"log_level"`
	LogMessage       string            `yaml:"log_message"`
	LogTemplate      string            `yaml:"log_template"`
	OperationHint    string            `yaml:"operation_hint"`
	TelemetryEvent   string            `yaml:"telemetry_event"`
	TelemetryTags    map[string]string `yaml:"telemetry_tags"`
}

// NewCatalog creates a new Catalog instance.
//
// The catalog uses lazy loading - data is only loaded when first accessed.
// Configuration is loaded from Crucible's embedded signals.yaml.
//
// Example:
//
//	catalog := NewCatalog()
func NewCatalog() *Catalog {
	return &Catalog{}
}

// GetDefaultCatalog returns a singleton catalog.
//
// This is a convenience function for applications that don't need custom
// catalog instances.
//
// Example:
//
//	catalog := GetDefaultCatalog()
//	version := catalog.Version()
func GetDefaultCatalog() *Catalog {
	defaultCatalogOnce.Do(func() {
		defaultCatalog = NewCatalog()
	})
	return defaultCatalog
}

var (
	defaultCatalog     *Catalog
	defaultCatalogOnce sync.Once
)

// load ensures the catalog is loaded from Crucible.
func (c *Catalog) load() error {
	c.configOnce.Do(func() {
		// Load from Crucible's embedded config
		data, err := crucible.ConfigRegistry.Library().Foundry().Signals()
		if err != nil {
			c.configErr = fmt.Errorf("failed to load signals catalog: %w", err)
			return
		}

		// Parse YAML
		var catalog SignalCatalog
		if err := yaml.Unmarshal(data, &catalog); err != nil {
			c.configErr = fmt.Errorf("failed to parse signals catalog: %w", err)
			return
		}

		// Build indexes for fast lookup
		catalog.signalsByID = make(map[string]*SignalDefinition, len(catalog.Signals))
		catalog.signalsByName = make(map[string]*SignalDefinition, len(catalog.Signals))
		for _, signal := range catalog.Signals {
			catalog.signalsByID[signal.ID] = signal
			catalog.signalsByName[signal.Name] = signal
		}

		c.config = &catalog
	})

	return c.configErr
}

// Version returns the catalog version string.
//
// Example:
//
//	version := catalog.Version() // "v1.0.0"
func (c *Catalog) Version() (string, error) {
	if err := c.load(); err != nil {
		return "", err
	}
	return c.config.Version, nil
}

// GetSignal retrieves a signal definition by ID (e.g., "term", "int", "hup").
//
// Returns an error if the signal ID is not found in the catalog.
//
// Example:
//
//	signal, err := catalog.GetSignal("term")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Exit code: %d\n", signal.ExitCode)
func (c *Catalog) GetSignal(id string) (*SignalDefinition, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	signal, exists := c.config.signalsByID[id]
	if !exists {
		return nil, fmt.Errorf("signal not found: %s", id)
	}

	return signal, nil
}

// GetSignalByName retrieves a signal definition by name (e.g., "SIGTERM", "SIGINT").
//
// Returns an error if the signal name is not found in the catalog.
//
// Example:
//
//	signal, err := catalog.GetSignalByName("SIGTERM")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Catalog) GetSignalByName(name string) (*SignalDefinition, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	signal, exists := c.config.signalsByName[name]
	if !exists {
		return nil, fmt.Errorf("signal not found: %s", name)
	}

	return signal, nil
}

// ListSignals returns all signal definitions in the catalog.
//
// Example:
//
//	signals, err := catalog.ListSignals()
//	for _, signal := range signals {
//	    fmt.Printf("%s (%s): %s\n", signal.ID, signal.Name, signal.Description)
//	}
func (c *Catalog) ListSignals() ([]*SignalDefinition, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	// Return a copy to prevent modifications
	signals := make([]*SignalDefinition, len(c.config.Signals))
	copy(signals, c.config.Signals)
	return signals, nil
}

// GetDescription returns the catalog description.
func (c *Catalog) GetDescription() (string, error) {
	if err := c.load(); err != nil {
		return "", err
	}
	return c.config.Description, nil
}
