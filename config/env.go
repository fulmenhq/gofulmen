package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fulmenhq/gofulmen/errors"
)

// EnvVarType describes how to parse an environment variable value.
type EnvVarType int

const (
	EnvString EnvVarType = iota
	EnvInt
	EnvFloat
	EnvBool
)

// EnvVarSpec maps an environment variable to a configuration path.
type EnvVarSpec struct {
	Name string
	Path []string
	Type EnvVarType
}

// LoadEnvOverrides builds a runtime override map from environment variables according to the provided specs.
func LoadEnvOverrides(specs []EnvVarSpec) (map[string]any, error) {
	return LoadEnvOverridesWithEnvelope(specs, "")
}

// LoadEnvOverridesWithEnvelope builds a runtime override map from environment variables with structured error reporting.
func LoadEnvOverridesWithEnvelope(specs []EnvVarSpec, correlationID string) (map[string]any, error) {
	overrides := make(map[string]any)
	for _, spec := range specs {
		if spec.Name == "" || len(spec.Path) == 0 {
			continue
		}
		value, ok := os.LookupEnv(spec.Name)
		if !ok {
			continue
		}
		parsed, err := parseEnvValue(value, spec.Type)
		if err != nil {
			envelope := errors.NewErrorEnvelope("CONFIG_ENV_PARSE_ERROR", fmt.Sprintf("Failed to parse environment variable %s", spec.Name))
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = envelope.WithCorrelationID(correlationID)
			envelope = errors.SafeWithContext(envelope, map[string]interface{}{
				"component":  "config",
				"operation":  "load_env_overrides",
				"error_type": "env_parse_error",
				"env_var":    spec.Name,
				"env_value":  value,
				"env_type":   envTypeToString(spec.Type),
			})
			envelope = envelope.WithOriginal(err)
			return nil, envelope
		}
		setNestedValue(overrides, spec.Path, parsed)
	}
	return overrides, nil
}

func envTypeToString(t EnvVarType) string {
	switch t {
	case EnvInt:
		return "int"
	case EnvFloat:
		return "float"
	case EnvBool:
		return "bool"
	default:
		return "string"
	}
}

func parseEnvValue(value string, t EnvVarType) (any, error) {
	switch t {
	case EnvInt:
		v, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid integer %q", value)
		}
		return v, nil
	case EnvFloat:
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q", value)
		}
		return v, nil
	case EnvBool:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "t", "true", "yes", "y":
			return true, nil
		case "0", "f", "false", "no", "n":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean %q", value)
		}
	default:
		return value, nil
	}
}

func setNestedValue(root map[string]any, path []string, value any) {
	if len(path) == 0 {
		return
	}
	current := root
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		child, ok := current[key].(map[string]any)
		if !ok {
			child = make(map[string]any)
			current[key] = child
		}
		current = child
	}
	current[path[len(path)-1]] = value
}
