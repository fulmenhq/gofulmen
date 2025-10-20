package logging

import (
	"fmt"
	"sort"
	"sync"
)

type Middleware interface {
	Process(event *LogEvent) *LogEvent
	Order() int
	Name() string
}

func normalizeConfig(config map[string]any) map[string]any {
	if config == nil {
		return make(map[string]any)
	}
	return config
}

type MiddlewarePipeline struct {
	middleware []Middleware
}

func NewMiddlewarePipeline(middleware []Middleware) *MiddlewarePipeline {
	sorted := make([]Middleware, len(middleware))
	copy(sorted, middleware)

	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Order() < sorted[j].Order()
	})

	return &MiddlewarePipeline{
		middleware: sorted,
	}
}

func (p *MiddlewarePipeline) Process(event *LogEvent) *LogEvent {
	current := event
	for _, m := range p.middleware {
		current = m.Process(current)
		if current == nil {
			return nil
		}
	}
	return current
}

type MiddlewareFactory func(config map[string]any) (Middleware, error)

type MiddlewareRegistry struct {
	mu        sync.RWMutex
	factories map[string]MiddlewareFactory
}

func NewMiddlewareRegistry() *MiddlewareRegistry {
	return &MiddlewareRegistry{
		factories: make(map[string]MiddlewareFactory),
	}
}

func (r *MiddlewareRegistry) Register(name string, factory MiddlewareFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = factory
}

func (r *MiddlewareRegistry) Create(name string, config map[string]any) (Middleware, error) {
	r.mu.RLock()
	factory, ok := r.factories[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("middleware %q not registered", name)
	}

	return factory(normalizeConfig(config))
}

var defaultRegistry = NewMiddlewareRegistry()

func DefaultRegistry() *MiddlewareRegistry {
	return defaultRegistry
}
