package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	AppInitiatedSSOToWebTokenLifetime = duration.Short
)

type AppInitiatedSSOToWebToken struct {
	AppID           string   `json:"app_id"`
	AuthorizationID string   `json:"authorization_id"`
	ClientID        string   `json:"client_id"`
	OfflineGrantID  string   `json:"offline_grant_id"`
	Scopes          []string `json:"scopes"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	TokenHash string    `json:"token_hash"`
}

type AppInitiatedSSOToWebTokenAccessGrantService interface {
	IssueAccessGrant(
		client *config.OAuthClientConfig,
		scopes []string,
		authzID string,
		userID string,
		sessionID string,
		sessionKind GrantSessionKind,
		refreshTokenHash string,
	) (*IssueAccessGrantResult, error)
}

type IssueAppInitiatedSSOToWebTokenResult struct {
	Token     string
	TokenHash string
	TokenType string
	ExpiresIn int
}

type IssueAppInitiatedSSOToWebTokenOptions struct {
	AppID           string
	ClientID        string
	OfflineGrantID  string
	AuthorizationID string
	Scopes          []string
}

type AppInitiatedSSOToWebTokenOfflineGrantService interface {
	CreateNewRefreshToken(
		grant *OfflineGrant,
		clientID string,
		scopes []string,
		authorizationID string,
	) (*CreateNewRefreshTokenResult, *OfflineGrant, error)
}
