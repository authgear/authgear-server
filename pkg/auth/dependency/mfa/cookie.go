package mfa

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/config"
	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
	"golang.org/x/net/publicsuffix"
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
	mfaConfig config.MFAConfiguration,
) BearerTokenCookieConfiguration {
	cfg := BearerTokenCookieConfiguration{
		Name:   CookieName,
		Path:   "/_auth/mfa/bearer_token/authenticate",
		Secure: !useInsecureCookie,
	}

	maxAge := 86400 * mfaConfig.BearerToken.ExpireInDays
	cfg.MaxAge = &maxAge

	if sConfig.CookieDomain != nil {
		cfg.Domain = *sConfig.CookieDomain
	} else {
		host := corehttp.GetHost(r)
		etldp1, err := publicsuffix.EffectiveTLDPlusOne(host)
		if err != nil {
			// Failed to derive eTLD+1: use host-only cookie
			cfg.Domain = ""
		} else {
			cfg.Domain = etldp1
		}
	}

	return cfg
}
