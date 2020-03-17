package http

import (
	"net/http"
	"time"
)

type CookieConfiguration struct {
	Name   string
	Path   string
	Domain string
	Secure bool
	MaxAge *int
}

func (c *CookieConfiguration) WriteTo(rw http.ResponseWriter, value string) {
	cookie := &http.Cookie{
		Name:     c.Name,
		Path:     c.Path,
		Domain:   c.Domain,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
	}

	cookie.Value = c.Name
	if c.MaxAge != nil {
		cookie.MaxAge = *c.MaxAge
	}

	UpdateCookie(rw, cookie)
}

func (c *CookieConfiguration) Clear(rw http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     c.Name,
		Path:     c.Path,
		Domain:   c.Domain,
		HttpOnly: true,
		Secure:   c.Secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}

	UpdateCookie(rw, cookie)
}

// Cookie names
const (
	CookieNameSession = "session"
	// nolint: gosec
	CookieNameMFABearerToken = "mfa_bearer_token"
)

func UpdateCookie(w http.ResponseWriter, cookie *http.Cookie) {
	header := w.Header()
	resp := http.Response{Header: header}
	cookies := resp.Cookies()
	updated := false
	for i, c := range cookies {
		if c.Name == cookie.Name && c.Domain == cookie.Domain && c.Path == cookie.Path {
			cookies[i] = cookie
			updated = true
		}
	}
	if !updated {
		cookies = append(cookies, cookie)
	}
	setCookies := make([]string, len(cookies))
	for i, c := range cookies {
		setCookies[i] = c.String()
	}
	header["Set-Cookie"] = setCookies
}
