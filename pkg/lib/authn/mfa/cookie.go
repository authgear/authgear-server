package mfa

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

const CookieName = "mfa_device_token"

type CookieDef struct {
	Def *httputil.CookieDef
}

func NewDeviceTokenCookieDef(r *http.Request, cfg *config.AuthenticationConfig) CookieDef {
	def := &httputil.CookieDef{
		Name:     CookieName,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	maxAge := int(cfg.DeviceToken.ExpireIn.Duration().Seconds())
	def.MaxAge = &maxAge

	return CookieDef{Def: def}
}
