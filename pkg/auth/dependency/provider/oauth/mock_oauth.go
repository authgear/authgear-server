package oauth

import "github.com/skygeario/skygear-server/pkg/server/skydb"

type MockProvider struct {
	Provider
	PrincipalMap map[string]Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		PrincipalMap: principalMap,
	}
}

func (m *MockProvider) genKey(providerName string, userID string) string {
	return providerName + "." + userID
}

func (m *MockProvider) GetPrincipalByUserID(providerName string, userID string) (*Principal, error) {
	key := m.genKey(providerName, userID)
	if principal, ok := m.PrincipalMap[key]; ok {
		return &principal, nil
	}

	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) CreatePrincipal(principal Principal) error {
	key := m.genKey(principal.ProviderName, principal.UserID)
	m.PrincipalMap[key] = principal
	return nil
}

func (m *MockProvider) UpdatePrincipal(principal *Principal) error {
	key := m.genKey(principal.ProviderName, principal.UserID)
	m.PrincipalMap[key] = *principal
	return nil
}
