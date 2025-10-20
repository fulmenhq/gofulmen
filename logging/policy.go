package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

type LoggingPolicy struct {
	AllowedProfiles     []LoggingProfile                      `json:"allowedProfiles" yaml:"allowedProfiles"`
	RequiredProfiles    map[string][]LoggingProfile           `json:"requiredProfiles" yaml:"requiredProfiles"`
	EnvironmentRules    map[string][]LoggingProfile           `json:"environmentRules" yaml:"environmentRules"`
	ProfileRequirements map[LoggingProfile]ProfileConstraints `json:"profileRequirements" yaml:"profileRequirements"`
	AuditSettings       PolicyAuditSettings                   `json:"auditSettings" yaml:"auditSettings"`
}

type ProfileConstraints struct {
	MinEnvironment    string   `json:"minEnvironment,omitempty" yaml:"minEnvironment,omitempty"`
	MaxEnvironment    string   `json:"maxEnvironment,omitempty" yaml:"maxEnvironment,omitempty"`
	RequiredFeatures  []string `json:"requiredFeatures,omitempty" yaml:"requiredFeatures,omitempty"`
	ForbiddenFeatures []string `json:"forbiddenFeatures,omitempty" yaml:"forbiddenFeatures,omitempty"`
}

type PolicyAuditSettings struct {
	LogPolicyViolations bool `json:"logPolicyViolations" yaml:"logPolicyViolations"`
	EnforceStrictMode   bool `json:"enforceStrictMode" yaml:"enforceStrictMode"`
	RequirePolicyFile   bool `json:"requirePolicyFile" yaml:"requirePolicyFile"`
}

func LoadPolicy(policyFile string) (*LoggingPolicy, error) {
	if policyFile == "" {
		return nil, fmt.Errorf("policy file path required")
	}

	orgPath := os.Getenv("FULMEN_ORG_PATH")
	if orgPath == "" {
		orgPath = "/opt/fulmen"
	}

	searchPaths := []string{
		policyFile,
		".fulmen/logging-policy.yaml",
		"/etc/fulmen/logging-policy.yaml",
		filepath.Join(orgPath, "logging-policy.yaml"),
		"/org/fulmen/logging-policy.yaml",
	}

	var lastErr error
	for _, path := range searchPaths {
		policy, err := loadPolicyFromPath(path)
		if err == nil {
			return policy, nil
		}
		lastErr = err
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load policy from %s: %w", path, err)
		}
	}

	return nil, fmt.Errorf("policy file not found in any search path: %w", lastErr)
}

func loadPolicyFromPath(path string) (*LoggingPolicy, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- Policy path is from predefined search paths
	if err != nil {
		return nil, err
	}

	if err := validatePolicySchema(data, path); err != nil {
		return nil, fmt.Errorf("policy schema validation failed: %w", err)
	}

	var policy LoggingPolicy
	if isYAML(path) {
		if err := yaml.Unmarshal(data, &policy); err != nil {
			return nil, fmt.Errorf("invalid YAML policy file: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &policy); err != nil {
			return nil, fmt.Errorf("invalid JSON policy file: %w", err)
		}
	}

	return &policy, nil
}

func validatePolicySchema(data []byte, path string) error {
	catalog := schema.DefaultCatalog()
	schemaID := "observability/logging/v1.0.0/logging-policy"

	validator, err := catalog.ValidatorByID(schemaID)
	if err != nil {
		return fmt.Errorf("failed to load policy schema: %w", err)
	}

	var diagnostics []schema.Diagnostic
	if isYAML(path) {
		var yamlData interface{}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
		diagnostics, err = validator.ValidateData(yamlData)
	} else {
		diagnostics, err = validator.ValidateJSON(data)
	}

	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(diagnostics) > 0 {
		return fmt.Errorf("policy validation failed with %d error(s): %s", len(diagnostics), diagnostics[0].Message)
	}

	return nil
}

func ValidateConfigAgainstPolicy(
	config *LoggerConfig,
	policy *LoggingPolicy,
	environment string,
	appType string,
) []error {
	var errors []error

	if len(policy.AllowedProfiles) > 0 {
		allowed := false
		for _, allowedProfile := range policy.AllowedProfiles {
			if config.Profile == allowedProfile {
				allowed = true
				break
			}
		}
		if !allowed {
			errors = append(errors, fmt.Errorf("profile %s not in allowed profiles: %v", config.Profile, policy.AllowedProfiles))
		}
	}

	if appType != "" {
		if requiredProfiles, ok := policy.RequiredProfiles[appType]; ok {
			required := false
			for _, reqProfile := range requiredProfiles {
				if config.Profile == reqProfile {
					required = true
					break
				}
			}
			if !required {
				errors = append(errors, fmt.Errorf("app type %s requires one of profiles: %v (got %s)", appType, requiredProfiles, config.Profile))
			}
		}
	}

	if environment != "" {
		if envProfiles, ok := policy.EnvironmentRules[environment]; ok {
			allowed := false
			for _, envProfile := range envProfiles {
				if config.Profile == envProfile {
					allowed = true
					break
				}
			}
			if !allowed {
				errors = append(errors, fmt.Errorf("environment %s requires one of profiles: %v (got %s)", environment, envProfiles, config.Profile))
			}
		}
	}

	if constraints, ok := policy.ProfileRequirements[config.Profile]; ok {
		if constraints.MinEnvironment != "" && compareEnvironments(environment, constraints.MinEnvironment) < 0 {
			errors = append(errors, fmt.Errorf("profile %s requires minimum environment %s (got %s)", config.Profile, constraints.MinEnvironment, environment))
		}
		if constraints.MaxEnvironment != "" && compareEnvironments(environment, constraints.MaxEnvironment) > 0 {
			errors = append(errors, fmt.Errorf("profile %s requires maximum environment %s (got %s)", config.Profile, constraints.MaxEnvironment, environment))
		}

		for _, feature := range constraints.RequiredFeatures {
			if !hasFeature(config, feature) {
				errors = append(errors, fmt.Errorf("profile %s requires feature: %s", config.Profile, feature))
			}
		}

		for _, feature := range constraints.ForbiddenFeatures {
			if hasFeature(config, feature) {
				errors = append(errors, fmt.Errorf("profile %s forbids feature: %s", config.Profile, feature))
			}
		}
	}

	return errors
}

func EnforcePolicy(
	config *LoggerConfig,
	policy *LoggingPolicy,
	environment string,
	appType string,
) error {
	violations := ValidateConfigAgainstPolicy(config, policy, environment, appType)

	if len(violations) == 0 {
		return nil
	}

	if policy.AuditSettings.LogPolicyViolations {
		for _, violation := range violations {
			fmt.Fprintf(os.Stderr, "WARNING: Policy violation: %s\n", violation)
		}
	}

	if policy.AuditSettings.EnforceStrictMode {
		var messages []string
		for _, v := range violations {
			messages = append(messages, v.Error())
		}
		return fmt.Errorf("policy violations in strict mode (%d violations): %s", len(violations), strings.Join(messages, "; "))
	}

	return nil
}

func compareEnvironments(a, b string) int {
	envRank := map[string]int{
		"development": 1,
		"staging":     2,
		"production":  3,
	}

	rankA, okA := envRank[a]
	rankB, okB := envRank[b]

	if !okA || !okB {
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	}

	return rankA - rankB
}

func hasFeature(config *LoggerConfig, feature string) bool {
	switch feature {
	case "middleware":
		return len(config.Middleware) > 0
	case "throttling":
		return config.Throttling != nil && config.Throttling.Enabled
	case "policy":
		return config.PolicyFile != ""
	case "caller":
		return config.EnableCaller
	case "stacktrace":
		return config.EnableStacktrace
	default:
		return false
	}
}
