package customtoken

const providerName string = "custom_token"

type Provider interface {
	Decode(tokenString string) (SSOCustomTokenClaims, error)
	CreatePrincipal(principal Principal) error
	GetPrincipalByTokenPrincipalID(tokenPrincipalID string) (*Principal, error)
}
