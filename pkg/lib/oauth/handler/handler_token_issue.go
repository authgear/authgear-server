package handler

import (
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const AppSessionTokenDuration = duration.Short

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
