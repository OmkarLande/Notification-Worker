package channels

import (
	"fmt"
	"sync"
)

// Registry acts as a central store for initialized channel implementations.
type Registry struct {
	mu       sync.RWMutex
	channels map[string]Channel
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		channels: make(map[string]Channel),
	}
}

// Register adds a configured Channel to the registry.
func (r *Registry) Register(c Channel) error {
	if c == nil {
		return fmt.Errorf("channel registry: cannot register nil channel")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	name := c.Name()
	if _, exists := r.channels[name]; exists {
		return fmt.Errorf("channel registry: channel %q already registered", name)
	}
	r.channels[name] = c
	return nil
}

// Get retrieves a Channel by name.
func (r *Registry) Get(name string) (Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, exists := r.channels[name]
	if !exists {
		return nil, fmt.Errorf("channel registry: channel %q not found", name)
	}
	return c, nil
}
