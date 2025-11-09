package signals

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPHandler(t *testing.T) {
	config := HTTPConfig{
		TokenAuth: "test-token",
		RateLimit: 10,
		RateBurst: 5,
	}

	handler := NewHTTPHandler(config)
	assert.NotNil(t, handler, "Handler should not be nil")
	assert.Equal(t, "test-token", handler.config.TokenAuth)
	assert.NotNil(t, handler.rateLimiter, "Rate limiter should be initialized")
}

func TestNewHTTPHandler_Defaults(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	assert.Equal(t, 6, handler.config.RateLimit, "Default rate limit should be 6")
	assert.Equal(t, 3, handler.config.RateBurst, "Default rate burst should be 3")
	assert.NotNil(t, handler.manager, "Default manager should be set")
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	req := httptest.NewRequest(http.MethodGet, "/admin/signal", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var resp SignalResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "POST")
}

func TestHTTPHandler_MissingAuth(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{
		TokenAuth: "secret-token",
	})

	body := SignalRequest{Signal: "SIGTERM"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHTTPHandler_ValidAuth(t *testing.T) {
	m := NewManager()
	var signalReceived bool

	m.OnShutdown(func(ctx context.Context) error {
		signalReceived = true
		return nil
	})

	handler := NewHTTPHandler(HTTPConfig{
		TokenAuth: "secret-token",
		Manager:   m,
	})

	body := SignalRequest{Signal: "SIGTERM"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SignalResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "SIGTERM", resp.Signal)
	assert.True(t, signalReceived, "Signal handler should have been called")
}

func TestHTTPHandler_InvalidJSON(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp SignalResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid request body")
}

func TestHTTPHandler_MissingSignal(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	body := SignalRequest{} // Missing signal
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp SignalResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "signal field is required")
}

func TestHTTPHandler_UnknownSignal(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	body := SignalRequest{Signal: "SIGINVALID"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp SignalResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid signal")
}

func TestHTTPHandler_RateLimit(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{
		RateLimit: 1, // 1 per minute
		RateBurst: 1, // Burst of 1
	})

	body := SignalRequest{Signal: "SIGTERM"}
	bodyBytes, _ := json.Marshal(body)

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second immediate request should be rate limited
	req2 := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestHTTPHandler_GracePeriod(t *testing.T) {
	m := NewManager()
	var ctxReceived context.Context

	m.OnShutdown(func(ctx context.Context) error {
		ctxReceived = ctx
		return nil
	})

	handler := NewHTTPHandler(HTTPConfig{
		Manager: m,
	})

	gracePeriod := 5
	body := SignalRequest{
		Signal:             "SIGTERM",
		GracePeriodSeconds: &gracePeriod,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/signal", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotNil(t, ctxReceived, "Context should have been passed to handler")

	// Verify context has deadline
	deadline, ok := ctxReceived.Deadline()
	assert.True(t, ok, "Context should have a deadline")
	assert.True(t, time.Until(deadline) > 0, "Deadline should be in the future")
	assert.True(t, time.Until(deadline) <= 5*time.Second, "Deadline should be within grace period")
}

func TestParseSignal(t *testing.T) {
	handler := NewHTTPHandler(HTTPConfig{})

	tests := []struct {
		name        string
		signalName  string
		shouldError bool
	}{
		{"SIGTERM", "SIGTERM", false},
		{"SIGINT", "SIGINT", false},
		{"SIGHUP", "SIGHUP", false},
		{"SIGQUIT", "SIGQUIT", false},
		{"SIGUSR1", "SIGUSR1", false},
		{"SIGUSR2", "SIGUSR2", false},
		{"Invalid", "SIGINVALID", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := handler.parseSignal(tt.signalName)

			if tt.shouldError {
				assert.Error(t, err, "Should return error for invalid signal")
				assert.Nil(t, sig, "Signal should be nil for invalid name")
			} else {
				assert.NoError(t, err, "Should not return error for valid signal")
				assert.NotNil(t, sig, "Signal should not be nil")
			}
		})
	}
}
