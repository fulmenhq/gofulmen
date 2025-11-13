package signals

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// HTTPConfig configures the HTTP admin signal endpoint.
type HTTPConfig struct {
	// TokenAuth is the bearer token for authentication.
	// If empty, token auth is disabled (not recommended for production).
	TokenAuth string

	// MTLSVerify enables mTLS certificate verification.
	// Requires properly configured TLS on the HTTP server.
	MTLSVerify bool

	// RateLimit is the maximum requests per minute.
	// Default: 6 requests per minute
	RateLimit int

	// RateBurst is the burst size for rate limiting.
	// Default: 3
	RateBurst int

	// Manager is the signal manager to dispatch signals to.
	// If nil, uses the default manager.
	Manager *Manager
}

// SignalRequest represents an HTTP signal request.
type SignalRequest struct {
	// Signal is the signal name (e.g., "SIGHUP", "SIGTERM").
	Signal string `json:"signal"`

	// Reason is an optional reason for sending the signal.
	Reason string `json:"reason,omitempty"`

	// GracePeriodSeconds is the grace period for shutdown signals.
	GracePeriodSeconds *int `json:"grace_period_seconds,omitempty"`

	// Requester is an optional identifier for audit logging.
	Requester string `json:"requester,omitempty"`
}

// SignalResponse represents an HTTP signal response.
type SignalResponse struct {
	// Success indicates whether the signal was processed successfully.
	Success bool `json:"success"`

	// Message is a human-readable message.
	Message string `json:"message"`

	// Signal is the signal that was processed.
	Signal string `json:"signal,omitempty"`

	// Error is an error message if processing failed.
	Error string `json:"error,omitempty"`
}

// HTTPHandler provides an HTTP handler for the /admin/signal endpoint.
type HTTPHandler struct {
	config      HTTPConfig
	manager     *Manager
	rateLimiter *rate.Limiter
}

// NewHTTPHandler creates a new HTTP signal handler with the given configuration.
//
// Example:
//
//	config := signals.HTTPConfig{
//	    TokenAuth:  os.Getenv("SIGNAL_ADMIN_TOKEN"),
//	    RateLimit:  6,
//	    RateBurst:  3,
//	}
//	handler := signals.NewHTTPHandler(config)
//	http.Handle("/admin/signal", handler)
func NewHTTPHandler(config HTTPConfig) *HTTPHandler {
	// Apply defaults
	if config.RateLimit == 0 {
		config.RateLimit = 6 // 6 requests per minute
	}
	if config.RateBurst == 0 {
		config.RateBurst = 3
	}
	if config.Manager == nil {
		config.Manager = GetDefaultManager()
	}

	// Create rate limiter (per-minute)
	limiter := rate.NewLimiter(rate.Limit(float64(config.RateLimit)/60.0), config.RateBurst)

	return &HTTPHandler{
		config:      config,
		manager:     config.Manager,
		rateLimiter: limiter,
	}
}

// ServeHTTP implements http.Handler for the signal endpoint.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Method check
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "only POST requests are allowed")
		return
	}

	// Rate limiting
	if !h.rateLimiter.Allow() {
		h.sendError(w, http.StatusTooManyRequests, "rate limit exceeded")
		return
	}

	// Authentication
	if !h.authenticate(r) {
		h.sendError(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	// Parse request
	var req SignalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Validate signal
	if req.Signal == "" {
		h.sendError(w, http.StatusBadRequest, "signal field is required")
		return
	}

	// Map signal name to os.Signal
	sig, err := h.parseSignal(req.Signal)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("invalid signal: %v", err))
		return
	}

	// Check if signal is supported
	if !Supports(sig) {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("signal %s is not supported on this platform", req.Signal))
		return
	}

	// Create context with grace period if specified
	ctx := r.Context()
	if req.GracePeriodSeconds != nil && *req.GracePeriodSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(*req.GracePeriodSeconds)*time.Second)
		defer cancel()
	}

	// Dispatch signal
	if err := h.manager.handleSignal(ctx, sig); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("signal processing failed: %v", err))
		return
	}

	// Success response
	h.sendSuccess(w, SignalResponse{
		Success: true,
		Message: fmt.Sprintf("signal %s processed successfully", req.Signal),
		Signal:  req.Signal,
	})
}

// authenticate checks the request authentication.
func (h *HTTPHandler) authenticate(r *http.Request) bool {
	// If no auth configured, allow (warning: not recommended for production)
	if h.config.TokenAuth == "" && !h.config.MTLSVerify {
		return true
	}

	// Token authentication
	if h.config.TokenAuth != "" {
		auth := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + h.config.TokenAuth
		if auth != expectedAuth {
			return false
		}
	}

	// mTLS verification (assumes TLS is configured on the server)
	if h.config.MTLSVerify {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			return false
		}
		// Additional certificate verification would go here
	}

	return true
}

// parseSignal converts a signal name to os.Signal.
func (h *HTTPHandler) parseSignal(name string) (os.Signal, error) {
	sig, ok := httpSignalMap[name]
	if !ok {
		return nil, fmt.Errorf("unknown signal: %s", name)
	}

	return sig, nil
}

var httpSignalMap = func() map[string]os.Signal {
	base := map[string]os.Signal{
		"SIGTERM": syscall.SIGTERM,
		"SIGINT":  syscall.SIGINT,
		"SIGHUP":  syscall.SIGHUP,
		"SIGQUIT": syscall.SIGQUIT,
	}

	for name, sig := range platformSpecificSignals {
		base[name] = sig
	}

	return base
}()

// sendError sends an error response.
func (h *HTTPHandler) sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SignalResponse{
		Success: false,
		Error:   message,
	})
}

// sendSuccess sends a success response.
func (h *HTTPHandler) sendSuccess(w http.ResponseWriter, resp SignalResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
