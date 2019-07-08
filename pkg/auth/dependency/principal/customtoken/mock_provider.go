package customtoken

import (
	"errors"

	"github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/core/skydb"
)

// MockProvider is the memory implementation of custom provider
type MockProvider struct {
	Provider
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

func (p MockProvider) Decode(tokenString string) (claims SSOCustomTokenClaims, err error) {
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

func (p MockProvider) CreatePrincipal(principal Principal) error {
	if _, existed := p.PrincipalMap[principal.ID]; existed {
		return skydb.ErrUserDuplicated
	}

	for _, p := range p.PrincipalMap {
		if p.TokenPrincipalID == principal.TokenPrincipalID {
			return skydb.ErrUserDuplicated
		}
	}

	p.PrincipalMap[principal.ID] = principal
	return nil
}

func (p MockProvider) GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error) {
	for _, p := range p.PrincipalMap {
		if p.TokenPrincipalID == tokenPrincipalID {
			return &p, nil
		}
	}

	return nil, skydb.ErrUserNotFound
}

func (p MockProvider) ID() string {
	return providerName
}

func (p MockProvider) DeriveClaims(_ principal.Principal) principal.Claims {
	// TODO(sso): return custom token email
	return principal.Claims{}
}
