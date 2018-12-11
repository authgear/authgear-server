package oauth

const providerName string = "oauth"

type Provider interface {
	GetPrincipalByProviderUserID(providerName string, providerUserID string) (*Principal, error)
	GetPrincipalByUserID(providerName string, userID string) (*Principal, error)
	CreatePrincipal(principal Principal) error
	UpdatePrincipal(principal *Principal) error
	GetPrincipalsByUserID(userID string) ([]*Principal, error)
}
