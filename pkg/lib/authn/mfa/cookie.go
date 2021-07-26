package mfa

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieDef struct {
	Def *httputil.CookieDef
}

func NewDeviceTokenCookieDef(cfg *config.AuthenticationConfig) CookieDef {
	def := &httputil.CookieDef{
		NameSuffix: "mfa_device_token",
		Path:       "/",
		SameSite:   http.SameSiteStrictMode,
	}

	maxAge := int(cfg.DeviceToken.ExpireIn.Duration().Seconds())
	def.MaxAge = &maxAge

	return CookieDef{Def: def}
}
