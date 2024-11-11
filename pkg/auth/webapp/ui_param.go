package webapp

import (
	"context"
	"errors"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/oauth/oauthsession"
	"github.com/authgear/authgear-server/pkg/lib/saml/samlsession"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

// ClientIDCookieDef is deprecated.
// var ClientIDCookieDef = &httputil.CookieDef{
// 	NameSuffix: "client_id",
// 	Path:       "/",
// 	SameSite:   http.SameSiteNoneMode,
// }

// UILocalesCookieDef is a HTTP session cookie.
var UILocalesCookieDef = &httputil.CookieDef{
	NameSuffix: "ui_locales",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// StateCookieDef is deprecated.
// var StateCookieDef = &httputil.CookieDef{
// 	NameSuffix: "state",
// 	Path:       "/",
// 	SameSite:   http.SameSiteNoneMode,
// }

// XStateCookieDef is deprecated.
// var XStateCookieDef = &httputil.CookieDef{
// 	NameSuffix: "x_state",
// 	Path:       "/",
// 	SameSite:   http.SameSiteNoneMode,
// }

type UIParamMiddleware struct {
	OAuthUIInfoResolver SessionMiddlewareOAuthUIInfoResolver
	OAuthSessions       SessionMiddlewareOAuthSessionService
	SAMLSessions        SessionMiddlewareSAMLSessionService
	Cookies             CookieManager
}

func (m *UIParamMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// restore uiparam from oauth session.
		var uiParam uiparam.T

		webSession := GetSession(r.Context())
		if webSession != nil {
			if webSession.OAuthSessionID != "" {
				p, ok := m.getUIParamFromOAuthSession(r.Context(), webSession.OAuthSessionID)
				if ok {
					uiParam = p
				}
			}

			if webSession.SAMLSessionID != "" {
				p, ok := m.getUIParamFromSAMLSession(r.Context(), webSession.SAMLSessionID)
				if ok {
					uiParam = p
				}
			}
		}

		// Allow overriding ui_locales with query.
		q := r.URL.Query()
		uiLocales := q.Get("ui_locales")
		if uiLocales != "" {
			httputil.UpdateCookie(w, m.Cookies.ValueCookie(UILocalesCookieDef, uiLocales))
		}
		if uiLocales == "" {
			if cookie, err := m.Cookies.GetCookie(r, UILocalesCookieDef); err == nil {
				uiLocales = cookie.Value
			}
		}
		uiParam.UILocales = uiLocales

		// Put uiParam into context
		ctx := r.Context()
		ctx = uiparam.WithUIParam(ctx, &uiParam)
		if uiParam.UILocales != "" {
			tags := intl.ParseUILocales(uiParam.UILocales)
			ctx = intl.WithPreferredLanguageTags(ctx, tags)
		}
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (m *UIParamMiddleware) getUIParamFromOAuthSession(ctx context.Context, oauthSessionID string) (
	uiParam uiparam.T, ok bool) {
	entry, err := m.OAuthSessions.Get(ctx, oauthSessionID)
	if err != nil && !errors.Is(err, oauthsession.ErrNotFound) {
		panic(err)
	}

	if entry != nil {
		uiInfo, err := m.OAuthUIInfoResolver.ResolveForUI(ctx, entry.T.AuthorizationRequest)
		if err != nil {
			panic(err)
		}

		return uiInfo.ToUIParam(), true
	}

	return uiParam, false
}

func (m *UIParamMiddleware) getUIParamFromSAMLSession(ctx context.Context, samlSessionID string) (
	uiParam uiparam.T, ok bool) {
	entry, err := m.SAMLSessions.Get(ctx, samlSessionID)
	if err != nil && !errors.Is(err, samlsession.ErrNotFound) {
		panic(err)
	}

	if entry != nil {
		uiParam = entry.UIInfo.ToUIParam()
		return uiParam, true
	}

	return uiParam, false
}
