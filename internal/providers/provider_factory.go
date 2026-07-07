package providers

import (
	"fmt"

	"github.com/OmkarLande/notification-worker/internal/interfaces"
)

// Factory wraps a Registry and provides the primary API for resolving and
// registering providers throughout the application. New providers can be
// added by calling Register without modifying any existing code.
type Factory struct {
	registry *Registry
}

// NewFactory returns a Factory backed by the given Registry.
func NewFactory(r *Registry) *Factory {
	return &Factory{registry: r}
}

// Get resolves and returns the Provider registered under the given application
// name (e.g. "stackday", "expense"). It returns a descriptive error if no
// provider has been registered for that name so callers can surface the issue
// clearly rather than receiving a nil pointer.
func (f *Factory) Get(appName string) (interfaces.Provider, error) {
	p, ok := f.registry.Get(appName)
	if !ok {
		return nil, fmt.Errorf("provider factory: no provider registered for app %q", appName)
	}
	return p, nil
}

// Register adds a Provider to the underlying registry. It propagates
// registration errors (e.g. duplicate names) to the caller.
func (f *Factory) Register(name string, p interfaces.Provider) error {
	return f.registry.Register(name, p)
}

// RegisteredNames returns the names of all currently registered providers.
// Intended for startup logging and diagnostics only.
func (f *Factory) RegisteredNames() []string {
	return f.registry.Names()
}
