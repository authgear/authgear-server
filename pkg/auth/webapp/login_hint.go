package webapp

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/anonymous"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/intents"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type anonymousTokenInput struct{ JWT string }

func (i *anonymousTokenInput) GetAnonymousRequestToken() string { return i.JWT }

type AnonymousIdentityProvider interface {
	ParseRequestUnverified(requestJWT string) (r *anonymous.Request, err error)
}

type LoginHintPageService interface {
	PostWithIntent(session *Session, intent interaction.Intent, inputFn func() (interface{}, error)) (*Result, error)
}

type LoginHintHandler struct {
	Config           *config.OAuthConfig
	Anonymous        AnonymousIdentityProvider
	OfflineGrants    oauth.OfflineGrantStore
	AppSessionTokens oauth.AppSessionTokenStore
	AppSessions      oauth.AppSessionStore
	Clock            clock.Clock
	Cookies          CookieManager
	Pages            LoginHintPageService
}

type HandleLoginHintOptions struct {
	SessionOptions      SessionOptions
	LoginHint           string
	UILocales           string
	OriginalRedirectURI string
}

func (r *LoginHintHandler) HandleLoginHint(options HandleLoginHintOptions) (httputil.Result, error) {
	loginHint := options.LoginHint
	if !strings.HasPrefix(loginHint, "https://authgear.com/login_hint?") {
		return nil, fmt.Errorf("invalid login_hint: %v", loginHint)
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

		switch request.Action {
		case anonymous.RequestActionPromote:
			intent := &intents.IntentAuthenticate{
				Kind: intents.IntentAuthenticateKindPromote,
			}
			inputer := func() (interface{}, error) {
				return &anonymousTokenInput{JWT: jwt}, nil
			}
			now := r.Clock.NowUTC()
			sessionOpts := options.SessionOptions
			sessionOpts.UpdatedAt = now
			session := NewSession(sessionOpts)
			result, err := r.Pages.PostWithIntent(session, intent, inputer)
			if err != nil {
				return nil, err
			}

			if result != nil {
				result.UILocales = options.UILocales
			}
			return result, nil
		case anonymous.RequestActionAuth:
			// TODO(webapp): support anonymous auth
			panic("webapp: anonymous auth through web app is not supported")
		default:
			return nil, errors.New("unknown anonymous request action")
		}
	case "app_session_token":
		token, err := r.resolveAppSessionToken(query.Get("app_session_token"))
		if err != nil {
			return nil, err
		}

		cookie := r.Cookies.ValueCookie(session.AppSessionTokenCookieDef, token)
		return &Result{
			Cookies:     []*http.Cookie{cookie},
			RedirectURI: options.OriginalRedirectURI,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported login hint type: %s", query.Get("type"))
	}
}

func (r *LoginHintHandler) resolveAppSessionToken(token string) (string, error) {
	// Redeem app session token
	sToken, err := r.AppSessionTokens.GetAppSessionToken(oauth.HashToken(token))
	if err != nil {
		return "", err
	}

	offlineGrant, err := r.OfflineGrants.GetOfflineGrant(sToken.OfflineGrantID)
	if err != nil {
		return "", err
	}

	expiry, err := oauth.ComputeOfflineGrantExpiryWithClients(offlineGrant, r.Config)
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
		ExpireAt:       expiry,
		TokenHash:      oauth.HashToken(token),
	}
	err = r.AppSessions.CreateAppSession(appSession)
	if err != nil {
		return "", err
	}

	return token, nil
}
