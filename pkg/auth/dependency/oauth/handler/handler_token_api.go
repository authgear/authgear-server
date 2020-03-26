package handler

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

func (h *TokenHandler) IssueAuthAPITokens(
	client config.OAuthClientConfiguration,
	attrs *authn.Attrs,
) (session auth.AuthSession, accessToken string, refreshToken string, err error) {
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Time.NowUTC(),
		h.AppID,
		client.ClientID(),
		attrs.UserID,
		scopes,
	)
	if err != nil {
		return nil, "", "", err
	}

	resp := protocol.TokenResponse{}

	offlineGrant, err := h.issueOfflineGrant(client, scopes, authz.ID, attrs, resp)
	if err != nil {
		return nil, "", "", err
	}

	err = h.issueAccessGrant(client, scopes, authz.ID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, "", "", err
	}

	return offlineGrant, resp.GetAccessToken(), resp.GetRefreshToken(), nil
}
