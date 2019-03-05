package anonymous

import (
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

// MockProvider is the memory implementation of password provider
type MockProvider struct {
	Provider
	PrincipalMap map[string]Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider() *MockProvider {
	return NewMockProviderWithPrincipalMap(map[string]Principal{})
}

// NewMockProviderWithPrincipalMap creates a new instance of mock provider with PrincipalMap
func NewMockProviderWithPrincipalMap(principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		PrincipalMap: principalMap,
	}
}

// CreatePrincipal creates principal in PrincipalMap
func (m *MockProvider) CreatePrincipal(principal Principal) error {
	if _, existed := m.PrincipalMap[principal.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	m.PrincipalMap[principal.ID] = principal
	return nil
}
