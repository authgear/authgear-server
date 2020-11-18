package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func (h *TokenHandler) IssueTokens(
	client *config.OAuthClientConfig,
	attrs *session.Attrs,
) (session.Session, protocol.TokenResponse, error) {
	scopes := []string{"openid", oauth.FullAccessScope}

	authz, err := checkAuthorization(
		h.Authorizations,
		h.Clock.NowUTC(),
		h.AppID,
		client.ClientID,
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
