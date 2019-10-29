package principal

type mockIdentityProvider struct {
	providers []Provider
}

func NewMockIdentityProvider(providers ...Provider) IdentityProvider {
	return &mockIdentityProvider{providers}
}

func (p *mockIdentityProvider) ListPrincipalsByClaim(claimName string, claimValue string) ([]Principal, error) {
	principals := []Principal{}
	for _, provider := range p.providers {
		providerPrincipals, err := provider.ListPrincipalsByClaim(claimName, claimValue)
		if err != nil {
			return nil, err
		}
		principals = append(principals, providerPrincipals...)
	}
	return principals, nil
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
	return nil, ErrNotFound
}
