package logging

import (
	"testing"
)

type mockMiddleware struct {
	name  string
	order int
	fn    func(*LogEvent) *LogEvent
}

func (m *mockMiddleware) Process(event *LogEvent) *LogEvent {
	return m.fn(event)
}

func (m *mockMiddleware) Order() int {
	return m.order
}

func (m *mockMiddleware) Name() string {
	return m.name
}

func TestMiddlewarePipeline_Ordering(t *testing.T) {
	callOrder := []string{}

	m1 := &mockMiddleware{
		name:  "first",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "first")
			return e
		},
	}

	m2 := &mockMiddleware{
		name:  "second",
		order: 20,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "second")
			return e
		},
	}

	m3 := &mockMiddleware{
		name:  "third",
		order: 30,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "third")
			return e
		},
	}

	pipeline := NewMiddlewarePipeline([]Middleware{m3, m1, m2})

	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)

	if result == nil {
		t.Fatal("pipeline returned nil, expected event")
	}

	if len(callOrder) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(callOrder))
	}

	if callOrder[0] != "first" || callOrder[1] != "second" || callOrder[2] != "third" {
		t.Errorf("wrong order: %v, want [first, second, third]", callOrder)
	}
}

func TestMiddlewarePipeline_ShortCircuit(t *testing.T) {
	callOrder := []string{}

	m1 := &mockMiddleware{
		name:  "first",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "first")
			return e
		},
	}

	m2 := &mockMiddleware{
		name:  "dropper",
		order: 20,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "dropper")
			return nil
		},
	}

	m3 := &mockMiddleware{
		name:  "third",
		order: 30,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "third")
			return e
		},
	}

	pipeline := NewMiddlewarePipeline([]Middleware{m1, m2, m3})

	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)

	if result != nil {
		t.Error("pipeline should return nil after dropper")
	}

	if len(callOrder) != 2 {
		t.Fatalf("expected 2 calls (first, dropper), got %d: %v", len(callOrder), callOrder)
	}

	if callOrder[0] != "first" || callOrder[1] != "dropper" {
		t.Errorf("wrong order: %v, want [first, dropper]", callOrder)
	}
}

func TestMiddlewarePipeline_EventTransformation(t *testing.T) {
	m1 := &mockMiddleware{
		name:  "prefix",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			e.Message = "prefix:" + e.Message
			return e
		},
	}

	m2 := &mockMiddleware{
		name:  "suffix",
		order: 20,
		fn: func(e *LogEvent) *LogEvent {
			e.Message = e.Message + ":suffix"
			return e
		},
	}

	pipeline := NewMiddlewarePipeline([]Middleware{m1, m2})

	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)

	if result == nil {
		t.Fatal("pipeline returned nil, expected event")
	}

	expected := "prefix:test:suffix"
	if result.Message != expected {
		t.Errorf("Message = %q, want %q", result.Message, expected)
	}
}

func TestMiddlewareRegistry_RegisterAndCreate(t *testing.T) {
	registry := NewMiddlewareRegistry()

	factory := func(config map[string]any) (Middleware, error) {
		return &mockMiddleware{
			name:  "test",
			order: 10,
			fn: func(e *LogEvent) *LogEvent {
				return e
			},
		}, nil
	}

	registry.Register("test", factory)

	middleware, err := registry.Create("test", nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if middleware.Name() != "test" {
		t.Errorf("Name() = %q, want %q", middleware.Name(), "test")
	}

	if middleware.Order() != 10 {
		t.Errorf("Order() = %d, want %d", middleware.Order(), 10)
	}
}

func TestMiddlewareRegistry_CreateUnregistered(t *testing.T) {
	registry := NewMiddlewareRegistry()

	_, err := registry.Create("nonexistent", nil)
	if err == nil {
		t.Error("Create should fail for unregistered middleware")
	}

	if err.Error() != `middleware "nonexistent" not registered` {
		t.Errorf("wrong error message: %v", err)
	}
}

func TestMiddlewareRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewMiddlewareRegistry()

	factory := func(config map[string]any) (Middleware, error) {
		return &mockMiddleware{
			name:  "concurrent",
			order: 10,
			fn:    func(e *LogEvent) *LogEvent { return e },
		}, nil
	}

	registry.Register("concurrent", factory)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := registry.Create("concurrent", nil)
			if err != nil {
				t.Errorf("concurrent Create failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	if registry == nil {
		t.Error("DefaultRegistry() should return initialized registry")
	}

	r1 := DefaultRegistry()
	r2 := DefaultRegistry()
	if r1 != r2 {
		t.Error("DefaultRegistry() should return same singleton instance")
	}
}

func TestMiddlewarePipeline_EqualOrderStable(t *testing.T) {
	callOrder := []string{}

	m1 := &mockMiddleware{
		name:  "first",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "first")
			return e
		},
	}

	m2 := &mockMiddleware{
		name:  "second",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "second")
			return e
		},
	}

	m3 := &mockMiddleware{
		name:  "third",
		order: 10,
		fn: func(e *LogEvent) *LogEvent {
			callOrder = append(callOrder, "third")
			return e
		},
	}

	pipeline := NewMiddlewarePipeline([]Middleware{m1, m2, m3})

	event := &LogEvent{Message: "test"}
	result := pipeline.Process(event)

	if result == nil {
		t.Fatal("pipeline returned nil, expected event")
	}

	if len(callOrder) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(callOrder))
	}

	if callOrder[0] != "first" || callOrder[1] != "second" || callOrder[2] != "third" {
		t.Errorf("equal order should preserve original order: %v, want [first, second, third]", callOrder)
	}
}

func TestMiddlewareRegistry_NilConfig(t *testing.T) {
	registry := NewMiddlewareRegistry()

	factory := func(config map[string]any) (Middleware, error) {
		if config == nil {
			t.Error("factory should receive non-nil config (normalized)")
		}
		if len(config) != 0 {
			t.Errorf("normalized empty config should have length 0, got %d", len(config))
		}
		return &mockMiddleware{
			name:  "test",
			order: 10,
			fn:    func(e *LogEvent) *LogEvent { return e },
		}, nil
	}

	registry.Register("test", factory)

	_, err := registry.Create("test", nil)
	if err != nil {
		t.Fatalf("Create with nil config should work: %v", err)
	}
}
