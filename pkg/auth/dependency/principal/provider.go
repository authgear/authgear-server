package principal

type Provider interface {
	ID() string
	ListPrincipalsByClaim(claimName string, claimValue string) ([]Principal, error)
	ListPrincipalsByUserID(userID string) ([]Principal, error)
	GetPrincipalByID(principalID string) (Principal, error)
}
