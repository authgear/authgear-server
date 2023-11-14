package web

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

// ErrorQueryKey is "q_error" so that it is not persisent across pages.
const ErrorQueryKey = "q_error"

type ErrorState struct {
	Form  url.Values
	Error *apierrors.APIError
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

type ErrorCookieCookieManager interface {
	GetCookie(r *http.Request, def *httputil.CookieDef) (*http.Cookie, error)
	ValueCookie(def *httputil.CookieDef, value string) *http.Cookie
	ClearCookie(def *httputil.CookieDef) *http.Cookie
}

type ErrorCookie struct {
	Cookie  ErrorCookieDef
	Cookies ErrorCookieCookieManager
}

func (c *ErrorCookie) GetError(r *http.Request) (*ErrorState, bool) {
	var value string

	// recoverable error takes procedence over persistent error.
	if cookie, cookieErr := c.Cookies.GetCookie(r, c.Cookie.Def); cookieErr == nil && cookie.Value != "" {
		value = cookie.Value
	} else if q := r.URL.Query(); q.Get(ErrorQueryKey) != "" {
		value = q.Get(ErrorQueryKey)
	}

	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, false
	}

	var errorState ErrorState
	if err := json.Unmarshal(data, &errorState); err != nil {
		return nil, false
	}
	return &errorState, true
}

func (c *ErrorCookie) ResetRecoverableError() *http.Cookie {
	cookie := c.Cookies.ClearCookie(c.Cookie.Def)
	return cookie
}

// SetRecoverableError stores the error in cookie and retains the form.
func (c *ErrorCookie) SetRecoverableError(r *http.Request, value *apierrors.APIError) (*http.Cookie, error) {
	data, err := json.Marshal(&ErrorState{
		Form:  r.Form,
		Error: value,
	})
	if err != nil {
		return nil, err
	}

	cookieValue := base64.RawURLEncoding.EncodeToString(data)
	cookie := c.Cookies.ValueCookie(c.Cookie.Def, cookieValue)
	return cookie, nil
}

// SetNonRecoverableError does NOT retain form.
func (c *ErrorCookie) SetNonRecoverableError(redirectURIString string, value *apierrors.APIError) (string, error) {
	data, err := json.Marshal(&ErrorState{
		Error: value,
	})
	if err != nil {
		return "", err
	}

	queryValue := base64.RawURLEncoding.EncodeToString(data)

	redirectURI, err := url.Parse(redirectURIString)
	if err != nil {
		return "", err
	}

	q := redirectURI.Query()
	q.Set(ErrorQueryKey, queryValue)
	redirectURI.RawQuery = q.Encode()

	return redirectURI.String(), nil
}
