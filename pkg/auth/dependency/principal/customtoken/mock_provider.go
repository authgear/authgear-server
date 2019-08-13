package customtoken

import (
	"errors"

	"github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

// MockProvider is the memory implementation of custom provider
type MockProvider struct {
	secret       string
	PrincipalMap map[string]Principal
}

// NewMockProvider creates a new instance of mock provider
func NewMockProvider(secret string) *MockProvider {
	return NewMockProviderWithPrincipalMap(secret, map[string]Principal{})
}

func NewMockProviderWithPrincipalMap(secret string, principalMap map[string]Principal) *MockProvider {
	return &MockProvider{
		secret:       secret,
		PrincipalMap: principalMap,
	}
}

func (p *MockProvider) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
	_, err = jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("fails to parse token")
			}
			return []byte(p.secret), nil
		},
	)

	return
}

func (p *MockProvider) CreatePrincipal(principal *Principal) error {
	if _, existed := p.PrincipalMap[principal.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	for _, p := range p.PrincipalMap {
		if p.TokenPrincipalID == principal.TokenPrincipalID {
			return skydb.ErrUserDuplicated
		}
	}

	p.PrincipalMap[principal.ID] = *principal
	return nil
}

func (p *MockProvider) UpdatePrincipal(principal *Principal) error {
	p.PrincipalMap[principal.ID] = *principal
	return nil
}

func (p *MockProvider) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	for _, p := range p.PrincipalMap {
		if p.TokenPrincipalID == tokenPrincipalID {
			return &p, nil
		}
	}

	return nil, skydb.ErrUserNotFound
}

func (p *MockProvider) ID() string {
	return providerName
}

func (p *MockProvider) ListPrincipalsByUserID(userID string) ([]principal.Principal, error) {
	var principals []principal.Principal
	for _, p := range p.PrincipalMap {
		if p.UserID == userID {
			principal := p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}

func (p *MockProvider) ListPrincipalsByClaim(claimName string, claimValue string) ([]principal.Principal, error) {
	var principals []principal.Principal
	for _, p := range p.PrincipalMap {
		if p.ClaimsValue[claimName] == claimValue {
			principal := p
			principals = append(principals, &principal)
		}
	}
	return principals, nil
}

func (p *MockProvider) GetPrincipalByID(principalID string) (principal.Principal, error) {
	for _, p := range p.PrincipalMap {
		if p.ID == principalID {
			principal := p
			return &principal, nil
		}
	}
	return nil, skydb.ErrUserNotFound
}

var (
	_ Provider = &MockProvider{}
)
