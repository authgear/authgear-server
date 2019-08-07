package principal

import (
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

type mockIdentityProvider struct {
	providers []Provider
}

func NewMockIdentityProvider(providers ...Provider) IdentityProvider {
	return &mockIdentityProvider{providers}
}

func (p *mockIdentityProvider) ListPrincipalsByUserID(userID string) ([]Principal, error) {
	principals := []Principal{}
	for _, provider := range p.providers {
		providerPrincipals, err := provider.ListPrincipalsByUserID(userID)
		if err != nil {
			return nil, err
		}
		principals = append(principals, providerPrincipals...)
	}
	return principals, nil
}

func (p *mockIdentityProvider) GetPrincipalByID(principalID string) (principal Principal, err error) {
	for _, provider := range p.providers {
		principal, err = provider.GetPrincipalByID(principalID)
		if err == nil {
			return
		}
	}
	return nil, skydb.ErrUserNotFound
}
