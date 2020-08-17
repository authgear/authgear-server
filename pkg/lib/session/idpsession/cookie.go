package idpsession

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieFactory interface {
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

const CookieName = "session"

type CookieDef struct {
	Def *httputil.CookieDef
}

func NewSessionCookieDef(sessionCfg *config.SessionConfig) CookieDef {
	def := &httputil.CookieDef{
		Name:     CookieName,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	if sessionCfg.CookieNonPersistent {
		// HTTP session cookie: no MaxAge
		def.MaxAge = nil
	} else {
		// HTTP permanent cookie: MaxAge = session lifetime
		maxAge := int(sessionCfg.Lifetime)
		def.MaxAge = &maxAge
	}

	if sessionCfg.CookieDomain != nil {
		def.Domain = *sessionCfg.CookieDomain
	}

	return CookieDef{Def: def}
}
