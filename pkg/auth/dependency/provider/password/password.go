package password

const providerPassword string = "password"

type Provider interface {
	IsAuthDataValid(authData map[string]string) bool
	IsAuthDataMatching(authData map[string]string) bool
	CreatePrincipalsByAuthData(authInfoID string, password string, authData map[string]string) error
	CreatePrincipal(principal Principal) error
	GetPrincipalByAuthData(authData map[string]string, principal *Principal) (err error)
	GetPrincipalsByUserID(userID string) ([]*Principal, error)
	GetPrincipalsByEmail(email string) ([]*Principal, error)
	UpdatePrincipal(principal Principal) error
}
