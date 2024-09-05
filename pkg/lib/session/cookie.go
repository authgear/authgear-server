package session

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type CookieDef struct {
	Def               *httputil.CookieDef
	SameSiteStrictDef *httputil.CookieDef
}

func NewSessionCookieDef(sessionCfg *config.SessionConfig) CookieDef {
	def := &httputil.CookieDef{
		NameSuffix:    "session",
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		IsNonHostOnly: true,
	}

	strictDef := &httputil.CookieDef{
		NameSuffix:    "same_site_strict",
		Path:          "/",
		SameSite:      http.SameSiteStrictMode,
		IsNonHostOnly: true,
	}

	if sessionCfg.CookieNonPersistent {
		// HTTP session cookie: no MaxAge
		def.MaxAge = nil
		strictDef.MaxAge = nil
	} else {
		// HTTP permanent cookie: MaxAge = session lifetime
		maxAge := int(sessionCfg.Lifetime)
		def.MaxAge = &maxAge
		strictDef.MaxAge = &maxAge
	}

	return CookieDef{
		Def:               def,
		SameSiteStrictDef: strictDef,
	}
}

var AppSessionTokenCookieDef = &httputil.CookieDef{
	NameSuffix:    "app_session",
	Path:          "/",
	SameSite:      http.SameSiteLaxMode,
	IsNonHostOnly: true,
}

var AppAccessTokenCookieDef = &httputil.CookieDef{
	NameSuffix:    "app_access_token",
	Path:          "/",
	SameSite:      http.SameSiteLaxMode,
	IsNonHostOnly: true,
}
