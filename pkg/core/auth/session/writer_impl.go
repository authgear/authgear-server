package session

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type writerImpl struct {
	authContext       auth.ContextGetter
	clientConfigs     map[string]config.APIClientConfiguration
	mfaConfiguration  config.MFAConfiguration
	useInsecureCookie bool
}

func NewWriter(
	authContext auth.ContextGetter,
	clientConfigs map[string]config.APIClientConfiguration,
	mfaConfiguration config.MFAConfiguration,
	useInsecureCookie bool,
) Writer {
	return &writerImpl{
		authContext:       authContext,
		clientConfigs:     clientConfigs,
		mfaConfiguration:  mfaConfiguration,
		useInsecureCookie: useInsecureCookie,
	}
}

func (w *writerImpl) WriteSession(rw http.ResponseWriter, accessToken *string, mfaBearerToken *string) {
	clientConfig := w.clientConfigs[w.authContext.AccessKey().ClientID]
	useCookie := clientConfig.SessionTransport == config.SessionTransportTypeCookie

	cookieSession := &http.Cookie{
		Name:     coreHttp.CookieNameSession,
		Path:     "/",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
	}
	cookieMFABearerToken := &http.Cookie{
		Name:     coreHttp.CookieNameMFABearerToken,
		Path:     "/_auth/mfa/bearer_token/authenticate",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
	}
	w.configureCookieSameSite(cookieSession, clientConfig.SameSite)
	w.configureCookieSameSite(cookieMFABearerToken, clientConfig.SameSite)

	if useCookie {
		cookieSession.Value = *accessToken
		*accessToken = ""
		cookieSession.MaxAge = int(time.Duration(clientConfig.AccessTokenLifetime).Seconds())

		if mfaBearerToken != nil {
			cookieMFABearerToken.Value = *mfaBearerToken
			*mfaBearerToken = ""
			cookieMFABearerToken.MaxAge = int(time.Duration(w.mfaConfiguration.BearerToken.ExpireInDays).Seconds())
		}
	} else {
		cookieSession.Expires = time.Unix(0, 0)
		cookieMFABearerToken.Expires = time.Unix(0, 0)
	}

	updateCookie(rw, cookieSession)
	if mfaBearerToken != nil {
		updateCookie(rw, cookieMFABearerToken)
	}
}

func (w *writerImpl) configureCookieSameSite(cookie *http.Cookie, sameSite config.SessionCookieSameSite) {
	switch sameSite {
	case config.SessionCookieSameSiteNone:
		cookie.SameSite = http.SameSiteDefaultMode
	case config.SessionCookieSameSiteLax:
		cookie.SameSite = http.SameSiteLaxMode
	case config.SessionCookieSameSiteStrict:
		cookie.SameSite = http.SameSiteStrictMode
	}
}

func (w *writerImpl) ClearSession(rw http.ResponseWriter) {
	updateCookie(rw, &http.Cookie{
		Name:     coreHttp.CookieNameSession,
		Path:     "/",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
		Expires:  time.Unix(0, 0),
	})
}

func updateCookie(rw http.ResponseWriter, cookie *http.Cookie) {
	header := rw.Header()
	resp := http.Response{Header: header}

	cookies := resp.Cookies()
	updated := false
	for i, c := range cookies {
		if c.Name == cookie.Name && c.Domain == cookie.Domain && c.Path == cookie.Path {
			cookies[i] = cookie
			updated = true
		}
	}
	if !updated {
		cookies = append(cookies, cookie)
	}

	setCookies := make([]string, len(cookies))
	for i, c := range cookies {
		setCookies[i] = c.String()
	}
	header["Set-Cookie"] = setCookies
}
