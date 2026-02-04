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

// EnvVarSpecWithAliases maps a canonical environment variable and any alias names
// to a configuration path.
//
// This type is intentionally separate from EnvVarSpec to avoid breaking
// downstream code that uses unkeyed EnvVarSpec literals.
type EnvVarSpecWithAliases struct {
	Name    string
	Aliases []string
	Path    []string
	Type    EnvVarType
}

// EnvVarSource indicates whether an override was sourced from a canonical name
// or an alias.
type EnvVarSource string

const (
	EnvVarSourceCanonical EnvVarSource = "canonical"
	EnvVarSourceAlias     EnvVarSource = "alias"
)

// EnvVarApplied records which environment variable name provided the effective
// value for a spec.
type EnvVarApplied struct {
	SpecName   string
	ChosenName string
	Source     EnvVarSource
	Path       []string
}

// EnvVarConflict indicates that multiple env vars for a spec were set with
// different values.
//
// Values are masked by default to avoid leaking secrets into logs/telemetry.
type EnvVarConflict struct {
	CanonicalName string
	AliasName     string
	Canonical     string
	Alias         string
	ChosenName    string
	ValueLen      int
	Masked        bool
}

// LoadEnvOverridesReport contains overrides plus diagnostics about which env vars
// were applied and any conflicts detected.
type LoadEnvOverridesReport struct {
	Overrides map[string]any
	Applied   []EnvVarApplied
	Conflicts []EnvVarConflict
}

// LoadEnvOverridesWithReport builds a runtime override map from environment
// variables according to the provided specs and returns diagnostics.
//
// Precedence: aliases override the canonical Name when both are set.
func LoadEnvOverridesWithReport(specs []EnvVarSpecWithAliases) (LoadEnvOverridesReport, error) {
	return LoadEnvOverridesWithEnvelopeAndReport(specs, "")
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

			display, masked := maskEnvValue(spec.Name, value)
			envelope = errors.SafeWithContext(envelope, map[string]interface{}{
				"component":  "config",
				"operation":  "load_env_overrides",
				"error_type": "env_parse_error",
				"env_var":    spec.Name,
				"env_value":  display,
				"env_masked": masked,
				"env_len":    len(value),
				"env_type":   envTypeToString(spec.Type),
			})
			envelope = envelope.WithOriginal(err)
			return nil, envelope
		}
		setNestedValue(overrides, spec.Path, parsed)
	}
	return overrides, nil
}

// LoadEnvOverridesWithEnvelopeAndReport builds a runtime override map from
// environment variables with structured error reporting and diagnostics.
func LoadEnvOverridesWithEnvelopeAndReport(specs []EnvVarSpecWithAliases, correlationID string) (LoadEnvOverridesReport, error) {
	overrides := make(map[string]any)
	var applied []EnvVarApplied
	var conflicts []EnvVarConflict

	for _, spec := range specs {
		if spec.Name == "" || len(spec.Path) == 0 {
			continue
		}

		canonicalValue, canonicalOK := os.LookupEnv(spec.Name)
		chosenName := ""
		chosenRaw := ""
		source := EnvVarSourceCanonical

		// Aliases take precedence when set. If multiple aliases are set, the first
		// one in spec.Aliases order wins.
		for _, alias := range spec.Aliases {
			if alias == "" {
				continue
			}
			v, ok := os.LookupEnv(alias)
			if !ok {
				continue
			}
			chosenName = alias
			chosenRaw = v
			source = EnvVarSourceAlias
			break
		}

		if chosenName == "" {
			if !canonicalOK {
				continue
			}
			chosenName = spec.Name
			chosenRaw = canonicalValue
			source = EnvVarSourceCanonical
		}

		if canonicalOK && source == EnvVarSourceAlias {
			if strings.TrimSpace(canonicalValue) != strings.TrimSpace(chosenRaw) {
				canonicalDisplay, canonicalMasked := maskEnvValue(spec.Name, canonicalValue)
				aliasDisplay, aliasMasked := maskEnvValue(chosenName, chosenRaw)
				masked := canonicalMasked || aliasMasked

				conflicts = append(conflicts, EnvVarConflict{
					CanonicalName: spec.Name,
					AliasName:     chosenName,
					Canonical:     canonicalDisplay,
					Alias:         aliasDisplay,
					ChosenName:    chosenName,
					ValueLen:      len(chosenRaw),
					Masked:        masked,
				})
			}
		}

		parsed, err := parseEnvValue(chosenRaw, spec.Type)
		if err != nil {
			envelope := errors.NewErrorEnvelope("CONFIG_ENV_PARSE_ERROR", fmt.Sprintf("Failed to parse environment variable %s", spec.Name))
			envelope = errors.SafeWithSeverity(envelope, errors.SeverityMedium)
			envelope = envelope.WithCorrelationID(correlationID)

			display, masked := maskEnvValue(chosenName, chosenRaw)
			envelope = errors.SafeWithContext(envelope, map[string]interface{}{
				"component":  "config",
				"operation":  "load_env_overrides",
				"error_type": "env_parse_error",
				"env_var":    chosenName,
				"env_value":  display,
				"env_masked": masked,
				"env_len":    len(chosenRaw),
				"env_type":   envTypeToString(spec.Type),
			})
			envelope = envelope.WithOriginal(err)
			return LoadEnvOverridesReport{}, envelope
		}
		setNestedValue(overrides, spec.Path, parsed)
		applied = append(applied, EnvVarApplied{
			SpecName:   spec.Name,
			ChosenName: chosenName,
			Source:     source,
			Path:       append([]string(nil), spec.Path...),
		})
	}

	return LoadEnvOverridesReport{
		Overrides: overrides,
		Applied:   applied,
		Conflicts: conflicts,
	}, nil
}

func isSensitiveEnvName(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	upper := strings.ToUpper(name)
	for _, needle := range []string{
		"TOKEN",
		"SECRET",
		"PASSWORD",
		"PASSWD",
		"PWD",
		"API_KEY",
		"PRIVATE_KEY",
		"CREDENTIAL",
		"AUTHORIZATION",
	} {
		if strings.Contains(upper, needle) {
			return true
		}
	}

	return false
}

func maskEnvValue(envName, raw string) (string, bool) {
	if isSensitiveEnvName(envName) {
		return "[set]", true
	}
	return strings.TrimSpace(raw), false
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
