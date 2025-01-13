package oauth

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/util/duration"
)

const (
	PreAuthenticatedURLTokenLifetime = duration.Short
)

type PreAuthenticatedURLToken struct {
	AppID           string   `json:"app_id"`
	AuthorizationID string   `json:"authorization_id"`
	ClientID        string   `json:"client_id"`
	OfflineGrantID  string   `json:"offline_grant_id"`
	Scopes          []string `json:"scopes"`

	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	TokenHash string    `json:"token_hash"`
}

type PreAuthenticatedURLTokenAccessGrantService interface {
	IssueAccessGrant(
		ctx context.Context,
		options IssueAccessGrantOptions,
	) (*IssueAccessGrantResult, error)
}

type PreAuthenticatedURLTokenOfflineGrantService interface {
	GetOfflineGrant(ctx context.Context, id string) (*OfflineGrant, error)
	CreateNewRefreshToken(
		ctx context.Context,
		grant *OfflineGrant,
		clientID string,
		scopes []string,
		authorizationID string,
		dpopJKT string,
	) (*CreateNewRefreshTokenResult, *OfflineGrant, error)
}
