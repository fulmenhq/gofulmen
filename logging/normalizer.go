package logging

import (
	"fmt"
	"strings"
)

type NormalizationResult struct {
	Config   *LoggerConfig
	Warnings []string
}

func NormalizeLoggerConfig(config *LoggerConfig) (*NormalizationResult, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	result := &NormalizationResult{
		Config:   config,
		Warnings: []string{},
	}

	normalizeProfile(config, result)
	normalizeMiddleware(config, result)
	normalizeThrottling(config, result)
	applyProfileDefaults(config, result)

	return result, nil
}

func normalizeProfile(config *LoggerConfig, result *NormalizationResult) {
	profileStr := string(config.Profile)
	upperProfile := strings.ToUpper(profileStr)

	validProfiles := map[string]LoggingProfile{
		"SIMPLE":     ProfileSimple,
		"STRUCTURED": ProfileStructured,
		"ENTERPRISE": ProfileEnterprise,
		"CUSTOM":     ProfileCustom,
	}

	if normalized, ok := validProfiles[upperProfile]; ok {
		if config.Profile != normalized {
			result.Warnings = append(result.Warnings, fmt.Sprintf("normalized profile '%s' to '%s'", config.Profile, normalized))
			config.Profile = normalized
		}
	} else if config.Profile == "" {
		config.Profile = ProfileSimple
		result.Warnings = append(result.Warnings, "no profile specified, defaulting to SIMPLE")
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("invalid profile '%s', will fail schema validation (valid: SIMPLE, STRUCTURED, ENTERPRISE, CUSTOM)", config.Profile))
	}
}

func normalizeMiddleware(config *LoggerConfig, result *NormalizationResult) {
	if len(config.Middleware) == 0 {
		return
	}

	seen := make(map[string]int)
	deduplicated := []MiddlewareConfig{}

	for i := range config.Middleware {
		mw := &config.Middleware[i]

		if mw.Config == nil {
			mw.Config = make(map[string]any)
		}

		if idx, exists := seen[mw.Name]; exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("duplicate middleware '%s', keeping last definition", mw.Name))
			deduplicated[idx] = *mw
		} else {
			seen[mw.Name] = len(deduplicated)
			deduplicated = append(deduplicated, *mw)
		}
	}

	config.Middleware = deduplicated
}

func normalizeThrottling(config *LoggerConfig, result *NormalizationResult) {
	if config.Throttling == nil {
		return
	}

	if config.Throttling.MaxRate == 0 {
		config.Throttling.MaxRate = 1000
		result.Warnings = append(result.Warnings, "throttling maxRate not set, defaulting to 1000 logs/second")
	}

	if config.Throttling.BurstSize == 0 {
		config.Throttling.BurstSize = config.Throttling.MaxRate * 2
		result.Warnings = append(result.Warnings, fmt.Sprintf("throttling burstSize not set, defaulting to %d (2x maxRate)", config.Throttling.BurstSize))
	}

	if config.Throttling.DropPolicy == "" {
		config.Throttling.DropPolicy = "drop-oldest"
		result.Warnings = append(result.Warnings, "throttling dropPolicy not set, defaulting to 'drop-oldest'")
	}

	validDropPolicies := map[string]bool{
		"drop-oldest": true,
		"drop-newest": true,
		"block":       true,
	}
	if !validDropPolicies[config.Throttling.DropPolicy] {
		result.Warnings = append(result.Warnings, fmt.Sprintf("invalid dropPolicy '%s', using 'drop-oldest'", config.Throttling.DropPolicy))
		config.Throttling.DropPolicy = "drop-oldest"
	}
}

func applyProfileDefaults(config *LoggerConfig, result *NormalizationResult) {
	switch config.Profile {
	case ProfileSimple:
		if len(config.Sinks) == 0 {
			config.Sinks = []SinkConfig{
				{
					Type:   "console",
					Format: "console",
					Console: &ConsoleSinkConfig{
						Stream:   "stderr",
						Colorize: false,
					},
				},
			}
			result.Warnings = append(result.Warnings, "SIMPLE profile: added default console sink")
		}

		if config.Throttling != nil && config.Throttling.Enabled {
			config.Throttling.Enabled = false
			result.Warnings = append(result.Warnings, "SIMPLE profile: disabled throttling (not allowed)")
		}

		if len(config.Middleware) > 0 {
			config.Middleware = []MiddlewareConfig{}
			result.Warnings = append(result.Warnings, "SIMPLE profile: removed middleware (not allowed)")
		}

	case ProfileStructured:
		if config.Throttling != nil && !config.Throttling.Enabled {
			result.Warnings = append(result.Warnings, "STRUCTURED profile: throttling disabled but allowed")
		}

		if len(config.Middleware) > 2 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("STRUCTURED profile: has %d middleware (max 2 recommended)", len(config.Middleware)))
		}

	case ProfileEnterprise:
		if config.Throttling == nil {
			config.Throttling = &ThrottlingConfig{
				Enabled:    true,
				MaxRate:    1000,
				BurstSize:  2000,
				DropPolicy: "drop-oldest",
			}
			result.Warnings = append(result.Warnings, "ENTERPRISE profile: added required throttling with defaults")
		} else if !config.Throttling.Enabled {
			config.Throttling.Enabled = true
			result.Warnings = append(result.Warnings, "ENTERPRISE profile: enabled required throttling")
		}

		enabledCount := 0
		for _, mw := range config.Middleware {
			if mw.Enabled {
				enabledCount++
			}
		}
		if enabledCount == 0 {
			config.Middleware = append(config.Middleware, MiddlewareConfig{
				Name:    "correlation",
				Enabled: true,
				Order:   100,
				Config:  make(map[string]any),
			})
			result.Warnings = append(result.Warnings, "ENTERPRISE profile: added default correlation middleware (required min 1 enabled)")
		}

		if config.PolicyFile == "" {
			result.Warnings = append(result.Warnings, "ENTERPRISE profile: no policyFile specified (policy enforcement required)")
		}

	case ProfileCustom:
	}
}
