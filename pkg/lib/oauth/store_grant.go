package oauth

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/session/access"
)

//go:generate go tool mockgen -source=store_grant.go -destination=store_grant_mock_test.go -package oauth

type CodeGrantStore interface {
	GetCodeGrant(ctx context.Context, codeHash string) (*CodeGrant, error)
	CreateCodeGrant(ctx context.Context, g *CodeGrant) error
	DeleteCodeGrant(ctx context.Context, g *CodeGrant) error
}

type SettingsActionGrantStore interface {
	GetSettingsActionGrant(ctx context.Context, codeHash string) (*SettingsActionGrant, error)
	CreateSettingsActionGrant(ctx context.Context, g *SettingsActionGrant) error
	DeleteSettingsActionGrant(ctx context.Context, g *SettingsActionGrant) error
}

type AddOfflineGrantRefreshTokenOptions struct {
	OfflineGrantID                 string
	AccessInfo                     access.Info
	OfflineGrantExpireAt           time.Time
	ShortLivedRefreshTokenExpireAt *time.Time
	TokenHash                      string
	ClientID                       string
	Scopes                         []string
	AuthorizationID                string
	DPoPJKT                        string
}

type RotateOfflineGrantRefreshTokenOptions struct {
	OfflineGrantID          string
	InitialRefreshTokenHash string
	NewRefreshTokenHash     string
}

type OfflineGrantStore interface {
	GetOfflineGrantWithoutExpireAt(ctx context.Context, id string) (*OfflineGrant, error)
	CreateOfflineGrant(ctx context.Context, offlineGrant *OfflineGrant) error
	DeleteOfflineGrant(ctx context.Context, g *OfflineGrant) error

	UpdateOfflineGrantWithMutator(ctx context.Context, grantID string, expireAt time.Time, mutator func(*OfflineGrant) *OfflineGrant) (*OfflineGrant, error)
	UpdateOfflineGrantDeviceInfo(ctx context.Context, id string, deviceInfo map[string]interface{}, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantAuthenticatedAt(ctx context.Context, id string, authenticatedAt time.Time, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantApp2AppDeviceKey(ctx context.Context, id string, newKey string, expireAt time.Time) (*OfflineGrant, error)
	UpdateOfflineGrantDeviceSecretHash(
		ctx context.Context,
		grantID string,
		newDeviceSecretHash string,
		dpopJKT string,
		expireAt time.Time) (*OfflineGrant, error)
	RemoveOfflineGrantRefreshTokens(ctx context.Context, grantID string, tokenHashes []string, expireAt time.Time) (*OfflineGrant, error)
	AddOfflineGrantRefreshToken(
		ctx context.Context,
		options AddOfflineGrantRefreshTokenOptions,
	) (*OfflineGrant, error)
	AddOfflineGrantSAMLServiceProviderParticipant(
		ctx context.Context,
		grantID string,
		newServiceProviderID string,
		expireAt time.Time,
	) (*OfflineGrant, error)
	RotateOfflineGrantRefreshToken(
		ctx context.Context,
		opts RotateOfflineGrantRefreshTokenOptions,
		expireAt time.Time,
	) (*OfflineGrant, error)

	ListOfflineGrants(ctx context.Context, userID string) ([]*OfflineGrant, error)
	ListClientOfflineGrants(ctx context.Context, clientID string, userID string) ([]*OfflineGrant, error)

	CleanUpForDeletingUserID(ctx context.Context, userID string) error
}

type AccessGrantStore interface {
	GetAccessGrant(ctx context.Context, tokenHash string) (*AccessGrant, error)
	CreateAccessGrant(ctx context.Context, g *AccessGrant) error
	DeleteAccessGrant(ctx context.Context, g *AccessGrant) error
}

type AppSessionStore interface {
	GetAppSession(ctx context.Context, tokenHash string) (*AppSession, error)
	CreateAppSession(ctx context.Context, s *AppSession) error
	DeleteAppSession(ctx context.Context, s *AppSession) error
}

type AppSessionTokenStore interface {
	GetAppSessionToken(ctx context.Context, tokenHash string) (*AppSessionToken, error)
	CreateAppSessionToken(ctx context.Context, t *AppSessionToken) error
	DeleteAppSessionToken(ctx context.Context, t *AppSessionToken) error
}

type PreAuthenticatedURLTokenStore interface {
	CreatePreAuthenticatedURLToken(ctx context.Context, t *PreAuthenticatedURLToken) error
	ConsumePreAuthenticatedURLToken(ctx context.Context, tokenHash string) (*PreAuthenticatedURLToken, error)
}
