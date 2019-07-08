package anonymous

import "github.com/skygeario/skygear-server/pkg/core/skydb"

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
func (m *MockProvider) CreatePrincipal(principal Principal) error {
	for _, p := range m.Principals {
		if p.ID == principal.ID {
			return skydb.ErrUserDuplicated
		}
	}

	m.Principals = append(m.Principals, principal)
	return nil
}
