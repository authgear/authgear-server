package password

const providerPassword string = "password"

type Provider interface {
	CreatePrincipal(principal Principal) error
	GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error
	GetPrincipalByUserID(userID string, principal *Principal) error
	UpdatePrincipal(principal Principal) error
}
