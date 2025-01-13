package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AccessGrantService struct {
	AppID config.AppID

	AccessGrants      AccessGrantStore
	AccessTokenIssuer AccessTokenEncoding
	Clock             clock.Clock
}

type IssueAccessGrantOptions struct {
	ClientConfig       *config.OAuthClientConfig
	Scopes             []string
	AuthorizationID    string
	AuthenticationInfo authenticationinfo.T
	SessionLike        SessionLike
	RefreshTokenHash   string
}

type IssueAccessGrantResult struct {
	Token     string
	TokenType string
	ExpiresIn int
}

func (s *AccessGrantService) IssueAccessGrant(
	ctx context.Context,
	options IssueAccessGrantOptions,
) (*IssueAccessGrantResult, error) {
	token := GenerateToken()
	now := s.Clock.NowUTC()

	accessGrant := &AccessGrant{
		AppID:            string(s.AppID),
		AuthorizationID:  options.AuthorizationID,
		SessionID:        options.SessionLike.SessionID(),
		SessionKind:      GrantSessionKindFromSessionType(options.SessionLike.SessionType()),
		CreatedAt:        now,
		ExpireAt:         now.Add(options.ClientConfig.AccessTokenLifetime.Duration()),
		Scopes:           options.Scopes,
		TokenHash:        HashToken(token),
		RefreshTokenHash: options.RefreshTokenHash,
	}
	err := s.AccessGrants.CreateAccessGrant(ctx, accessGrant)
	if err != nil {
		return nil, err
	}

	clientLike := ClientClientLike(options.ClientConfig, options.Scopes)
	at, err := s.AccessTokenIssuer.EncodeAccessToken(ctx, EncodeAccessTokenOptions{
		OriginalToken:      token,
		ClientConfig:       options.ClientConfig,
		ClientLike:         clientLike,
		AccessGrant:        accessGrant,
		AuthenticationInfo: options.AuthenticationInfo,
	})
	if err != nil {
		return nil, err
	}

	result := &IssueAccessGrantResult{
		Token:     at,
		TokenType: "Bearer",
		ExpiresIn: int(options.ClientConfig.AccessTokenLifetime),
	}

	return result, nil
}
