package webapp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

// ErrorQueryKey is "q_error" so that it is not persisent across pages.
const ErrorQueryKey = "q_error"

type ErrorState struct {
	Form  url.Values
	Error *apierrors.APIError
}

type ErrorCookie struct {
	Cookie  ErrorCookieDef
	Cookies CookieManager
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
func (c *ErrorCookie) SetNonRecoverableError(result *Result, value *apierrors.APIError) error {
	data, err := json.Marshal(&ErrorState{
		Error: value,
	})
	if err != nil {
		return err
	}

	queryValue := base64.RawURLEncoding.EncodeToString(data)

	redirectURI, err := url.Parse(result.RedirectURI)
	if err != nil {
		return err
	}

	q := redirectURI.Query()
	q.Set(ErrorQueryKey, queryValue)
	redirectURI.RawQuery = q.Encode()

	result.RedirectURI = redirectURI.String()
	return nil
}
