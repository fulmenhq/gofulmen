package exporters

import (
	"time"
)

// PrometheusConfig holds configuration for the Prometheus exporter
type PrometheusConfig struct {
	// Prefix is prepended to all metric names (e.g., "myapp" -> "myapp_metric_name")
	Prefix string

	// Endpoint is the HTTP address to listen on (e.g., ":9090")
	Endpoint string

	// BearerToken for HTTP authentication (optional, empty = no auth)
	BearerToken string

	// RateLimit sets the maximum requests per minute (0 = no limit)
	// Default: 60 requests/minute with burst of 10
	RateLimitPerMinute int

	// RateLimitBurst sets the burst size for rate limiting
	// Default: 10
	RateLimitBurst int

	// RefreshInterval sets how often to refresh the registry (0 = on-demand only)
	// Default: 0 (immediate refresh on metric emission)
	RefreshInterval time.Duration

	// QuietMode suppresses HTTP request logging to stderr
	// Default: false
	QuietMode bool

	// ReadHeaderTimeout prevents Slowloris attacks
	// Default: 10 seconds
	ReadHeaderTimeout time.Duration
}

// DefaultPrometheusConfig returns sensible defaults for Prometheus exporter
func DefaultPrometheusConfig() *PrometheusConfig {
	return &PrometheusConfig{
		Prefix:             "",
		Endpoint:           ":9090",
		BearerToken:        "",
		RateLimitPerMinute: 60,
		RateLimitBurst:     10,
		RefreshInterval:    0, // Immediate refresh on emission
		QuietMode:          false,
		ReadHeaderTimeout:  10 * time.Second,
	}
}

// Validate checks configuration values and returns an error if invalid
func (c *PrometheusConfig) Validate() error {
	if c.Endpoint == "" {
		c.Endpoint = ":9090"
	}
	if c.RateLimitPerMinute < 0 {
		c.RateLimitPerMinute = 0
	}
	if c.RateLimitBurst < 0 {
		c.RateLimitBurst = 0
	}
	if c.ReadHeaderTimeout <= 0 {
		c.ReadHeaderTimeout = 10 * time.Second
	}
	return nil
}
