package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

//go:generate mockgen -source=cookie.go -destination=cookie_mock_test.go -package webapp

type CookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type SessionCookieDef struct {
	Def *httputil.CookieDef
}

func NewSessionCookieDef() SessionCookieDef {
	def := &httputil.CookieDef{
		// NOTE: We plan to add `authgear_` prefix for the cookies we write.
		// See DEV-2227
		NameSuffix:        "authgear_web_session",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode, // For resumption after redirecting from OAuth providers
		MaxAge:            nil,                   // Use HTTP session cookie; expires when browser closes
	}
	return SessionCookieDef{Def: def}
}

type ErrorTokenCookieDef struct {
	Def *httputil.CookieDef
}

func NewErrorTokenCookieDef() ErrorTokenCookieDef {
	def := &httputil.CookieDef{
		NameSuffix:        "web_error_token",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteLaxMode,
		MaxAge:            nil, // Use HTTP session cookie; expires when browser closes
	}
	return ErrorTokenCookieDef{Def: def}
}

type SignedUpCookieDef struct {
	Def *httputil.CookieDef
}

func NewSignedUpCookieDef() SignedUpCookieDef {
	long := int(duration.Long.Seconds())
	def := &httputil.CookieDef{
		NameSuffix:        "signed_up",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteLaxMode,
		MaxAge:            &long,
	}
	return SignedUpCookieDef{Def: def}
}
