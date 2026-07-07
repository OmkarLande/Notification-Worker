// Package providers contains the infrastructure for resolving app-specific
// notification providers at runtime. Actual provider implementations
// (e.g. stackday, expense) live in sub-packages and are registered with the
// factory during application startup.
package providers

import (
	"fmt"
	"sync"

	"github.com/OmkarLande/notification-worker/internal/interfaces"
)

// Registry is a thread-safe store that maps application names to their
// corresponding Provider implementations. It is the single source of truth for
// provider lookup within the application.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]interfaces.Provider
}

// NewRegistry returns an empty, ready-to-use Registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]interfaces.Provider),
	}
}

// Register associates a Provider with the given application name. The name
// must be unique; re-registering the same name returns an error to prevent
// silent overwrites caused by duplicate registrations.
func (r *Registry) Register(name string, p interfaces.Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider registry: provider %q is already registered", name)
	}

	r.providers[name] = p
	return nil
}

// Get retrieves the Provider registered under the given application name.
// It returns false as the second value if no provider is found, consistent
// with idiomatic Go map access.
func (r *Registry) Get(name string) (interfaces.Provider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.providers[name]
	return p, ok
}

// Names returns a sorted slice of all registered provider names. Useful for
// logging and diagnostics.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
