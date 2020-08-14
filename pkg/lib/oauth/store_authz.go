package oauth

type AuthorizationStore interface {
	Get(userID, clientID string) (*Authorization, error)
	GetByID(id string) (*Authorization, error)
	Create(*Authorization) error
	Delete(*Authorization) error
	UpdateScopes(*Authorization) error
}
