package signals

import (
	"fmt"
	"strconv"
	"strings"
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

	// Internal indexes for fast lookups
	signalsByID     map[string]*SignalDefinition
	signalsByName   map[string]*SignalDefinition
	signalsByNumber map[int]*SignalDefinition
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
		catalog.signalsByNumber = make(map[int]*SignalDefinition, len(catalog.Signals))
		for _, signal := range catalog.Signals {
			catalog.signalsByID[signal.ID] = signal
			catalog.signalsByName[signal.Name] = signal
			catalog.signalsByNumber[signal.UnixNumber] = signal
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

// ResolveSignal resolves a signal from common name variants.
//
// This function provides ergonomic signal lookup for CLI applications,
// accepting various input formats that users commonly provide.
// Returns nil if the signal cannot be resolved.
//
// Resolution order (per Crucible Foundry interface standard):
//  1. Trim leading/trailing whitespace
//  2. Return nil if empty after trim
//  3. Exact catalog name match (e.g., "SIGTERM")
//  4. Numeric match by unix_number (e.g., "15" -> SIGTERM)
//  5. Uppercase normalization with SIG prefix handling
//  6. ID fallback (lowercase lookup)
//  7. Return nil if not found
//
// Example:
//
//	signal := catalog.ResolveSignal("term")     // Returns SIGTERM
//	signal = catalog.ResolveSignal("SIGTERM")  // Returns SIGTERM
//	signal = catalog.ResolveSignal("15")       // Returns SIGTERM
//	signal = catalog.ResolveSignal("  term  ") // Returns SIGTERM
//	signal = catalog.ResolveSignal("SIGFOO")   // Returns nil
func (c *Catalog) ResolveSignal(name string) *SignalDefinition {
	if err := c.load(); err != nil {
		return nil
	}

	// Step 1: Trim whitespace
	name = strings.TrimSpace(name)

	// Step 2: Return nil if empty
	if name == "" {
		return nil
	}

	// Step 3: Exact catalog name match
	if signal, exists := c.config.signalsByName[name]; exists {
		return signal
	}

	// Step 4: Numeric match (positive integers only)
	if num, err := strconv.Atoi(name); err == nil && num > 0 {
		if signal, exists := c.config.signalsByNumber[num]; exists {
			return signal
		}
	}

	// Step 5: Uppercase normalization
	upper := strings.ToUpper(name)
	if strings.HasPrefix(upper, "SIG") {
		// Already has SIG prefix, try lookup
		if signal, exists := c.config.signalsByName[upper]; exists {
			return signal
		}
	} else {
		// Prepend SIG and try lookup
		if signal, exists := c.config.signalsByName["SIG"+upper]; exists {
			return signal
		}
	}

	// Step 6: ID fallback (lowercase)
	lower := strings.ToLower(name)
	if signal, exists := c.config.signalsByID[lower]; exists {
		return signal
	}

	// Step 7: Not found
	return nil
}

// ListSignalNames returns all signal names for CLI completion.
//
// The returned names are in catalog order and can be used for
// shell completion, help text, or validation feedback.
//
// Example:
//
//	names, err := catalog.ListSignalNames()
//	// names: ["SIGHUP", "SIGINT", "SIGQUIT", "SIGTERM", ...]
func (c *Catalog) ListSignalNames() ([]string, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	names := make([]string, len(c.config.Signals))
	for i, signal := range c.config.Signals {
		names[i] = signal.Name
	}
	return names, nil
}

// MatchSignalNames returns signal names matching a glob pattern.
//
// Supports simple glob patterns:
//   - * matches zero or more characters
//   - ? matches exactly one character
//   - Matching is case-insensitive
//
// Example:
//
//	matches, _ := catalog.MatchSignalNames("SIG*")    // All signals
//	matches, _ = catalog.MatchSignalNames("*USR*")   // SIGUSR1, SIGUSR2
//	matches, _ = catalog.MatchSignalNames("SIG???")  // SIGINT, SIGHUP (3-char suffix)
func (c *Catalog) MatchSignalNames(pattern string) ([]string, error) {
	if err := c.load(); err != nil {
		return nil, err
	}

	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return []string{}, nil
	}

	var matches []string
	patternLower := strings.ToLower(pattern)

	for _, signal := range c.config.Signals {
		nameLower := strings.ToLower(signal.Name)
		if globMatch(patternLower, nameLower) {
			matches = append(matches, signal.Name)
		}
	}

	return matches, nil
}

// globMatch performs simple glob pattern matching.
// Supports * (zero or more chars) and ? (exactly one char).
// Both pattern and text should be pre-lowercased for case-insensitive matching.
func globMatch(pattern, text string) bool {
	px, tx := 0, 0
	nextPx, nextTx := 0, -1

	for tx < len(text) || px < len(pattern) {
		if px < len(pattern) {
			switch pattern[px] {
			case '*':
				// Try to match at current position, or advance text
				nextPx = px
				nextTx = tx + 1
				px++
				continue
			case '?':
				// Match exactly one character
				if tx < len(text) {
					px++
					tx++
					continue
				}
			default:
				// Match literal character
				if tx < len(text) && pattern[px] == text[tx] {
					px++
					tx++
					continue
				}
			}
		}

		// Mismatch - backtrack if we have a saved * position
		if nextTx > 0 && nextTx <= len(text) {
			px = nextPx + 1
			tx = nextTx
			nextTx++
			continue
		}

		return false
	}

	return true
}
