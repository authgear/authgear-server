package session

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieFactory interface {
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type CookieDef struct {
	Def               *httputil.CookieDef
	SameSiteStrictDef *httputil.CookieDef
}

func NewSessionCookieDef(httpCfg *config.HTTPConfig, sessionCfg *config.SessionConfig) CookieDef {
	def := &httputil.CookieDef{
		Name:     httpCfg.CookiePrefix + "session",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	strictDef := &httputil.CookieDef{
		Name:     httpCfg.CookiePrefix + "same_site_strict",
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
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

	if httpCfg.CookieDomain != nil {
		def.Domain = *httpCfg.CookieDomain
		strictDef.Domain = *httpCfg.CookieDomain
	}

	return CookieDef{
		Def:               def,
		SameSiteStrictDef: strictDef,
	}
}
