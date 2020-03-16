package session

import (
	"net/http"
	"time"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const CookieName = "session"

func WriteCookie(rw http.ResponseWriter, token string, cfg CookieConfiguration) {
	cookieSession := &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		Domain:   cfg.Domain,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	}

	cookieSession.Value = token
	if cfg.MaxAge != nil {
		cookieSession.MaxAge = *cfg.MaxAge
	}

	corehttp.UpdateCookie(rw, cookieSession)
}

func ClearCookie(rw http.ResponseWriter, cfg CookieConfiguration) {
	corehttp.UpdateCookie(rw, &http.Cookie{
		Name:     CookieName,
		Path:     "/",
		Domain:   cfg.Domain,
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
}
