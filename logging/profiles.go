package logging

import "fmt"

type LoggingProfile string

const (
	ProfileSimple     LoggingProfile = "SIMPLE"
	ProfileStructured LoggingProfile = "STRUCTURED"
	ProfileEnterprise LoggingProfile = "ENTERPRISE"
	ProfileCustom     LoggingProfile = "CUSTOM"
)

type ProfileRequirements struct {
	RequiredSinks      []string
	AllowedFormats     []string
	MaxMiddleware      *int
	MinMiddleware      *int
	ThrottlingAllowed  bool
	ThrottlingRequired bool
	PolicyEnforcement  bool
}

func GetProfileRequirements(profile LoggingProfile) ProfileRequirements {
	switch profile {
	case ProfileSimple:
		maxMW := 0
		minMW := 0
		return ProfileRequirements{
			RequiredSinks:      []string{"console"},
			AllowedFormats:     []string{"console", "text"},
			MaxMiddleware:      &maxMW,
			MinMiddleware:      &minMW,
			ThrottlingAllowed:  false,
			ThrottlingRequired: false,
			PolicyEnforcement:  false,
		}

	case ProfileStructured:
		maxMW := 2
		return ProfileRequirements{
			RequiredSinks:      nil,
			AllowedFormats:     []string{"json", "text", "console"},
			MaxMiddleware:      &maxMW,
			MinMiddleware:      nil,
			ThrottlingAllowed:  true,
			ThrottlingRequired: false,
			PolicyEnforcement:  false,
		}

	case ProfileEnterprise:
		minMW := 1
		return ProfileRequirements{
			RequiredSinks:      nil,
			AllowedFormats:     []string{"json"},
			MaxMiddleware:      nil,
			MinMiddleware:      &minMW,
			ThrottlingAllowed:  true,
			ThrottlingRequired: true,
			PolicyEnforcement:  true,
		}

	case ProfileCustom:
		return ProfileRequirements{
			RequiredSinks:      nil,
			AllowedFormats:     nil,
			MaxMiddleware:      nil,
			MinMiddleware:      nil,
			ThrottlingAllowed:  true,
			ThrottlingRequired: false,
			PolicyEnforcement:  false,
		}

	default:
		return ProfileRequirements{}
	}
}

func ValidateProfileRequirements(
	profile LoggingProfile,
	sinks []SinkConfig,
	middleware []MiddlewareConfig,
	format string,
	throttlingEnabled bool,
	policyEnabled bool,
) []error {
	reqs := GetProfileRequirements(profile)
	var errors []error

	if len(reqs.RequiredSinks) > 0 {
		sinkTypes := make(map[string]bool)
		for _, sink := range sinks {
			sinkTypes[sink.Type] = true
		}

		for _, required := range reqs.RequiredSinks {
			if !sinkTypes[required] {
				errors = append(errors, fmt.Errorf("profile %s requires sink type: %s", profile, required))
			}
		}
	}

	if len(reqs.AllowedFormats) > 0 {
		// Validate global format if specified
		if format != "" {
			allowed := false
			for _, allowedFormat := range reqs.AllowedFormats {
				if format == allowedFormat {
					allowed = true
					break
				}
			}
			if !allowed {
				errors = append(errors, fmt.Errorf("profile %s does not allow format: %s (allowed: %v)", profile, format, reqs.AllowedFormats))
			}
		}

		// Validate per-sink formats
		for i, sink := range sinks {
			if sink.Format == "" {
				continue
			}
			allowed := false
			for _, allowedFormat := range reqs.AllowedFormats {
				if sink.Format == allowedFormat {
					allowed = true
					break
				}
			}
			if !allowed {
				errors = append(errors, fmt.Errorf("profile %s does not allow format '%s' on sink %d (type: %s, allowed: %v)", profile, sink.Format, i, sink.Type, reqs.AllowedFormats))
			}
		}
	}

	middlewareCount := len(middleware)
	if reqs.MinMiddleware != nil && middlewareCount < *reqs.MinMiddleware {
		errors = append(errors, fmt.Errorf("profile %s requires at least %d middleware (got %d)", profile, *reqs.MinMiddleware, middlewareCount))
	}

	if reqs.MaxMiddleware != nil && middlewareCount > *reqs.MaxMiddleware {
		errors = append(errors, fmt.Errorf("profile %s allows at most %d middleware (got %d)", profile, *reqs.MaxMiddleware, middlewareCount))
	}

	if !reqs.ThrottlingAllowed && throttlingEnabled {
		errors = append(errors, fmt.Errorf("profile %s does not allow throttling", profile))
	}

	if reqs.ThrottlingRequired && !throttlingEnabled {
		errors = append(errors, fmt.Errorf("profile %s requires throttling to be enabled", profile))
	}

	if reqs.PolicyEnforcement && !policyEnabled {
		errors = append(errors, fmt.Errorf("profile %s requires policy enforcement", profile))
	}

	return errors
}
