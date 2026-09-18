package oauth

import (
	"context"
	"strings"

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

type PrepareUserAccessGrantOptions struct {
	ClientConfig             *config.OAuthClientConfig
	Scopes                   []string
	AuthorizationID          string
	AuthenticationInfo       authenticationinfo.T
	SessionLike              SessionLike
	InitialRefreshTokenHash  string
	UserBlockingEventContext *UserBlockingEventContext
	// ResourceURI is empty when no resource was requested/bound. Threaded
	// into both the persisted AccessGrant and token issuance
	// (EncodeUserAccessTokenOptions).
	ResourceURI string
}

type IssueAccessGrantResult struct {
	Token     string
	TokenType string
	ExpiresIn int
	// Scopes is what the issued token actually carries. WriteTo always
	// reports it in the response, matching RFC 6749 §5.1 and the
	// client_credentials grant, which already does this.
	Scopes []string
}

func (r *IssueAccessGrantResult) WriteTo(resp protocol.TokenResponse) {
	if r != nil && resp != nil {
		resp.TokenType(r.TokenType)
		resp.AccessToken(r.Token)
		resp.ExpiresIn(r.ExpiresIn)
		resp.Scope(strings.Join(r.Scopes, " "))
	}
}

func (s *AccessGrantService) PrepareUserAccessGrant(
	ctx context.Context,
	options PrepareUserAccessGrantOptions,
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
		ResourceURI:             options.ResourceURI,
	}
	err := s.AccessGrants.CreateAccessGrant(ctx, accessGrant)
	if err != nil {
		return nil, err
	}

	clientLike := ClientClientLike(options.ClientConfig, options.Scopes)
	preparation, err := s.AccessTokenIssuer.PrepareUserAccessToken(ctx, EncodeUserAccessTokenOptions{
		OriginalToken:            token,
		ClientConfig:             options.ClientConfig,
		ClientLike:               clientLike,
		AccessGrant:              accessGrant,
		AuthenticationInfo:       options.AuthenticationInfo,
		UserBlockingEventContext: options.UserBlockingEventContext,
		ResourceURI:              options.ResourceURI,
	})
	if err != nil {
		return nil, err
	}

	return preparation, nil
}
