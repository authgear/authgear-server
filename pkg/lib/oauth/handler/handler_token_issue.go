package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const AppSessionTokenDuration = duration.Short

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

	err = h.issueAccessGrant(client, scopes, authz.ID, authz.UserID,
		offlineGrant.ID, oauth.GrantSessionKindOffline, resp)
	if err != nil {
		return nil, nil, err
	}

	return offlineGrant, resp, nil
}

func (h *TokenHandler) IssueAppSessionToken(refreshToken string) (string, *oauth.AppSessionToken, error) {
	authz, grant, err := h.parseRefreshToken(refreshToken)
	if err != nil {
		return "", nil, err
	}

	// Ensure client is authorized with full user access (i.e. first-party client)
	if !authz.IsAuthorized([]string{oauth.FullAccessScope}) {
		return "", nil, protocol.NewError("access_denied", "the client is not authorized to have full user access")
	}

	now := h.Clock.NowUTC()
	token := oauth.GenerateToken()
	sToken := &oauth.AppSessionToken{
		AppID:          grant.AppID,
		OfflineGrantID: grant.ID,
		CreatedAt:      now,
		ExpireAt:       now.Add(AppSessionTokenDuration),
		TokenHash:      oauth.HashToken(token),
	}

	err = h.AppSessionTokens.CreateAppSessionToken(sToken)
	if err != nil {
		return "", nil, err
	}

	return token, sToken, err
}
