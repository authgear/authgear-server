package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/httputil"
)

const CookieName = "session"

type CookieDef struct {
	*httputil.CookieDef
}

func NewSessionCookieDef(r *http.Request, useInsecureCookie bool, sConfig config.SessionConfiguration) CookieDef {
	def := &httputil.CookieDef{Name: CookieName, Path: "/", Secure: !useInsecureCookie}

	if sConfig.CookieNonPersistent {
		// HTTP session cookie: no MaxAge
		def.MaxAge = nil
	} else {
		// HTTP permanent cookie: MaxAge = session lifetime
		maxAge := sConfig.Lifetime
		def.MaxAge = &maxAge
	}

	if sConfig.CookieDomain != nil {
		def.Domain = *sConfig.CookieDomain
	} else {
		// FIXME: use ServerConfig
		def.Domain = httputil.CookieDomainFromETLDPlusOneWithoutPort(httputil.GetHost(r, true))
	}

	return CookieDef{def}
}
