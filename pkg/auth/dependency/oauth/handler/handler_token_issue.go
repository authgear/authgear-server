package handler

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func (h *TokenHandler) IssueTokens(
	client config.OAuthClientConfig,
	attrs *authn.Attrs,
) (auth.AuthSession, protocol.TokenResponse, error) {
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID(),
		attrs.UserID,
		scopes,
	)
	if err != nil {
		return nil, nil, err
	}

	resp := protocol.TokenResponse{}

	offlineGrant, err := h.issueOfflineGrant(client, scopes, authz.ID, attrs, resp)
	if err != nil {
		return nil, nil, err
	}

	err = h.issueAccessGrant(client, scopes, authz.ID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, nil, err
	}

	return offlineGrant, resp, nil
}

func (h *TokenHandler) RefreshAPIToken(
	client config.OAuthClientConfig,
	refreshToken string,
) (accessToken string, err error) {
	resp, err := h.handleRefreshToken(client, protocol.TokenRequest{
		"client_id":     client.ClientID(),
		"refresh_token": refreshToken,
	})
	if err != nil {
		return "", err
	}
	return resp.GetAccessToken(), nil
}
