package oauth

import (
	"time"
)

type CodeGrantStore interface {
	GetCodeGrant(codeHash string) (*CodeGrant, error)
	CreateCodeGrant(*CodeGrant) error
	DeleteCodeGrant(*CodeGrant) error
}

type OfflineGrantStore interface {
	GetOfflineGrant(id string) (*OfflineGrant, error)
	CreateOfflineGrant(offlineGrant *OfflineGrant, expireAt time.Time) error
	UpdateOfflineGrant(offlineGrant *OfflineGrant, expireAt time.Time) error
	DeleteOfflineGrant(*OfflineGrant) error

	ListOfflineGrants(userID string) ([]*OfflineGrant, error)
}

type AccessGrantStore interface {
	GetAccessGrant(tokenHash string) (*AccessGrant, error)
	CreateAccessGrant(*AccessGrant) error
	DeleteAccessGrant(*AccessGrant) error
}

type AppSessionStore interface {
	GetAppSession(tokenHash string) (*AppSession, error)
	CreateAppSession(*AppSession) error
	DeleteAppSession(*AppSession) error
}

type AppSessionTokenStore interface {
	GetAppSessionToken(tokenHash string) (*AppSessionToken, error)
	CreateAppSessionToken(*AppSessionToken) error
	DeleteAppSessionToken(*AppSessionToken) error
}
