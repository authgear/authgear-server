package webapp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
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

type ErrorCookie struct {
	Cookie  ErrorCookieDef
	Cookies CookieManager
}

type ClientIDCookieDef struct {
	Def *httputil.CookieDef
}

func NewClientIDCookieDef() ClientIDCookieDef {
	def := &httputil.CookieDef{
		NameSuffix: "client_id",
		Path:       "/",
		SameSite:   http.SameSiteNoneMode,
	}
	return ClientIDCookieDef{Def: def}
}

func (c *ErrorCookie) GetError(r *http.Request) (*apierrors.APIError, bool) {
	cookie, err := c.Cookies.GetCookie(r, c.Cookie.Def)
	if err != nil || cookie.Value == "" {
		return nil, false
	}

	data, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, false
	}

	var apiError apierrors.APIError
	if err := json.Unmarshal(data, &apiError); err != nil {
		return nil, false
	}
	return &apiError, true
}

func (c *ErrorCookie) ResetError() *http.Cookie {
	cookie := c.Cookies.ClearCookie(c.Cookie.Def)
	return cookie
}

func (c *ErrorCookie) SetError(value *apierrors.APIError) (*http.Cookie, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	cookieValue := base64.RawURLEncoding.EncodeToString(data)
	cookie := c.Cookies.ValueCookie(c.Cookie.Def, cookieValue)
	return cookie, nil
}
