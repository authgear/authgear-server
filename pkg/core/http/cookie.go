package http

import (
	"net/http"
)

// Cookie names
const (
	CookieNameSSOData            = "sso_data"
	CookieNameOpenIDConnectNonce = "oidc_nonce"
	CookieNameSession            = "session"
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
