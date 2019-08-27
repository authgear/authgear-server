package session

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/model"
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

func (w *writerImpl) WriteSession(rw http.ResponseWriter, resp *model.AuthResponse) {
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
		token := resp.AccessToken
		resp.AccessToken = ""

		cookie.Value = token
		cookie.MaxAge = int(time.Duration(clientConfig.AccessTokenLifetime).Seconds())
	} else {
		cookie.Expires = time.Unix(0, 0)
	}

	http.SetCookie(rw, cookie)
}

func (w *writerImpl) ClearSession(rw http.ResponseWriter) {
	clientConfig := w.clientConfigs[w.authContext.AccessKey().ClientID]
	useCookie := clientConfig.SessionTransport == config.SessionTransportTypeCookie
	if useCookie {
		cookie := &http.Cookie{
			Name:     coreHttp.CookieNameSession,
			Path:     "/",
			HttpOnly: true,
			Secure:   !w.useInsecureCookie,
			Expires:  time.Unix(0, 0),
		}
		http.SetCookie(rw, cookie)
	}
}
