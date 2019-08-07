package principal

type Provider interface {
	ID() string
	ListPrincipalsByUserID(userID string) ([]Principal, error)
	GetPrincipalByID(principalID string) (Principal, error)
}
