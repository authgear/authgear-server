package webapp

import "github.com/authgear/authgear-server/pkg/lib/config"

type CSRFCookieDef struct {
	Name string
}

func NewCSRFCookieDef(cfg *config.HTTPConfig) CSRFCookieDef {
	return CSRFCookieDef{
		Name: cfg.CookiePrefix + "csrf_token",
	}
}
