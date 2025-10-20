package logging

import (
	"sync"
	"time"
)

// ThrottlingMiddleware implements rate limiting using a token bucket algorithm.
//
// Controls log output rate to protect downstream sinks from being overwhelmed.
// Thread-safe for concurrent use.
type ThrottlingMiddleware struct {
	order      int
	maxRate    int
	burstSize  int
	dropPolicy string

	mu            sync.Mutex
	tokens        int
	lastRefill    time.Time
	droppedEvents int
}

const (
	DropPolicyOldest = "drop-oldest"
	DropPolicyNewest = "drop-newest"
	DropPolicyBlock  = "block"
)

// NewThrottlingMiddleware creates a new throttling middleware instance.
//
// Config options:
//   - order (int): Execution order in pipeline (default: 20)
//   - maxRate (int): Maximum events per second (default: 1000)
//   - burstSize (int): Burst capacity (default: 100)
//   - dropPolicy (string): "drop-oldest", "drop-newest", "block" (default: "drop-oldest")
func NewThrottlingMiddleware(config map[string]any) (Middleware, error) {
	order := 20
	if configOrder, ok := config["order"].(int); ok {
		order = configOrder
	} else if configOrder, ok := config["order"].(float64); ok {
		order = int(configOrder)
	}

	maxRate := 1000
	if configRate, ok := config["maxRate"].(int); ok {
		maxRate = configRate
	} else if configRate, ok := config["maxRate"].(float64); ok {
		maxRate = int(configRate)
	}

	burstSize := 100
	if configBurst, ok := config["burstSize"].(int); ok {
		burstSize = configBurst
	} else if configBurst, ok := config["burstSize"].(float64); ok {
		burstSize = int(configBurst)
	}

	dropPolicy := DropPolicyOldest
	if configPolicy, ok := config["dropPolicy"].(string); ok {
		dropPolicy = configPolicy
	}

	return &ThrottlingMiddleware{
		order:      order,
		maxRate:    maxRate,
		burstSize:  burstSize,
		dropPolicy: dropPolicy,
		tokens:     burstSize,
		lastRefill: time.Now(),
	}, nil
}

// Process applies rate limiting to log events.
//
// Returns nil if event should be dropped per configured policy.
// Thread-safe for concurrent access.
func (m *ThrottlingMiddleware) Process(event *LogEvent) *LogEvent {
	if event == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.refillTokens()

	if m.tokens > 0 {
		m.tokens--
		event.ThrottleBucket = "allowed"
		return event
	}

	m.droppedEvents++

	switch m.dropPolicy {
	case DropPolicyNewest:
		event.ThrottleBucket = "dropped-newest"
		return nil
	case DropPolicyBlock:
		event.ThrottleBucket = "blocked"
		return event
	default:
		event.ThrottleBucket = "dropped-oldest"
		return nil
	}
}

// refillTokens adds tokens based on elapsed time since last refill.
//
// Called with mutex held.
func (m *ThrottlingMiddleware) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(m.lastRefill)

	tokensToAdd := int(elapsed.Seconds() * float64(m.maxRate))

	if tokensToAdd > 0 {
		m.tokens += tokensToAdd
		if m.tokens > m.burstSize {
			m.tokens = m.burstSize
		}
		m.lastRefill = now
	}
}

// DroppedCount returns the number of events dropped by throttling.
//
// Thread-safe for concurrent access.
func (m *ThrottlingMiddleware) DroppedCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.droppedEvents
}

// ResetStats resets throttling statistics.
//
// Thread-safe for concurrent access.
func (m *ThrottlingMiddleware) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.droppedEvents = 0
}

func (m *ThrottlingMiddleware) Order() int {
	return m.order
}

func (m *ThrottlingMiddleware) Name() string {
	return "throttle"
}

func init() {
	DefaultRegistry().Register("throttle", NewThrottlingMiddleware)
}
