// Package interfaces defines the core contracts (interfaces) for the
// Notification Worker. Interfaces live here — not in the packages that
// implement them — so that consumers own the contracts and implementations
// remain free to evolve independently.
package interfaces

// Provider is the contract that every app-specific notification provider must
// satisfy. A provider encapsulates the connection and authentication logic
// required to integrate with one external application (e.g. Stackday, Expense
// Tracker, CRM).
//
// Implementations are registered with the provider factory and resolved at
// runtime by application name.
type Provider interface {
	// Name returns the unique, human-readable identifier for this provider.
	// It must match the name registered in the provider factory
	// (e.g. "stackday", "expense").
	Name() string

	// Initialize performs any one-time setup required before the provider can
	// be used (e.g. opening a database connection, loading credentials).
	// It is called once during application startup.
	Initialize() error
}