package storage

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
)

// ErrNoBackend is returned when Resolve cannot find a registered backend for the URI scheme.
var ErrNoBackend = errors.New("storage: no backend registered")

// SchemeError wraps ErrNoBackend with the offending scheme.
type SchemeError struct {
	Scheme string
}

func (e *SchemeError) Error() string {
	return ErrNoBackend.Error() + " for scheme " + e.Scheme
}

func (e *SchemeError) Unwrap() error {
	return ErrNoBackend
}

// defaultReg is the process-wide registry, populated by backend init() functions.
// After init completes the map is never mutated — only store internals are
// reconfigured via Configure. Resolve is lock-free.
var defaultReg atomic.Pointer[Registry]

func init() {
	defaultReg.Store(&Registry{backends: make(map[string]Store)})
}

// Register adds a store to the default registry under scheme.
// It must be called only from package init functions (single-threaded).
func Register(scheme string, store Store) {
	defaultReg.Load().backends[scheme] = store
}

// Unregister removes the backend associated with scheme.
// Returns true if a backend was removed.
func Unregister(scheme string) bool {
	_, exists := defaultReg.Load().backends[scheme]
	if exists {
		delete(defaultReg.Load().backends, scheme)
	}
	return exists
}

// Backend returns the store registered under scheme, or false if none.
// Use this to retrieve a store for configuration.
func Backend(scheme string) (Store, bool) {
	s, ok := defaultReg.Load().backends[scheme]
	return s, ok
}

// Default returns the current default registry.
func Default() *Registry {
	return defaultReg.Load()
}

// Resolve returns the Store for the URI scheme of location from the default registry.
func Resolve(location string) (Store, error) {
	return defaultReg.Load().Resolve(location)
}

// Registry routes storage operations to the correct backend based on the URI scheme.
type Registry struct {
	backends map[string]Store
}

// Resolve returns the Store whose key matches the URI scheme of location.
// A bare path with no scheme is treated as "file".
func (r *Registry) Resolve(location string) (Store, error) {
	scheme := extractScheme(location)
	s, ok := r.backends[scheme]
	if !ok {
		return nil, fmt.Errorf("storage: no backend registered for scheme %q", scheme)
	}
	return s, nil
}

// extractScheme returns the scheme portion before "://", or "file" if absent.
// This avoids url.Parse overhead on the hot path.
func extractScheme(location string) string {
	if idx := strings.Index(location, "://"); idx > 0 {
		return location[:idx]
	}
	return "file"
}
