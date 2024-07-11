package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/lib/session/idpsession"
)

//go:generate mockgen -source=store_grant.go -destination=store_grant_mock.go -package oauth

type CodeGrantStore interface {
	GetCodeGrant(codeHash string) (*CodeGrant, error)
	CreateCodeGrant(*CodeGrant) error
	DeleteCodeGrant(*CodeGrant) error
}

type SettingsActionGrantStore interface {
	GetSettingsActionGrant(codeHash string) (*SettingsActionGrant, error)
	CreateSettingsActionGrant(*SettingsActionGrant) error
	DeleteSettingsActionGrant(*SettingsActionGrant) error
}

type OfflineGrantStore interface {
	GetOfflineGrant(id string) (*OfflineGrant, error)
	CreateOfflineGrant(offlineGrant *OfflineGrant, expireAt time.Time) error
	DeleteOfflineGrant(*OfflineGrant) error

	AccessWithID(id string, accessEvent access.Event, expireAt time.Time) (*OfflineGrant, error)
	AccessOfflineGrantAndUpdateDeviceInfo(id string, accessEvent access.Event, deviceInfo map[string]interface{}, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantAuthenticatedAt(id string, authenticatedAt time.Time, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantApp2AppDeviceKey(id string, newKey string, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantDeviceSecretHash(
		grantID string,
		newDeviceSecretHash string,
		dpopJKT string,
		expireAt time.Time) (*OfflineGrant, error)
	RemoveOfflineGrantRefreshTokens(grantID string, tokenHashes []string, expireAt time.Time) (*OfflineGrant, error)
	AddOfflineGrantRefreshToken(
		grantID string,
		expireAt time.Time,
		tokenHash string,
		clientID string,
		scopes []string,
		authorizationID string,
		dpopJKT string,
	) (*OfflineGrant, error)

	ListOfflineGrants(userID string) ([]*OfflineGrant, error)
	ListClientOfflineGrants(clientID string, userID string) ([]*OfflineGrant, error)
}

type IDPSessionProvider interface {
	Get(id string) (*idpsession.IDPSession, error)
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

type PreAuthenticatedURLTokenStore interface {
	CreatePreAuthenticatedURLToken(*PreAuthenticatedURLToken) error
	ConsumePreAuthenticatedURLToken(tokenHash string) (*PreAuthenticatedURLToken, error)
}
