package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type CookieDef struct {
	Def *httputil.CookieDef
}

func NewUATokenCookieDef(httpCfg *config.HTTPConfig) CookieDef {
	def := &httputil.CookieDef{
		Name:              httpCfg.CookiePrefix + "ua_token",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode, // Ensure resume-able after redirecting from external site
		MaxAge:            nil,                   // Use HTTP session cookie; expires when browser closes
	}

	return CookieDef{Def: def}
}
