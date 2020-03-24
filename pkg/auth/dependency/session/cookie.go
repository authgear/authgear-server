package session

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const CookieName = "session"

type CookieConfiguration corehttp.CookieConfiguration

func (c *CookieConfiguration) WriteTo(rw http.ResponseWriter, value string) {
	(*corehttp.CookieConfiguration)(c).WriteTo(rw, value)
}

func (c *CookieConfiguration) Clear(rw http.ResponseWriter) {
	(*corehttp.CookieConfiguration)(c).Clear(rw)
}

func NewSessionCookieConfiguration(r *http.Request, useInsecureCookie bool, sConfig config.SessionConfiguration) CookieConfiguration {
	cfg := CookieConfiguration{Name: CookieName, Path: "/", Secure: !useInsecureCookie}

	if sConfig.CookieNonPersistent {
		// HTTP session cookie: no MaxAge
		cfg.MaxAge = nil
	} else {
		// HTTP permanent cookie: MaxAge = session lifetime
		maxAge := sConfig.Lifetime
		cfg.MaxAge = &maxAge
	}

	if sConfig.CookieDomain != nil {
		cfg.Domain = *sConfig.CookieDomain
	} else {
		cfg.Domain = corehttp.CookieDomainFromETLDPlusOneWithoutPort(corehttp.GetHost(r))
	}

	return cfg
}
