package logging

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fulmenhq/crucible"
	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

// LoggerConfig holds logger configuration matching crucible schema
type LoggerConfig struct {
	Profile          LoggingProfile     `json:"profile"`
	DefaultLevel     string             `json:"defaultLevel"`
	Service          string             `json:"service"`
	Component        string             `json:"component,omitempty"`
	Environment      string             `json:"environment"`
	PolicyFile       string             `json:"policyFile,omitempty"`
	Sinks            []SinkConfig       `json:"sinks"`
	Middleware       []MiddlewareConfig `json:"middleware,omitempty"`
	Throttling       *ThrottlingConfig  `json:"throttling,omitempty"`
	StaticFields     map[string]any     `json:"staticFields,omitempty"`
	EnableCaller     bool               `json:"enableCaller"`
	EnableStacktrace bool               `json:"enableStacktrace"`
}

// MiddlewareConfig defines middleware pipeline configuration
type MiddlewareConfig struct {
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Order   int            `json:"order"`
	Config  map[string]any `json:"config,omitempty"`
}

// ThrottlingConfig controls log output rate
type ThrottlingConfig struct {
	Enabled    bool   `json:"enabled"`
	MaxRate    int    `json:"maxRate"`    // logs/second
	BurstSize  int    `json:"burstSize"`  // burst capacity
	WindowSize int    `json:"windowSize"` // seconds
	DropPolicy string `json:"dropPolicy"` // "drop-oldest" | "drop-newest" | "block"
}

// SinkConfig defines an output sink
type SinkConfig struct {
	Type    string             `json:"type"` // console, file
	Level   string             `json:"level,omitempty"`
	Format  string             `json:"format"` // json, text, console
	Console *ConsoleSinkConfig `json:"console,omitempty"`
	File    *FileSinkConfig    `json:"file,omitempty"`
}

// ConsoleSinkConfig configures console output
type ConsoleSinkConfig struct {
	Stream   string `json:"stream"` // Must be "stderr"
	Colorize bool   `json:"colorize"`
}

// FileSinkConfig configures file output
type FileSinkConfig struct {
	Path       string `json:"path"`
	MaxSize    int    `json:"maxSize"`    // MB
	MaxAge     int    `json:"maxAge"`     // days
	MaxBackups int    `json:"maxBackups"` // number of old files to keep
	Compress   bool   `json:"compress"`
}

// LoadConfig loads and validates logger configuration from a file
func LoadConfig(path string) (*LoggerConfig, error) {
	return LoadConfigWithOptions(path, "")
}

// LoadConfigWithOptions loads and validates logger configuration with optional app type for policy enforcement
func LoadConfigWithOptions(path string, appType string) (*LoggerConfig, error) {
	// Read file
	// #nosec G304 -- intentional user-controlled file access for loading logger configuration from user-specified path
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Convert YAML to JSON if needed
	var jsonData []byte
	if isYAML(path) {
		var yamlContent any
		if err := yaml.Unmarshal(data, &yamlContent); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		jsonData, err = json.Marshal(yamlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
		}
	} else {
		jsonData = data
	}

	// Validate against crucible schema
	if err := ValidateConfig(jsonData); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Unmarshal to struct
	var config LoggerConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults
	applyDefaults(&config)

	// Validate console sink discipline (stderr only)
	if err := validateConsoleSinks(config.Sinks); err != nil {
		return nil, fmt.Errorf("sink validation failed: %w", err)
	}

	// Enforce logging policy if specified
	if config.PolicyFile != "" {
		policy, err := LoadPolicy(config.PolicyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load policy: %w", err)
		}

		if err := EnforcePolicy(&config, policy, config.Environment, appType); err != nil {
			return nil, fmt.Errorf("policy enforcement failed: %w", err)
		}
	}

	return &config, nil
}

// ValidateConfig validates logger config against crucible schema
func ValidateConfig(jsonData []byte) error {
	// Get logging schemas from crucible
	logging, err := crucible.SchemaRegistry.Observability().Logging().V1_0_0()
	if err != nil {
		return fmt.Errorf("failed to get logging schemas: %w", err)
	}

	// Load logger config schema
	schemaData, err := logging.LoggerConfig()
	if err != nil {
		return fmt.Errorf("failed to load logger config schema: %w", err)
	}

	// Create validator
	validator, err := schema.NewValidator(schemaData)
	if err != nil {
		return fmt.Errorf("failed to create validator: %w", err)
	}

	// Validate
	diags, err := validator.ValidateJSON(jsonData)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if verrs := schema.DiagnosticsToValidationErrors(diags); len(verrs) > 0 {
		return verrs
	}
	return nil
}

// applyDefaults applies default values to config
func applyDefaults(config *LoggerConfig) {
	if config.Profile == "" {
		config.Profile = ProfileSimple
	}
	if config.DefaultLevel == "" {
		config.DefaultLevel = "INFO"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	if config.StaticFields == nil {
		config.StaticFields = make(map[string]any)
	}

	// Apply per-sink defaults
	for i := range config.Sinks {
		sink := &config.Sinks[i]
		if sink.Format == "" {
			// Default format based on profile
			switch config.Profile {
			case ProfileSimple:
				sink.Format = "console"
			default:
				sink.Format = "json"
			}
		}
		if sink.Type == "console" && sink.Console == nil {
			sink.Console = &ConsoleSinkConfig{
				Stream:   "stderr",
				Colorize: false,
			}
		}
	}
}

// validateConsoleSinks ensures console sinks only write to stderr
func validateConsoleSinks(sinks []SinkConfig) error {
	for _, sink := range sinks {
		if sink.Type == "console" {
			if sink.Console != nil && sink.Console.Stream != "stderr" && sink.Console.Stream != "" {
				return fmt.Errorf("console sink must use stderr (stdout is forbidden), got: %s", sink.Console.Stream)
			}
		}
	}
	return nil
}

// isYAML checks if a file path indicates YAML format
func isYAML(path string) bool {
	return len(path) > 5 && (path[len(path)-5:] == ".yaml" || path[len(path)-4:] == ".yml")
}

// DefaultConfig returns a default logger configuration
func DefaultConfig(service string) *LoggerConfig {
	return &LoggerConfig{
		Profile:      ProfileSimple,
		DefaultLevel: "INFO",
		Service:      service,
		Environment:  "development",
		Sinks: []SinkConfig{
			{
				Type: "console",
				// Format intentionally omitted - applyDefaults will set based on profile
				Console: &ConsoleSinkConfig{
					Stream:   "stderr",
					Colorize: false,
				},
			},
		},
		StaticFields:     make(map[string]any),
		EnableCaller:     false,
		EnableStacktrace: false,
	}
}
