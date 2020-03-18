package oauth

type CodeGrantStore interface {
	GetCodeGrant(codeHash string) (*CodeGrant, error)
	CreateCodeGrant(*CodeGrant) error
	UpdateCodeGrant(*CodeGrant) error
	DeleteCodeGrant(*CodeGrant) error
}

type OfflineGrantStore interface {
	GetOfflineGrant(id string) (*OfflineGrant, error)
	CreateOfflineGrant(*OfflineGrant) error
	UpdateOfflineGrant(*OfflineGrant) error
	DeleteOfflineGrant(*OfflineGrant) error

	GetOfflineGrantByID(id string) (*OfflineGrant, error)
	ListOfflineGrants(userID string) ([]*OfflineGrant, error)
}

type AccessGrantStore interface {
	GetAccessGrant(tokenHash string) (*AccessGrant, error)
	CreateAccessGrant(*AccessGrant) error
	UpdateAccessGrant(*AccessGrant) error
	DeleteAccessGrant(*AccessGrant) error
}
