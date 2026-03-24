package e2eclient

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"
)

func TestHostAwareCookieJar(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("unexpected jar error: %v", err)
	}

	hostAwareJar := &HostAwareCookieJar{
		Jar:           jar,
		CorrectedHost: "app.authgeare2e.localhost:4000",
	}

	projectURL, err := url.Parse("http://app.authgeare2e.localhost:4000/")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	localURL, err := url.Parse("http://127.0.0.1:4000/")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	hostAwareJar.SetCookies(projectURL, []*http.Cookie{
		{
			Name:  "session",
			Value: "ok",
			Path:  "/",
		},
	})

	if got := jar.Cookies(localURL); len(got) != 0 {
		t.Fatalf("expected plain jar lookup by local URL to miss project cookie, got %v", got)
	}

	got := hostAwareJar.Cookies(localURL)
	if len(got) != 1 || got[0].Name != "session" || got[0].Value != "ok" {
		t.Fatalf("expected host-aware jar to return project cookie, got %v", got)
	}
}
