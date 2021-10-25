package webapp

import "github.com/authgear/authgear-server/pkg/lib/config"

// CSRFFieldName is the same as the default, but public.
const CSRFFieldName = "gorilla.csrf.Token"

type CSRFCookieDef struct {
	Name   string
	Domain string
}

func NewCSRFCookieDef(cfg *config.HTTPConfig) CSRFCookieDef {
	def := CSRFCookieDef{
		Name: cfg.CookiePrefix + "csrf_token",
	}

	if cfg.CookieDomain != nil {
		def.Domain = *cfg.CookieDomain
	}

	return def
}
