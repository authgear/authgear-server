package golangbugs

import (
	"net/http"
	"net/url"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

type bugObserver struct {
	ObservedURL *url.URL
}

func (o *bugObserver) SetCookies(u *url.URL, cookies []*http.Cookie) {
	uu := *u
	o.ObservedURL = &uu
}

func (o *bugObserver) Cookies(u *url.URL) []*http.Cookie {
	return nil
}

var _ http.CookieJar = &bugObserver{}

// TestHTTPClientCookieJarBug tests if https://github.com/golang/go/issues/38988 is fixed.
// On Go 1.26+, http.Client should pass Request.Host to the CookieJar when it is set.
func TestHTTPClientCookieJarBug(t *testing.T) {
	jar := &bugObserver{}
	client := &http.Client{
		Jar: jar,
	}

	gock.InterceptClient(client)
	defer gock.Off()

	r, _ := http.NewRequest("GET", "http://127.0.0.1:4000/", nil)
	r.Host = "app.authgeare2e.localhost:4000"

	cookie := &http.Cookie{
		Name:   "name",
		Value:  "Value",
		Domain: "app.authgeare2e.localhost",
	}

	gock.New("http://127.0.0.1:4000/").
		Reply(200).
		SetHeader("Set-Cookie", cookie.String())
	defer func() { gock.Flush() }()

	_, err := client.Do(r)
	if err != nil {
		t.Errorf("unexpected error: %v\n", err)
	}

	if jar.ObservedURL.String() != "http://app.authgeare2e.localhost:4000/" {
		t.Errorf("expected CookieJar to observe Request.Host, got %q", jar.ObservedURL.String())
	}
}
