package password

const providerPassword string = "password"

type Provider interface {
	IsAuthDataValid(authData map[string]interface{}) bool
	CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]interface{}) error
	CreatePrincipal(principal Principal) error
	GetPrincipalByAuthData(authData map[string]interface{}, principal *Principal) error
	GetPrincipalByUserID(userID string) ([]*Principal, error)
	UpdatePrincipal(principal Principal) error
}
