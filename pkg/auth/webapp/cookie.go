package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

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
		NameSuffix:        "web_session",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode, // For resumption after redirecting from OAuth providers
		MaxAge:            nil,                   // Use HTTP session cookie; expires when browser closes
	}
	return SessionCookieDef{Def: def}
}

type ErrorCookieDef struct {
	Def *httputil.CookieDef
}

func NewErrorCookieDef() ErrorCookieDef {
	def := &httputil.CookieDef{
		NameSuffix:        "web_err",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteLaxMode,
		MaxAge:            nil, // Use HTTP session cookie; expires when browser closes
	}
	return ErrorCookieDef{Def: def}
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
