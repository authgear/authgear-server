package facade

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/handler"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
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
	) (*oauth.OfflineGrant, error)
	IssueAccessGrant(
		client *config.OAuthClientConfig,
		scopes []string,
		authzID string,
		userID string,
		sessionID string,
		sessionKind oauth.GrantSessionKind,
		resp protocol.TokenResponse,
	) error
}

type OAuthFacade struct {
	Config         *config.OAuthConfig
	Authorizations OAuthAuthorizationService
	Tokens         OAuthTokenService
	Clock          clock.Clock
}

func (f *OAuthFacade) CreateSession(clientID string, userID string) (protocol.TokenResponse, error) {
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

	client, ok := f.Config.GetClient(clientID)
	if !ok {
		return nil, apierrors.NewInvalid("invalid client ID")
	}
	if client.ClientParty() != config.ClientPartyFirst {
		return nil, apierrors.NewForbidden("cannot create session for non-first party client")
	}

	authz, err := f.Authorizations.CheckAndGrant(
		clientID,
		userID,
		scopes,
	)
	if err != nil {
		return nil, err
	}

	offlineGrantOpts := handler.IssueOfflineGrantOptions{
		Scopes:             scopes,
		AuthorizationID:    authz.ID,
		AuthenticationInfo: authenticationInfo,
		DeviceInfo:         deviceInfo,
	}

	resp := protocol.TokenResponse{}
	offlineGrant, err := f.Tokens.IssueOfflineGrant(client, offlineGrantOpts, resp)
	if err != nil {
		return nil, err
	}

	err = f.Tokens.IssueAccessGrant(
		client,
		scopes,
		authz.ID,
		authz.UserID,
		offlineGrant.ID,
		oauth.GrantSessionKindOffline,
		resp,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
