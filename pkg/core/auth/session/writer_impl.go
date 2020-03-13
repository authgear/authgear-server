package session

import (
	"context"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type writerImpl struct {
	ctx               context.Context
	clientConfigs     []config.OAuthClientConfiguration
	mfaConfiguration  *config.MFAConfiguration
	useInsecureCookie bool
}

func NewWriter(
	ctx context.Context,
	clientConfigs []config.OAuthClientConfiguration,
	mfaConfiguration *config.MFAConfiguration,
	useInsecureCookie bool,
) Writer {
	return &writerImpl{
		ctx:               ctx,
		clientConfigs:     clientConfigs,
		mfaConfiguration:  mfaConfiguration,
		useInsecureCookie: useInsecureCookie,
	}
}

func (w *writerImpl) WriteSession(rw http.ResponseWriter, accessToken *string, mfaBearerToken *string) {
	clientConfig := auth.GetAccessKey(w.ctx).Client
	useCookie := clientConfig.AuthAPIUseCookie()

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
	cookieSession.SameSite = http.SameSiteLaxMode
	cookieMFABearerToken.SameSite = http.SameSiteLaxMode

	if useCookie {
		cookieSession.Value = *accessToken
		*accessToken = ""
		cookieSession.MaxAge = clientConfig.AccessTokenLifetime()

		if mfaBearerToken != nil {
			cookieMFABearerToken.Value = *mfaBearerToken
			*mfaBearerToken = ""
			cookieMFABearerToken.MaxAge = 86400 * w.mfaConfiguration.BearerToken.ExpireInDays
		}
	} else {
		cookieSession.Expires = time.Unix(0, 0)
		cookieMFABearerToken.Expires = time.Unix(0, 0)
	}

	coreHttp.UpdateCookie(rw, cookieSession)
	if mfaBearerToken != nil {
		coreHttp.UpdateCookie(rw, cookieMFABearerToken)
	}
}

func (w *writerImpl) ClearSession(rw http.ResponseWriter) {
	coreHttp.UpdateCookie(rw, &http.Cookie{
		Name:     coreHttp.CookieNameSession,
		Path:     "/",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
		Expires:  time.Unix(0, 0),
	})
}

func (w *writerImpl) ClearMFABearerToken(rw http.ResponseWriter) {
	coreHttp.UpdateCookie(rw, &http.Cookie{
		Name:     coreHttp.CookieNameMFABearerToken,
		Path:     "/_auth/mfa/bearer_token/authenticate",
		HttpOnly: true,
		Secure:   !w.useInsecureCookie,
		Expires:  time.Unix(0, 0),
	})
}
