package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

			UpdateCookie(w, cookie)
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

			UpdateCookie(w, cookie)
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

			UpdateCookie(w, cookie)
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

			UpdateCookie(w, cookie)
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

			UpdateCookie(w, cookie)
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
