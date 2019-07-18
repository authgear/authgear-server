package oauth

import (
	"reflect"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type MockProvider struct {
	Provider
	Principals []*Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(principals []*Principal) *MockProvider {
	return &MockProvider{
		Principals: principals,
	}
}

func (m *MockProvider) GetPrincipalByProvider(options GetByProviderOptions) (*Principal, error) {
	if options.ProviderKeys == nil {
		options.ProviderKeys = map[string]interface{}{}
	}
	for _, p := range m.Principals {
		if p.ProviderType == options.ProviderType &&
			reflect.DeepEqual(p.ProviderKeys, options.ProviderKeys) &&
			p.ProviderUserID == options.ProviderUserID {
			return p, nil
		}
	}
	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) GetPrincipalByUser(options GetByUserOptions) (*Principal, error) {
	if options.ProviderKeys == nil {
		options.ProviderKeys = map[string]interface{}{}
	}
	for _, p := range m.Principals {
		if p.ProviderType == options.ProviderType &&
			reflect.DeepEqual(p.ProviderKeys, options.ProviderKeys) &&
			p.UserID == options.UserID {
			return p, nil
		}
	}
	return nil, skydb.ErrUserNotFound
}

func (m *MockProvider) CreatePrincipal(principal *Principal) error {
	m.Principals = append(m.Principals, principal)
	return nil
}

func (m *MockProvider) UpdatePrincipal(principal *Principal) error {
	for i, p := range m.Principals {
		if p.ID == principal.ID {
			m.Principals[i] = principal
		}
	}
	return nil
}

func (m *MockProvider) DeletePrincipal(principal *Principal) error {
	j := -1
	for i, p := range m.Principals {
		if p.ID == principal.ID {
			j = i
			break
		}
	}
	if j != -1 {
		m.Principals = append(m.Principals[:j], m.Principals[j+1:]...)
	}
	return nil
}

func (m *MockProvider) GetPrincipalsByUserID(userID string) ([]*Principal, error) {
	var principals []*Principal
	for _, p := range m.Principals {
		if p.UserID == userID {
			var principal Principal
			principal = *p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}

func (m *MockProvider) ID() string {
	return providerName
}

func (m *MockProvider) DeriveClaims(pp principal.Principal) (claims principal.Claims) {
	claims = principal.Claims{}
	attrs := pp.Attributes()
	providerType, ok := attrs["provider_type"].(string)
	if !ok {
		return
	}
	rawProfile, ok := attrs["raw_profile"].(map[string]interface{})
	if !ok {
		return
	}
	decoder := sso.GetUserInfoDecoder(config.OAuthProviderType(providerType))
	providerUserInfo := decoder.DecodeUserInfo(rawProfile)
	if providerUserInfo.Email != "" {
		claims["email"] = providerUserInfo.Email
	}
	return
}
