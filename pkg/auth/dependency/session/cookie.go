package session

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/config"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const CookieName = "session"

// FIXME(session): domain, session/permanent cookies

func WriteCookie(rw http.ResponseWriter, token string, useInsecureCookie bool, cfg config.APIClientConfiguration) {
	cookieSession := &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   !useInsecureCookie,
		SameSite: http.SameSiteLaxMode,
	}

	cookieSession.Value = token
	// FIXME(session): use session lifetime
	cookieSession.MaxAge = cfg.RefreshTokenLifetime

	corehttp.UpdateCookie(rw, cookieSession)
}

func ClearCookie(rw http.ResponseWriter, useInsecureCookie bool) {
	corehttp.UpdateCookie(rw, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   !useInsecureCookie,
		Expires:  time.Unix(0, 0),
	})
}
