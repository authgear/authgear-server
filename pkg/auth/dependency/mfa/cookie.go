package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

const CookieName = "mfa_bearer_token"

type BearerTokenCookieConfiguration corehttp.CookieConfiguration

func (c *BearerTokenCookieConfiguration) WriteTo(rw http.ResponseWriter, value string) {
	(*corehttp.CookieConfiguration)(c).WriteTo(rw, value)
}

func (c *BearerTokenCookieConfiguration) Clear(rw http.ResponseWriter) {
	(*corehttp.CookieConfiguration)(c).Clear(rw)
}

func NewBearerTokenCookieConfiguration(
	r *http.Request,
	useInsecureCookie bool,
	sConfig config.SessionConfiguration,
	bearerTokenConfig config.AuthenticatorBearerTokenConfiguration,
) BearerTokenCookieConfiguration {
	cfg := BearerTokenCookieConfiguration{
		Name:   CookieName,
		Path:   "/_auth/mfa/bearer_token/authenticate",
		Secure: !useInsecureCookie,
	}

	maxAge := 86400 * bearerTokenConfig.ExpireInDays
	cfg.MaxAge = &maxAge

	if sConfig.CookieDomain != nil {
		cfg.Domain = *sConfig.CookieDomain
	} else {
		cfg.Domain = corehttp.CookieDomainFromETLDPlusOneWithoutPort(corehttp.GetHost(r))
	}

	return cfg
}
