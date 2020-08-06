package session

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/httputil"
)

const CookieName = "session"

type CookieDef struct {
	Def *httputil.CookieDef
}

func NewSessionCookieDef(r *http.Request, sessionCfg *config.SessionConfig, serverCfg *config.ServerConfig) CookieDef {
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
