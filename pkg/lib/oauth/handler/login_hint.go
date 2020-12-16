package handler

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type LoginHintResolver struct {
	Anonymous        AnonymousIdentityProvider
	OfflineGrants    oauth.OfflineGrantStore
	AppSessionTokens oauth.AppSessionTokenStore
	AppSessions      oauth.AppSessionStore
	Clock            clock.Clock
}

func (r *LoginHintResolver) ResolveLoginHint(loginHint string) (interface{}, error) {
	if !strings.HasPrefix(loginHint, "https://authgear.com/login_hint?") {
		return nil, nil
	}

	u, err := url.Parse(loginHint)
	if err != nil {
		return nil, err
	}
	query := u.Query()

	switch query.Get("type") {
	case "anonymous":
		jwt := query.Get("jwt")
		request, err := r.Anonymous.ParseRequestUnverified(jwt)
		if err != nil {
			return nil, err
		}

		return webapp.AnonymousRequest{
			JWT:     jwt,
			Request: request,
		}, nil

	case "app_session_token":
		token, err := r.resolveAppSessionToken(query.Get("app_session_token"))
		if err != nil {
			// If app session token cannot be resolved: ignore and continue.
			return nil, nil
		}

		return webapp.RawSessionCookieRequest{
			Value: token,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported login hint type: %s", query.Get("type"))
	}
}

func (r *LoginHintResolver) resolveAppSessionToken(token string) (string, error) {
	// Redeem app session token
	sToken, err := r.AppSessionTokens.GetAppSessionToken(oauth.HashToken(token))
	if err != nil {
		return "", err
	}

	offlineGrant, err := r.OfflineGrants.GetOfflineGrant(sToken.OfflineGrantID)
	if err != nil {
		return "", err
	}

	err = r.AppSessionTokens.DeleteAppSessionToken(sToken)
	if err != nil {
		return "", err
	}

	// Create app session
	token = oauth.GenerateToken()
	appSession := &oauth.AppSession{
		AppID:          offlineGrant.AppID,
		OfflineGrantID: offlineGrant.ID,
		CreatedAt:      r.Clock.NowUTC(),
		ExpireAt:       offlineGrant.ExpireAt,
		TokenHash:      oauth.HashToken(token),
	}
	err = r.AppSessions.CreateAppSession(appSession)
	if err != nil {
		return "", err
	}

	return token, nil
}
