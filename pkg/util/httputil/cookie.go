package httputil

import (
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// CookieDef defines a cookie that is written to the response.
// All cookies in our server expects to be created with this definition.
type CookieDef struct {
	// NameSuffix means the cookie could have prefix.
	NameSuffix string
	Path       string
	// Domain is omitted because it is controlled somewhere else.
	// Domain            string
	AllowScriptAccess bool
	SameSite          http.SameSite
	MaxAge            *int

	// This flag is the inverse of http cookie host-only-flag (RFC6265 section5.3.6), default false
	IsNonHostOnly bool
}

func (cd *CookieDef) HostOnly() bool {
	return !cd.IsNonHostOnly
}

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

// CookieDomainWithoutPort derives host from r.
// If host has port, the port is removed.
// If host-1 is longer than ETLD+1, host-1 is returned.
// If ETLD+1 cannot be derived, an empty string is returned.
// The return value never have port.
func CookieDomainWithoutPort(host string) string {
	// Trim the port if it is present.
	// We have to trim the port first.
	// Passing host:port to EffectiveTLDPlusOne confuses it.
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		ipv6Str := host[1 : len(host)-1]
		if ipv6 := net.ParseIP(ipv6Str); ipv6 != nil {
			return ""
		}
	}

	if ipv4or6 := net.ParseIP(host); ipv4or6 != nil {
		return ""
	}

	eTLDPlusOne, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return ""
	}

	// host has a valid ETLDPlusOne, it's safe to split the domain with .
	hostMinusOne := host[1+strings.Index(host, "."):]
	if len(hostMinusOne) > len(eTLDPlusOne) {
		return hostMinusOne
	}

	return eTLDPlusOne
}

type CookieManager struct {
	Request      *http.Request
	TrustProxy   bool
	CookiePrefix string
	CookieDomain string
}

// CookieName returns the full name, that is, CookiePrefix followed by NameSuffix.
func (f *CookieManager) CookieName(def *CookieDef) string {
	return f.CookiePrefix + def.NameSuffix
}

// GetCookie is wrapper around http.Request.Cookie, taking care of cookie name.
func (f *CookieManager) GetCookie(r *http.Request, def *CookieDef) (*http.Cookie, error) {
	cookieName := f.CookieName(def)
	return r.Cookie(cookieName)
}

// ValueCookie generates a cookie that when set, the cookie is set to the specified value.
func (f *CookieManager) ValueCookie(def *CookieDef, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     f.CookieName(def),
		Path:     def.Path,
		HttpOnly: !def.AllowScriptAccess,
		Value:    value,
	}

	secure := f.secure()
	cookie.Secure = secure
	cookie.SameSite = f.sameSite(def, secure)

	if !def.HostOnly() {
		cookie.Domain = f.CookieDomain
	}

	if def.MaxAge != nil {
		cookie.MaxAge = *def.MaxAge
	}

	return cookie
}

// ClearCookie generates a cookie that when set, the cookie is clear.
func (f *CookieManager) ClearCookie(def *CookieDef) *http.Cookie {
	emptyValue := ""
	cookie := f.ValueCookie(def, emptyValue)

	// Suppress the MaxAge attribute written by ValueCookie
	cookie.MaxAge = 0
	// This is the defacto way to ask the browser to clear the cookie.
	cookie.Expires = time.Unix(0, 0).UTC()

	return cookie
}

func (f *CookieManager) secure() bool {
	proto := GetProto(f.Request, f.TrustProxy)
	return proto == "https"
}

func (f *CookieManager) sameSite(def *CookieDef, secure bool) http.SameSite {
	if def.SameSite == http.SameSiteNoneMode &&
		!ShouldSendSameSiteNone(f.Request.UserAgent(), secure) {
		return 0
	}

	return def.SameSite
}
