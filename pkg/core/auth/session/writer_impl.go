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
	useInsecureCookie bool
}

func NewWriter(
	authContext auth.ContextGetter,
	clientConfigs map[string]config.APIClientConfiguration,
	useInsecureCookie bool,
) Writer {
	return &writerImpl{
		authContext:       authContext,
		clientConfigs:     clientConfigs,
		useInsecureCookie: useInsecureCookie,
	}
}

func (w *writerImpl) WriteSession(rw http.ResponseWriter, accessToken *string) {
	clientConfig := w.clientConfigs[w.authContext.AccessKey().ClientID]
	useCookie := clientConfig.SessionTransport == config.SessionTransportTypeCookie

	cookie := &http.Cookie{
		Name:     coreHttp.CookieNameSession,
		Path:     "/",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
	}
	switch clientConfig.SameSite {
	case config.SessionCookieSameSiteNone:
		cookie.SameSite = http.SameSiteDefaultMode
	case config.SessionCookieSameSiteLax:
		cookie.SameSite = http.SameSiteLaxMode
	case config.SessionCookieSameSiteStrict:
		cookie.SameSite = http.SameSiteStrictMode
	}

	if useCookie {
		token := *accessToken
		*accessToken = ""

		cookie.Value = token
		cookie.MaxAge = int(time.Duration(clientConfig.AccessTokenLifetime).Seconds())
	} else {
		cookie.Expires = time.Unix(0, 0)
	}

	updateCookie(rw, cookie)
}

func (w *writerImpl) ClearSession(rw http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     coreHttp.CookieNameSession,
		Path:     "/",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
		Expires:  time.Unix(0, 0),
	}
	updateCookie(rw, cookie)
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
