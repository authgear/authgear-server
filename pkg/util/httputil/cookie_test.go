package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/authgear/authgear-server/pkg/util/httputil"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUpdateCookie(t *testing.T) {
	Convey("UpdateCookie", t, func() {
		Convey("append new cookie", func() {
			w := httptest.NewRecorder()

			cookie := &http.Cookie{
				Name:  "a",
				Value: "b",
			}

			httputil.UpdateCookie(w, cookie)
			So(w.Header(), ShouldResemble, http.Header{
				"Set-Cookie": []string{"a=b"},
			})
		})

		Convey("update existing cookie", func() {
			w := httptest.NewRecorder()
			header := w.Header()
			header["Set-Cookie"] = []string{"a=b"}

			cookie := &http.Cookie{
				Name:  "a",
				Value: "c",
			}

			httputil.UpdateCookie(w, cookie)
			So(w.Header(), ShouldResemble, http.Header{
				"Set-Cookie": []string{"a=c"},
			})
		})

		Convey("update non host-only cookie", func() {
			w := httptest.NewRecorder()
			header := w.Header()
			header["Set-Cookie"] = []string{"a=b", "a=b; Domain=example.com"}

			cookie := &http.Cookie{
				Name:   "a",
				Value:  "c",
				Domain: "example.com",
			}

			httputil.UpdateCookie(w, cookie)
			So(w.Header(), ShouldResemble, http.Header{
				"Set-Cookie": []string{"a=b", "a=c; Domain=example.com"},
			})
		})

		Convey("update path-set cookie", func() {
			w := httptest.NewRecorder()
			header := w.Header()
			header["Set-Cookie"] = []string{"a=b", "a=b; Path=/"}

			cookie := &http.Cookie{
				Name:  "a",
				Value: "c",
				Path:  "/",
			}

			httputil.UpdateCookie(w, cookie)
			So(w.Header(), ShouldResemble, http.Header{
				"Set-Cookie": []string{"a=b", "a=c; Path=/"},
			})
		})

		Convey("update non host-only path-set cookie", func() {
			w := httptest.NewRecorder()
			header := w.Header()
			header["Set-Cookie"] = []string{
				"a=b",
				"a=b; Domain=example.com",
				"a=b; Path=/",
				"a=b; Path=/; Domain=example.com",
			}

			cookie := &http.Cookie{
				Name:   "a",
				Value:  "c",
				Domain: "example.com",
				Path:   "/",
			}

			httputil.UpdateCookie(w, cookie)
			So(w.Header(), ShouldResemble, http.Header{
				"Set-Cookie": []string{
					"a=b",
					"a=b; Domain=example.com",
					"a=b; Path=/",
					"a=c; Path=/; Domain=example.com",
				},
			})
		})
	})
}

func TestCookieDomainWithoutPort(t *testing.T) {
	Convey("CookieDomainWithoutPort", t, func() {
		check := func(in string, out string) {
			actual := httputil.CookieDomainWithoutPort(in)
			So(out, ShouldEqual, actual)
		}
		check("localhost", "")
		check("localhost:3000", "")

		check("[::1]:3000", "")
		check("[::1]", "")

		check("10.0.2.2:3000", "")
		check("10.0.2.2", "")

		check("accounts.localhost", "accounts.localhost")
		check("accounts.localhost:8081", "accounts.localhost")

		check("example.com", "example.com")
		check("example.com:80", "example.com")
		check("example.com:8080", "example.com")

		check("www.example.com", "example.com")
		check("www.example.com:80", "example.com")
		check("www.example.com:8080", "example.com")

		check("example.co.jp", "example.co.jp")
		check("example.co.jp:80", "example.co.jp")
		check("example.co.jp:8080", "example.co.jp")

		check("www.example.co.jp", "example.co.jp")
		check("www.example.co.jp:80", "example.co.jp")
		check("www.example.co.jp:8080", "example.co.jp")

		check("auth.app.example.co.jp", "app.example.co.jp")
		check("auth.app.example.co.jp:80", "app.example.co.jp")
		check("auth.app.example.co.jp:8080", "app.example.co.jp")
	})
}

func TestCookieManager(t *testing.T) {
	cookieHostOnly := &httputil.CookieDef{
		NameSuffix: "csrf_token",
		Path:       "/",
		SameSite:   http.SameSiteLaxMode,
	}

	cookieNonHostOnly := &httputil.CookieDef{
		NameSuffix:    "session",
		Path:          "/",
		SameSite:      http.SameSiteLaxMode,
		IsNonHostOnly: true,
	}

	age := int((20 * time.Minute).Seconds())
	cookieWithMaxAge := &httputil.CookieDef{
		NameSuffix: "web_session",
		Path:       "/",
		SameSite:   http.SameSiteLaxMode,
		MaxAge:     &age,
	}

	r, _ := http.NewRequest("GET", "", nil)
	r.Header.Set("X-Forwarded-Proto", "https")

	cm := &httputil.CookieManager{
		TrustProxy:   true,
		Request:      r,
		CookiePrefix: "prefix_",
		CookieDomain: "cookiedomain.com",
	}

	Convey("CookieManager.CookieName supports prefix", t, func() {
		cookie := cm.ValueCookie(cookieHostOnly, "test")
		So(cookie.String(), ShouldEqual, "prefix_csrf_token=test; Path=/; HttpOnly; Secure; SameSite=Lax")
	})

	Convey("CookieManager.ValueCookie supports host-only cookie", t, func() {
		cookie := cm.ValueCookie(cookieHostOnly, "test")
		So(cookie.String(), ShouldEqual, "prefix_csrf_token=test; Path=/; HttpOnly; Secure; SameSite=Lax")
	})

	Convey("CookieManager.ValueCookie supports non-host-only cookie", t, func() {
		cookie := cm.ValueCookie(cookieNonHostOnly, "test")
		So(cookie.String(), ShouldEqual, "prefix_session=test; Path=/; Domain=cookiedomain.com; HttpOnly; Secure; SameSite=Lax")
	})

	Convey("CookieManager.ValueCookie supports cookie with Max-Age", t, func() {
		cookie := cm.ValueCookie(cookieWithMaxAge, "test")
		So(cookie.String(), ShouldEqual, "prefix_web_session=test; Path=/; Max-Age=1200; HttpOnly; Secure; SameSite=Lax")
	})

	Convey("CookieManager.ClearCookie does not set Max-Age, but set Expires", t, func() {
		cookie := cm.ClearCookie(cookieWithMaxAge)
		So(cookie.String(), ShouldEqual, "prefix_web_session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:00 GMT; HttpOnly; Secure; SameSite=Lax")
	})
}
