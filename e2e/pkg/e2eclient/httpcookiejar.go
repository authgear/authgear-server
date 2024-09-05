package e2eclient

import (
	"net/http"
	"net/url"
)

type JarWorkingAroundGolangIssue38988 struct {
	Jar           http.CookieJar
	CorrectedHost string
}

var _ http.CookieJar = &JarWorkingAroundGolangIssue38988{}

func (j *JarWorkingAroundGolangIssue38988) SetCookies(u *url.URL, cookies []*http.Cookie) {
	u = j.fixURL(u)
	j.Jar.SetCookies(u, cookies)

}
func (j *JarWorkingAroundGolangIssue38988) Cookies(u *url.URL) []*http.Cookie {
	u = j.fixURL(u)
	return j.Jar.Cookies(u)
}

func (j *JarWorkingAroundGolangIssue38988) fixURL(u *url.URL) *url.URL {
	// This is a workaround for this bug
	// https://github.com/golang/go/issues/38988
	//
	// http.Client always pass request.URL to http.CookieJar.
	// But a more correct behavior should be passing a net.URL
	// with net.URL.Host = http.Request.Host (if http.Request.Host is non-zero)
	//
	// To work around this problem, we correct request.URL.Host
	uu := *u
	uu.Host = j.CorrectedHost
	return &uu
}
