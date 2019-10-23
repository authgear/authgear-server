package anonymous

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	Principals []Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// CreatePrincipal creates principal in PrincipalMap
func (m *MockProvider) CreatePrincipal(pp Principal) error {
	for _, p := range m.Principals {
		if p.ID == pp.ID {
			return principal.ErrAlreadyExists
		}
	}

	m.Principals = append(m.Principals, pp)
	return nil
}

func (m *MockProvider) ID() string {
	return providerAnonymous
}
