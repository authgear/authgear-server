package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

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

type CSRFDebugMiddleware struct {
	Cookies CookieManager
}

var CSRFDebugCookieMaxAge = int(duration.UserInteraction.Seconds())

// NOTE: SameSiteDefaultMode means do not emit attribute,
// ref: https://github.com/golang/go/blob/3e10c1ff8141fae6b4d35a42e2631e7830c79830/src/net/http/cookie.go#L279

var CSRFDebugCookieSameSiteOmitDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_omit",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteDefaultMode,
	MaxAge:            &CSRFDebugCookieMaxAge,
	IsNonHostOnly:     false,
}

var CSRFDebugCookieSameSiteNoneDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_none",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteNoneMode,
	MaxAge:            &CSRFDebugCookieMaxAge,
	IsNonHostOnly:     false,
}

var CSRFDebugCookieSameSiteLaxDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_lax",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteLaxMode,
	MaxAge:            &CSRFDebugCookieMaxAge,
	IsNonHostOnly:     false,
}

var CSRFDebugCookieSameSiteStrictDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_strict",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteStrictMode,
	MaxAge:            &CSRFDebugCookieMaxAge,
	IsNonHostOnly:     false,
}
