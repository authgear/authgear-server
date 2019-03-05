package oauth

import (
	"github.com/skygeario/skygear-server/pkg/core/skydb"
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

// NewMockProviderWithPrincipals creates a new instance of mock provider with principals
func NewMockProviderWithPrincipals(principals []*Principal) *MockProvider {
	provider := NewMockProvider(
		map[string]string{},
		map[string]Principal{},
	)
	for _, p := range principals {
		provider.CreatePrincipal(*p)
	}
	return provider
}

func NewMockProviderKey(providerName string, providerUserID string) string {
	return providerName + "." + providerUserID
}

func (m *MockProvider) GetPrincipalByProviderUserID(providerName string, providerUserID string) (*Principal, error) {
	key := NewMockProviderKey(providerName, providerUserID)
	if principal, ok := m.OAuthMap[key]; ok {
		return &principal, nil
	}

	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) GetPrincipalByUserID(providerName string, userID string) (*Principal, error) {
	if oauthKey, ok := m.PrincipalMap[userID]; ok {
		principal := m.OAuthMap[oauthKey]
		return &principal, nil
	}

	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) CreatePrincipal(principal Principal) error {
	key := NewMockProviderKey(principal.ProviderName, principal.ProviderUserID)
	m.OAuthMap[key] = principal
	m.PrincipalMap[principal.UserID] = key
	return nil
}

func (m *MockProvider) UpdatePrincipal(principal *Principal) error {
	key := NewMockProviderKey(principal.ProviderName, principal.ProviderUserID)
	m.OAuthMap[key] = *principal
	return nil
}

func (m *MockProvider) DeletePrincipal(principal *Principal) error {
	key := NewMockProviderKey(principal.ProviderName, principal.ProviderUserID)
	delete(m.OAuthMap, key)
	delete(m.PrincipalMap, principal.UserID)
	return nil
}

func (m *MockProvider) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	var principals []*Principal
	for _, p := range m.OAuthMap {
		if p.UserID == userID {
			var principal Principal
			principal = p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}
