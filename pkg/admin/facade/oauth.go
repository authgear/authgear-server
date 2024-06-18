package facade

import (
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
		clientID string,
		userID string,
		scopes []string,
	) (*oauth.Authorization, error)
}

type OAuthTokenService interface {
	IssueOfflineGrant(
		client *config.OAuthClientConfig,
		opts handler.IssueOfflineGrantOptions,
		resp protocol.TokenResponse,
	) (offlineGrant *oauth.OfflineGrant, tokenHash string, err error)
	IssueAccessGrant(
		client *config.OAuthClientConfig,
		scopes []string,
		authzID string,
		userID string,
		sessionID string,
		sessionKind oauth.GrantSessionKind,
		refreshTokenHash string,
		resp protocol.TokenResponse,
	) error
}

type OAuthClientResolver interface {
	ResolveClient(clientID string) *config.OAuthClientConfig
}

type OAuthFacade struct {
	Config              *config.OAuthConfig
	Users               UserService
	Authorizations      OAuthAuthorizationService
	Tokens              OAuthTokenService
	Clock               clock.Clock
	OAuthClientResolver OAuthClientResolver
}

func (f *OAuthFacade) CreateSession(clientID string, userID string) (session.ListableSession, protocol.TokenResponse, error) {
	scopes := []string{
		"openid",
		"offline_access",
		oauth.FullAccessScope,
	}
	authenticationInfo := authenticationinfo.T{
		UserID:          userID,
		AuthenticatedAt: f.Clock.NowUTC(),
	}
	deviceInfo := make(map[string]interface{})

	client := f.OAuthClientResolver.ResolveClient(clientID)
	if client == nil {
		return nil, nil, apierrors.NewInvalid("invalid client ID")
	}
	if !client.IsFirstParty() {
		return nil, nil, apierrors.NewForbidden("cannot create session for non-first party client")
	}

	// Check user existence.
	_, err := f.Users.GetRaw(userID)
	if err != nil {
		return nil, nil, err
	}

	authz, err := f.Authorizations.CheckAndGrant(
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
	}

	resp := protocol.TokenResponse{}
	offlineGrant, tokenHash, err := f.Tokens.IssueOfflineGrant(client, offlineGrantOpts, resp)
	if err != nil {
		return nil, nil, err
	}

	err = f.Tokens.IssueAccessGrant(
		client,
		scopes,
		authz.ID,
		authz.UserID,
		offlineGrant.ID,
		oauth.GrantSessionKindOffline,
		tokenHash,
		resp,
	)
	if err != nil {
		return nil, nil, err
	}

	return offlineGrant, resp, nil
}
