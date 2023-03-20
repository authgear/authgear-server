package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/uiparam"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

// ClientIDCookieDef is a HTTP session cookie.
var ClientIDCookieDef = &httputil.CookieDef{
	NameSuffix: "client_id",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// UILocalesCookieDef is a HTTP session cookie.
var UILocalesCookieDef = &httputil.CookieDef{
	NameSuffix: "ui_locales",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// StateCookieDef is a HTTP session cookie.
var StateCookieDef = &httputil.CookieDef{
	NameSuffix: "state",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

// XStateCookieDef is a HTTP session cookie.
var XStateCookieDef = &httputil.CookieDef{
	NameSuffix: "x_state",
	Path:       "/",
	SameSite:   http.SameSiteNoneMode,
}

type UIParamMiddleware struct {
	Cookies CookieManager
}

func (m *UIParamMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uiParam uiparam.T

		q := r.URL.Query()

		// client_id
		clientID := q.Get("client_id")
		if clientID != "" {
			httputil.UpdateCookie(w, m.Cookies.ValueCookie(ClientIDCookieDef, clientID))
		}
		if clientID == "" {
			if cookie, err := m.Cookies.GetCookie(r, ClientIDCookieDef); err == nil {
				clientID = cookie.Value
			}
		}
		uiParam.ClientID = clientID

		// ui_locales
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

		// state
		state := q.Get("state")
		if state != "" {
			httputil.UpdateCookie(w, m.Cookies.ValueCookie(StateCookieDef, state))
		}
		if state == "" {
			if cookie, err := m.Cookies.GetCookie(r, StateCookieDef); err == nil {
				state = cookie.Value
			}
		}
		uiParam.State = state

		// x_state
		xState := q.Get("x_state")
		if xState != "" {
			httputil.UpdateCookie(w, m.Cookies.ValueCookie(XStateCookieDef, xState))
		}
		if xState == "" {
			if cookie, err := m.Cookies.GetCookie(r, XStateCookieDef); err == nil {
				xState = cookie.Value
			}
		}
		uiParam.XState = xState

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
