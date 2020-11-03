package webapp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type UATokenCookieDef struct {
	Def *httputil.CookieDef
}

func NewUATokenCookieDef(httpCfg *config.HTTPConfig) UATokenCookieDef {
	def := &httputil.CookieDef{
		Name:              httpCfg.CookiePrefix + "ua_token",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteNoneMode, // Ensure resume-able after redirecting from external site
		MaxAge:            nil,                   // Use HTTP session cookie; expires when browser closes
	}

	if httpCfg.CookieDomain != nil {
		def.Domain = *httpCfg.CookieDomain
	}

	return UATokenCookieDef{Def: def}
}

type ErrorCookieDef struct {
	Def *httputil.CookieDef
}

func NewErrorCookieDef(httpCfg *config.HTTPConfig) ErrorCookieDef {
	def := &httputil.CookieDef{
		Name:              httpCfg.CookiePrefix + "web_err",
		Path:              "/",
		AllowScriptAccess: false,
		SameSite:          http.SameSiteLaxMode,
		MaxAge:            nil, // Use HTTP session cookie; expires when browser closes
	}

	if httpCfg.CookieDomain != nil {
		def.Domain = *httpCfg.CookieDomain
	}

	return ErrorCookieDef{Def: def}
}

type ErrorCookie struct {
	Cookie        ErrorCookieDef
	CookieFactory CookieFactory
}

func (c *ErrorCookie) GetError(r *http.Request) (*apierrors.APIError, bool) {
	cookie, err := r.Cookie(c.Cookie.Def.Name)
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
	cookie := c.CookieFactory.ClearCookie(c.Cookie.Def)
	return cookie
}

func (c *ErrorCookie) SetError(value *apierrors.APIError) (*http.Cookie, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	cookieValue := base64.RawURLEncoding.EncodeToString(data)
	cookie := c.CookieFactory.ValueCookie(c.Cookie.Def, cookieValue)
	return cookie, nil
}
