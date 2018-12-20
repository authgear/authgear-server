package oauth

import (
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type MockProvider struct {
	Provider
	PrincipalMap map[string]string
	OAuthMap     map[string]Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(principalMap map[string]string, oauthMap map[string]Principal) *MockProvider {
	return &MockProvider{
		PrincipalMap: principalMap,
		OAuthMap:     oauthMap,
	}
}

func (m *MockProvider) genKey(providerName string, providerUserID string) string {
	return providerName + "." + providerUserID
}

func (m *MockProvider) GetPrincipalByProviderUserID(providerName string, providerUserID string) (*Principal, error) {
	key := m.genKey(providerName, providerUserID)
	if principal, ok := m.OAuthMap[key]; ok {
		return &principal, nil
	}

	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) GetPrincipalByUserID(userID string) (*Principal, error) {
	if oauthKey, ok := m.PrincipalMap[userID]; ok {
		principal := m.OAuthMap[oauthKey]
		return &principal, nil
	}

	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) CreatePrincipal(principal Principal) error {
	key := m.genKey(principal.ProviderName, principal.ProviderUserID)
	m.OAuthMap[key] = principal
	m.PrincipalMap[principal.UserID] = key
	return nil
}

func (m *MockProvider) UpdatePrincipal(principal *Principal) error {
	key := m.genKey(principal.ProviderName, principal.ProviderUserID)
	m.OAuthMap[key] = *principal
	return nil
}
