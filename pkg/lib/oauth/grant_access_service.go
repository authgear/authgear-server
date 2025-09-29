package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AccessGrantService struct {
	AppID config.AppID

	AccessGrants      AccessGrantStore
	AccessTokenIssuer AccessTokenEncoding
	Clock             clock.Clock
}

type IssueAccessGrantOptions struct {
	ClientConfig            *config.OAuthClientConfig
	Scopes                  []string
	AuthorizationID         string
	AuthenticationInfo      authenticationinfo.T
	SessionLike             SessionLike
	InitialRefreshTokenHash string
}

type IssueAccessGrantResult struct {
	Token     string
	TokenType string
	ExpiresIn int
}

func (r *IssueAccessGrantResult) WriteTo(resp protocol.TokenResponse) {
	if r != nil && resp != nil {
		resp.TokenType(r.TokenType)
		resp.AccessToken(r.Token)
		resp.ExpiresIn(r.ExpiresIn)
	}
}

func (s *AccessGrantService) IssueAccessGrant(
	ctx context.Context,
	options IssueAccessGrantOptions,
) (PrepareUserAccessTokenResult, error) {
	token := GenerateToken()
	now := s.Clock.NowUTC()

	accessGrant := &AccessGrant{
		AppID:                   string(s.AppID),
		AuthorizationID:         options.AuthorizationID,
		SessionID:               options.SessionLike.SessionID(),
		SessionKind:             GrantSessionKindFromSessionType(options.SessionLike.SessionType()),
		CreatedAt:               now,
		ExpireAt:                now.Add(options.ClientConfig.AccessTokenLifetime.Duration()),
		Scopes:                  options.Scopes,
		TokenHash:               HashToken(token),
		InitialRefreshTokenHash: options.InitialRefreshTokenHash,
	}
	err := s.AccessGrants.CreateAccessGrant(ctx, accessGrant)
	if err != nil {
		return nil, err
	}

	clientLike := ClientClientLike(options.ClientConfig, options.Scopes)
	preparation, err := s.AccessTokenIssuer.PrepareUserAccessToken(ctx, EncodeUserAccessTokenOptions{
		OriginalToken:      token,
		ClientConfig:       options.ClientConfig,
		ClientLike:         clientLike,
		AccessGrant:        accessGrant,
		AuthenticationInfo: options.AuthenticationInfo,
	})
	if err != nil {
		return nil, err
	}

	return preparation, nil
}
