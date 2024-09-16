package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// CSRFFieldName is the same as the default, but public.
const CSRFFieldName = "gorilla.csrf.Token"

var CSRFCookieMaxAge = int(duration.UserInteraction.Seconds())

var CSRFCookieDef = &httputil.CookieDef{
	NameSuffix:        "csrf_token",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteNoneMode,
	MaxAge:            &CSRFCookieMaxAge,
}

var CSRFDebugCookieSameSiteOmitDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_omit",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteDefaultMode,
	MaxAge:            &CSRFCookieMaxAge,
}

var CSRFDebugCookieSameSiteNoneDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_none",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteNoneMode,
	MaxAge:            &CSRFCookieMaxAge,
}

var CSRFDebugCookieSameSiteLaxDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_lax",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteLaxMode,
	MaxAge:            &CSRFCookieMaxAge,
}

var CSRFDebugCookieSameSiteStrictDef = &httputil.CookieDef{
	NameSuffix:        "debug_csrf_same_site_strict",
	Path:              "/",
	AllowScriptAccess: false,
	SameSite:          http.SameSiteStrictMode,
	MaxAge:            &CSRFCookieMaxAge,
}

type CSRFDebugMiddleware struct {
	Cookies CookieManager
}
