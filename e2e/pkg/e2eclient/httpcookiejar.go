package e2eclient

import (
	"net/http"
	"net/url"
)

// HostAwareCookieJar makes cookie lookup consistent when the client connects to
// the local listen address but uses Request.Host for project routing.
type HostAwareCookieJar struct {
	Jar           http.CookieJar
	CorrectedHost string
}

var _ http.CookieJar = &HostAwareCookieJar{}

func (j *HostAwareCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.Jar.SetCookies(j.fixURL(u), cookies)
}

func (j *HostAwareCookieJar) Cookies(u *url.URL) []*http.Cookie {
	return j.Jar.Cookies(j.fixURL(u))
}

func (j *HostAwareCookieJar) fixURL(u *url.URL) *url.URL {
	uu := *u
	uu.Host = j.CorrectedHost
	return &uu
}
