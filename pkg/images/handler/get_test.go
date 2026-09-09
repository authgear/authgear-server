package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/h2non/gock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
)

func TestGetHandler(t *testing.T) {
	Convey("GetHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gock.Intercept()
		defer gock.Off()

		directorMaker := NewMockDirectorMaker(ctrl)
		vipsDaemon := NewMockVipsDaemon(ctrl)

		router := httproute.NewRouter()
		h := GetHandler{
			DirectorMaker: directorMaker,
			VipsDaemon:    vipsDaemon,
		}
		router.Add(ConfigureGetRoute(httproute.Route{}), &h)

		Convey("Ignore any non-200 response", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			gock.New("http://localhost:3004").Reply(404)

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "")
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("return 404 for invalid options", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/invalid", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "")
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("strip upstream headers", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			vipsDaemon.EXPECT().Process(gomock.Any()).Times(1).Return(&vipsutil.Output{
				Data:          nil,
				FileExtension: "",
			}, nil)
			gock.New("http://localhost:3004").
				Reply(200).
				SetHeader("foobar", "42")

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Result().Header.Get("foobar"), ShouldBeEmpty)
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("set headers", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			vipsDaemon.EXPECT().Process(gomock.Any()).Times(1).Return(&vipsutil.Output{
				Data:          nil,
				FileExtension: ".jpeg",
			}, nil)
			gock.New("http://localhost:3004").
				Reply(200)

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Result().Header.Get("Content-Length"), ShouldEqual, "0")
			So(w.Result().Header.Get("Content-Type"), ShouldEqual, "image/jpeg")
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "public, immutable, max-age=900")
			So(w.Result().ContentLength, ShouldEqual, 0)
			So(gock.IsDone(), ShouldBeTrue)
		})
	})
}

func TestGetHandlerOriginalVariant(t *testing.T) {
	Convey("GetHandler original variant", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		directorMaker := NewMockDirectorMaker(ctrl)
		vipsDaemon := NewMockVipsDaemon(ctrl)

		// The upstream is a stand-in for cloud storage. It echoes back whatever
		// content type was recorded at upload, which is client controlled.
		var upstreamBody string
		var upstreamContentType string
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if upstreamContentType != "" {
				w.Header().Set("Content-Type", upstreamContentType)
			}
			w.WriteHeader(200)
			_, _ = io.WriteString(w, upstreamBody)
		}))
		defer upstream.Close()

		directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {
			u, _ := url.Parse(upstream.URL)
			r.URL.Scheme = u.Scheme
			r.URL.Host = u.Host
			r.URL.Path = "/blob"
			r.Host = u.Host
		})

		h := GetHandler{
			DirectorMaker: directorMaker,
			VipsDaemon:    vipsDaemon,
		}
		router := httproute.NewRouter()
		router.Add(ConfigureGetRoute(httproute.Route{}), &h)
		// Go through a real server so that net/http's own content sniffing
		// applies, exactly as it does in production.
		srv := httptest.NewServer(router.HTTPHandler())
		defer srv.Close()

		get := func() *http.Response {
			resp, err := http.Get(srv.URL + "/_images/app/objectid/original")
			So(err, ShouldBeNil)
			return resp
		}

		Convey("serves a real image inline under its own media type", func() {
			upstreamBody = "\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 32)
			upstreamContentType = "image/png"

			resp := get()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("Content-Type"), ShouldEqual, "image/png")
			So(resp.Header.Get("Content-Disposition"), ShouldBeEmpty)

			body, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, upstreamBody)
		})

		Convey("never serves an uploaded HTML file as a document", func() {
			// Without an explicit Content-Type, net/http sniffs the body and
			// declares text/html, which the browser executes on this origin.
			upstreamBody = "<html><body><script>alert(document.domain)</script></body></html>"
			upstreamContentType = "image/png"

			resp := get()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("Content-Type"), ShouldEqual, "application/octet-stream")
			So(resp.Header.Get("Content-Disposition"), ShouldEqual, "attachment")

			// The bytes are still served unchanged, just not as a document.
			body, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, upstreamBody)
		})

		Convey("does not trust the content type recorded at upload", func() {
			upstreamBody = "<html><body><script>alert(document.domain)</script></body></html>"
			upstreamContentType = "image/jpeg"

			resp := get()
			So(resp.Header.Get("Content-Type"), ShouldEqual, "application/octet-stream")
			So(resp.Header.Get("Content-Disposition"), ShouldEqual, "attachment")
		})

		Convey("does not serve SVG inline", func() {
			// Inert under Go's sniffing today, but that is an accident of the
			// sniff table rather than a guarantee.
			upstreamBody = `<svg xmlns="http://www.w3.org/2000/svg" onload="alert(1)"/>`
			upstreamContentType = "image/svg+xml"

			resp := get()
			So(resp.Header.Get("Content-Type"), ShouldEqual, "application/octet-stream")
			So(resp.Header.Get("Content-Disposition"), ShouldEqual, "attachment")
		})

		Convey("handles a body shorter than the sniff window", func() {
			upstreamBody = "GIF89a"
			upstreamContentType = "image/gif"

			resp := get()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("Content-Type"), ShouldEqual, "image/gif")

			body, err := io.ReadAll(resp.Body)
			So(err, ShouldBeNil)
			So(string(body), ShouldEqual, upstreamBody)
		})

		Convey("handles an empty body", func() {
			upstreamBody = ""
			upstreamContentType = ""

			resp := get()
			So(resp.StatusCode, ShouldEqual, 200)
			So(resp.Header.Get("Content-Type"), ShouldEqual, "application/octet-stream")
			So(resp.Header.Get("Content-Disposition"), ShouldEqual, "attachment")
		})
	})
}
