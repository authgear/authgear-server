package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

const CookieName = "session"

type CookieDef struct {
	*httputil.CookieDef
}

func NewSessionCookieDef(r *http.Request, sessionCfg *config.SessionConfig, serverCfg *config.ServerConfig) CookieDef {
	secure := httputil.GetProto(r, serverCfg.TrustProxy) == "https"
	def := &httputil.CookieDef{Name: CookieName, Path: "/", Secure: secure}

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
	} else {
		def.Domain = httputil.CookieDomainFromETLDPlusOneWithoutPort(httputil.GetHost(r, serverCfg.TrustProxy))
	}

	return CookieDef{def}
}
