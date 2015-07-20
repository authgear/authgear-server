package provider

import ()

// Registry contains registered providers, which provide additional
// functionality implemented by plugins.
type Registry struct {
	authProviders map[string]AuthProvider
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		authProviders: map[string]AuthProvider{},
	}
}

// RegisterAuthProvider registers an AuthProvider with the registry.
func (r *Registry) RegisterAuthProvider(name string, p AuthProvider) {
	r.authProviders[name] = p
}

// GetAuthProvider gets an AuthProvider from the registry.
func (r *Registry) GetAuthProvider(name string) AuthProvider {
	provider := r.authProviders[name]
	return provider
}
