package password

const providerPassword string = "password"

type Provider interface {
	IsAuthDataValid(authData map[string]interface{}) bool
	CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]interface{}) error
	CreatePrincipal(principal Principal) error
	GetPrincipalsByAuthData(authData map[string]interface{}) ([]*Principal, error)
	GetPrincipalsByUserID(userID string) ([]*Principal, error)
	GetPrincipalsByEmail(email string) ([]*Principal, error)
	UpdatePrincipal(principal Principal) error
}
