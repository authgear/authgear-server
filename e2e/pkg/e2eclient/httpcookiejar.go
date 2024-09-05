package e2eclient

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type CustomJar struct {
	Jar *cookiejar.Jar
}

var _ http.CookieJar = &CustomJar{}

func NewCustomJar(jar *cookiejar.Jar) *CustomJar {
	return &CustomJar{Jar: jar}
}

func (j *CustomJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if j.Jar == nil {
		return
	}

	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		// This is a workaround for this bug
		// https://github.com/golang/go/issues/38988
		//
		// http.Client always pass request.URL to http.CookieJar.
		// But a more correct behavior should be passing a net.URL
		// with net.URL.Host = http.Request.Host (if http.Request.Host is non-zero)
		//
		// To work around this problem, we hardcode the cookie
		// with a buggy domain, that is request.URL.Host = "127.0.0.1"
		cookie.Domain = "127.0.0.1"
	}
	j.Jar.SetCookies(u, cookies)

}
func (j *CustomJar) Cookies(u *url.URL) []*http.Cookie {
	if j.Jar == nil {
		return []*http.Cookie{}
	}
	return j.Jar.Cookies(u)
}
