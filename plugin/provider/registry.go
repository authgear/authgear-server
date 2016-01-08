package provider

import (
	"sync"
)

// Registry contains registered providers, which provide additional
// functionality implemented by plugins.
type Registry struct {
	mutex         sync.RWMutex
	authProviders map[string]AuthProvider
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		mutex:         sync.RWMutex{},
		authProviders: map[string]AuthProvider{},
	}
}

// RegisterAuthProvider registers an AuthProvider with the registry.
func (r *Registry) RegisterAuthProvider(name string, p AuthProvider) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.authProviders[name] = p
}

// GetAuthProvider gets an AuthProvider from the registry.
func (r *Registry) GetAuthProvider(name string) AuthProvider {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	provider := r.authProviders[name]
	return provider
}
