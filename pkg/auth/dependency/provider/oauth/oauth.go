package oauth

const providerName string = "oauth"

type Provider interface {
	GetPrincipalByUserID(providerName string, userID string) (*Principal, error)
}
