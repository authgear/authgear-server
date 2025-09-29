package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type OAuthAuthorizationService interface {
	CheckAndGrant(
		ctx context.Context,
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
}

type OAuthTokenService interface {
	IssueOfflineGrant(
		ctx context.Context,
		client *config.OAuthClientConfig,
		opts handler.IssueOfflineGrantOptions,
		resp protocol.TokenResponse,
	) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error)
	IssueAccessGrantByRefreshToken(
		ctx context.Context,
		options handler.IssueAccessGrantByRefreshTokenOptions,
	) (*handler.IssueAccessGrantByRefreshTokenResult, error)
}

type OAuthAccessTokenEncoding interface {
	MakeUserAccessTokenFromPreparationResult(
		ctx context.Context,
		options oauth.MakeUserAccessTokenFromPreparationOptions,
	) (*oauth.IssueAccessGrantResult, error)
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type OAuthFacade struct {
	Config              *config.OAuthConfig
	Users               UserService
	Authorizations      OAuthAuthorizationService
	Tokens              OAuthTokenService
	AccessTokenCoding   OAuthAccessTokenEncoding
	Clock               clock.Clock
	OAuthClientResolver OAuthClientResolver
}

func (f *OAuthFacade) CreateSession(ctx context.Context, clientID string, userID string, deviceInfo map[string]interface{}) (session.ListableSession, protocol.TokenResponse, error) {
	scopes := []string{
		"openid",
		oauth.OfflineAccess,
		oauth.FullAccessScope,
	}
	authenticationInfo := authenticationinfo.T{
		UserID:          userID,
		AuthenticatedAt: f.Clock.NowUTC(),
	}

	client := f.OAuthClientResolver.ResolveClient(clientID)
	if client == nil {
		return nil, nil, apierrors.NewInvalid("invalid client ID")
	}
	if !client.IsFirstParty() {
		return nil, nil, apierrors.NewForbidden("cannot create session for non-first party client")
	}

	// Check user existence.
	_, err := f.Users.GetRaw(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	authz, err := f.Authorizations.CheckAndGrant(
		ctx,
		clientID,
		userID,
		scopes,
	)
	if err != nil {
		return nil, nil, err
	}

	offlineGrantOpts := handler.IssueOfflineGrantOptions{
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: authenticationInfo,
		DeviceInfo:         deviceInfo,
		// dpop not supported for offline grants created by this method
		DPoPJKT: "",
	}

	resp := protocol.TokenResponse{}
	offlineGrant, newTokenHash, err := f.Tokens.IssueOfflineGrant(ctx, client, offlineGrantOpts, resp)
	if err != nil {
		return nil, nil, err
	}

	result1, err := f.Tokens.IssueAccessGrantByRefreshToken(
		ctx,
		handler.IssueAccessGrantByRefreshTokenOptions{
			IssueAccessGrantOptions: oauth.IssueAccessGrantOptions{
				ClientConfig:            client,
				Scopes:                  scopes,
				AuthorizationID:         authz.ID,
				AuthenticationInfo:      offlineGrant.GetAuthenticationInfo(),
				SessionLike:             offlineGrant,
				InitialRefreshTokenHash: newTokenHash,
			},
			ShouldRotateRefreshToken: false, // The token is new, no need to rotate
		},
	)
	if err != nil {
		return nil, nil, err
	}

	result2, err := f.AccessTokenCoding.MakeUserAccessTokenFromPreparationResult(ctx, oauth.MakeUserAccessTokenFromPreparationOptions{
		PreparationResult: result1.PreparationResult,
		ClientConfig:      client,
	})
	if err != nil {
		return nil, nil, err
	}

	result1.RotateRefreshTokenResult.WriteTo(resp)
	result2.WriteTo(resp)
	return offlineGrant, resp, nil
}
