package logging

import (
	"sync"
	"testing"
	"time"
)

func TestNewThrottlingMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		config     map[string]any
		wantOrder  int
		wantRate   int
		wantBurst  int
		wantPolicy string
	}{
		{
			name:       "defaults",
			config:     map[string]any{},
			wantOrder:  20,
			wantRate:   1000,
			wantBurst:  100,
			wantPolicy: DropPolicyOldest,
		},
		{
			name: "custom_int_values",
			config: map[string]any{
				"order":      15,
				"maxRate":    500,
				"burstSize":  50,
				"dropPolicy": DropPolicyNewest,
			},
			wantOrder:  15,
			wantRate:   500,
			wantBurst:  50,
			wantPolicy: DropPolicyNewest,
		},
		{
			name: "float64_values_from_json",
			config: map[string]any{
				"order":      float64(10),
				"maxRate":    float64(2000),
				"burstSize":  float64(200),
				"dropPolicy": DropPolicyBlock,
			},
			wantOrder:  10,
			wantRate:   2000,
			wantBurst:  200,
			wantPolicy: DropPolicyBlock,
		},
		{
			name: "partial_config",
			config: map[string]any{
				"maxRate": 300,
			},
			wantOrder:  20,
			wantRate:   300,
			wantBurst:  100,
			wantPolicy: DropPolicyOldest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := NewThrottlingMiddleware(tt.config)
			if err != nil {
				t.Fatalf("NewThrottlingMiddleware() error = %v", err)
			}

			throttle := mw.(*ThrottlingMiddleware)

			if throttle.Order() != tt.wantOrder {
				t.Errorf("Order() = %v, want %v", throttle.Order(), tt.wantOrder)
			}
			if throttle.maxRate != tt.wantRate {
				t.Errorf("maxRate = %v, want %v", throttle.maxRate, tt.wantRate)
			}
			if throttle.burstSize != tt.wantBurst {
				t.Errorf("burstSize = %v, want %v", throttle.burstSize, tt.wantBurst)
			}
			if throttle.dropPolicy != tt.wantPolicy {
				t.Errorf("dropPolicy = %v, want %v", throttle.dropPolicy, tt.wantPolicy)
			}
			if throttle.tokens != tt.wantBurst {
				t.Errorf("initial tokens = %v, want %v (should equal burstSize)", throttle.tokens, tt.wantBurst)
			}
		})
	}
}

func TestThrottlingMiddleware_Name(t *testing.T) {
	mw, _ := NewThrottlingMiddleware(map[string]any{})
	if mw.Name() != "throttle" {
		t.Errorf("Name() = %v, want 'throttle'", mw.Name())
	}
}

func TestThrottlingMiddleware_Process_NilEvent(t *testing.T) {
	mw, _ := NewThrottlingMiddleware(map[string]any{})
	result := mw.Process(nil)
	if result != nil {
		t.Errorf("Process(nil) = %v, want nil", result)
	}
}

func TestThrottlingMiddleware_Process_WithinBurst(t *testing.T) {
	config := map[string]any{
		"maxRate":   100,
		"burstSize": 10,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 10; i++ {
		event := &LogEvent{Message: "test"}
		result := throttle.Process(event)

		if result == nil {
			t.Fatalf("Event %d dropped within burst limit", i)
		}
		if result.ThrottleBucket != "allowed" {
			t.Errorf("Event %d: ThrottleBucket = %v, want 'allowed'", i, result.ThrottleBucket)
		}
	}

	if throttle.tokens != 0 {
		t.Errorf("After burst: tokens = %v, want 0", throttle.tokens)
	}
	if throttle.DroppedCount() != 0 {
		t.Errorf("DroppedCount = %v, want 0", throttle.DroppedCount())
	}
}

func TestThrottlingMiddleware_Process_DropOldest(t *testing.T) {
	config := map[string]any{
		"maxRate":    100,
		"burstSize":  5,
		"dropPolicy": DropPolicyOldest,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 5; i++ {
		event := &LogEvent{Message: "allowed"}
		if throttle.Process(event) == nil {
			t.Fatalf("Event %d dropped within burst", i)
		}
	}

	event := &LogEvent{Message: "should drop"}
	result := throttle.Process(event)

	if result != nil {
		t.Errorf("Process() = %v, want nil (dropped)", result)
	}
	if throttle.DroppedCount() != 1 {
		t.Errorf("DroppedCount = %v, want 1", throttle.DroppedCount())
	}
}

func TestThrottlingMiddleware_Process_DropNewest(t *testing.T) {
	config := map[string]any{
		"maxRate":    100,
		"burstSize":  3,
		"dropPolicy": DropPolicyNewest,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 3; i++ {
		throttle.Process(&LogEvent{Message: "allowed"})
	}

	event := &LogEvent{Message: "should drop"}
	result := throttle.Process(event)

	if result != nil {
		t.Errorf("Process() = %v, want nil (dropped)", result)
	}
	if throttle.DroppedCount() != 1 {
		t.Errorf("DroppedCount = %v, want 1", throttle.DroppedCount())
	}
}

func TestThrottlingMiddleware_Process_Block(t *testing.T) {
	config := map[string]any{
		"maxRate":    100,
		"burstSize":  2,
		"dropPolicy": DropPolicyBlock,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 2; i++ {
		throttle.Process(&LogEvent{Message: "allowed"})
	}

	event := &LogEvent{Message: "blocked"}
	result := throttle.Process(event)

	if result == nil {
		t.Errorf("Process() = nil, want non-nil (blocked event passes through)")
		return
	}
	if result.ThrottleBucket != "blocked" {
		t.Errorf("ThrottleBucket = %v, want 'blocked'", result.ThrottleBucket)
	}
	if throttle.DroppedCount() != 1 {
		t.Errorf("DroppedCount = %v, want 1 (counted as dropped)", throttle.DroppedCount())
	}
}

func TestThrottlingMiddleware_TokenRefill(t *testing.T) {
	config := map[string]any{
		"maxRate":   1000,
		"burstSize": 10,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 10; i++ {
		throttle.Process(&LogEvent{Message: "drain"})
	}

	if throttle.tokens != 0 {
		t.Fatalf("tokens = %v, want 0 after draining", throttle.tokens)
	}

	time.Sleep(100 * time.Millisecond)

	event := &LogEvent{Message: "after refill"}
	result := throttle.Process(event)

	if result == nil {
		t.Errorf("Process() after refill = nil, want allowed")
		return
	}
	if result.ThrottleBucket != "allowed" {
		t.Errorf("ThrottleBucket = %v, want 'allowed'", result.ThrottleBucket)
	}
}

func TestThrottlingMiddleware_RefillCapping(t *testing.T) {
	config := map[string]any{
		"maxRate":   10000,
		"burstSize": 100,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 50; i++ {
		throttle.Process(&LogEvent{Message: "half drain"})
	}

	time.Sleep(200 * time.Millisecond)

	throttle.mu.Lock()
	refillBefore := throttle.tokens
	throttle.mu.Unlock()

	if refillBefore < 50 {
		t.Fatalf("tokens before cap check = %v, want >= 50", refillBefore)
	}

	if refillBefore > throttle.burstSize {
		t.Errorf("tokens = %v exceeds burstSize %v (should cap)", refillBefore, throttle.burstSize)
	}
}

func TestThrottlingMiddleware_ResetStats(t *testing.T) {
	config := map[string]any{
		"maxRate":   100,
		"burstSize": 2,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	throttle.Process(&LogEvent{Message: "1"})
	throttle.Process(&LogEvent{Message: "2"})
	throttle.Process(&LogEvent{Message: "dropped"})

	if throttle.DroppedCount() != 1 {
		t.Fatalf("DroppedCount = %v, want 1 before reset", throttle.DroppedCount())
	}

	throttle.ResetStats()

	if throttle.DroppedCount() != 0 {
		t.Errorf("DroppedCount after reset = %v, want 0", throttle.DroppedCount())
	}
}

func TestThrottlingMiddleware_ConcurrentAccess(t *testing.T) {
	config := map[string]any{
		"maxRate":   10000,
		"burstSize": 1000,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	var wg sync.WaitGroup
	goroutines := 10
	eventsPerGoroutine := 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := &LogEvent{Message: "concurrent"}
				throttle.Process(event)
			}
		}(i)
	}

	wg.Wait()

	totalProcessed := goroutines * eventsPerGoroutine

	throttle.mu.Lock()
	actualTokensRemaining := throttle.tokens
	throttle.mu.Unlock()

	if actualTokensRemaining < 0 {
		t.Errorf("tokens = %v, should never be negative", actualTokensRemaining)
	}

	t.Logf("Processed %d events concurrently, dropped %d, tokens remaining: %d",
		totalProcessed, throttle.DroppedCount(), actualTokensRemaining)
}

func TestThrottlingMiddleware_RegistryIntegration(t *testing.T) {
	factory := DefaultRegistry().factories["throttle"]
	if factory == nil {
		t.Fatal("throttle middleware not registered in DefaultRegistry")
	}

	mw, err := factory(map[string]any{"maxRate": 500})
	if err != nil {
		t.Fatalf("factory() error = %v", err)
	}

	if mw.Name() != "throttle" {
		t.Errorf("factory created middleware with Name() = %v, want 'throttle'", mw.Name())
	}
}

func TestThrottlingMiddleware_ZeroRateHandling(t *testing.T) {
	config := map[string]any{
		"maxRate":   1,
		"burstSize": 5,
	}
	mw, _ := NewThrottlingMiddleware(config)
	throttle := mw.(*ThrottlingMiddleware)

	for i := 0; i < 5; i++ {
		event := &LogEvent{Message: "burst"}
		result := throttle.Process(event)
		if result == nil {
			t.Fatalf("Event %d dropped within burst", i)
		}
	}

	time.Sleep(10 * time.Millisecond)

	event := &LogEvent{Message: "after minimal wait"}
	result := throttle.Process(event)

	if result != nil && result.ThrottleBucket == "allowed" {
		t.Logf("Event allowed after %dms wait (tokens refilled)", 10)
	}
}
