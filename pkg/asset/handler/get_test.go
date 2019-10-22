package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/h2non/gock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/cloudstorage"
)

func TestGetHandler(t *testing.T) {
	Convey("GetHandler", t, func() {
		h := &GetHandler{}
		provider := &cloudstorage.MockProvider{}
		h.CloudStorageProvider = provider

		Convey("remove x-skygear-* headers", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			provider.GetURL = &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/a",
			}

			gock.New("http://example.com").
				Get("/a").
				MatchHeader("x-skygear-a", "this-should-be-removed").
				Reply(200)

			req, _ := http.NewRequest("GET", "/a", nil)
			req.Header.Set("x-skygear-a", "this-should-be-removed")
			resp := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			// 502 means the request cannot be matched.
			So(resp.Code, ShouldEqual, 502)
		})

		Convey("remove Range and If-Range", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			provider.GetURL = &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/a",
			}

			gock.New("http://example.com").
				Get("/a").
				MatchHeader("Range", "range").
				MatchHeader("If-Range", "if-range").
				Reply(200)

			// has pipeline
			req, _ := http.NewRequest("GET", "/a?pipeline", nil)
			req.Header.Set("Range", "range")
			req.Header.Set("If-Range", "if-range")
			resp := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			// 502 means the request cannot be matched.
			So(resp.Code, ShouldEqual, 502)

			// no pipeline
			req, _ = http.NewRequest("GET", "/a", nil)
			req.Header.Set("Range", "range")
			req.Header.Set("If-Range", "if-range")
			resp = httptest.NewRecorder()
			router = mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
		})

		Convey("remove Accept-Ranges", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			provider.GetURL = &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/a",
			}

			gock.New("http://example.com").
				Get("/a").
				Persist().
				Reply(200).
				SetHeader("Accept-Ranges", "accept-ranges")

			// has pipeline
			req, _ := http.NewRequest("GET", "/a?pipeline", nil)
			resp := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Result().Header.Get("Accept-Ranges"), ShouldEqual, "")

			// no pipeline
			req, _ = http.NewRequest("GET", "/a", nil)
			resp = httptest.NewRecorder()
			router = mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
			So(resp.Result().Header.Get("Accept-Ranges"), ShouldEqual, "accept-ranges")
		})

		Convey("return 401 if no signature and private", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			provider.GetURL = &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/a",
			}
			provider.GetAccessType = cloudstorage.AccessTypePrivate
			provider.OriginallySigned = false

			gock.New("http://example.com").
				Get("/a").
				Persist().
				Reply(200)

			req, _ := http.NewRequest("GET", "/a", nil)
			resp := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 401)

			provider.OriginallySigned = true
			req, _ = http.NewRequest("GET", "/a", nil)
			resp = httptest.NewRecorder()
			router = mux.NewRouter()
			router.Handle("/{asset_name}", h)
			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, 200)
		})
	})
}
