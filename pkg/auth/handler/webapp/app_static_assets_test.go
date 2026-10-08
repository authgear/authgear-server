package webapp

import (
	"context"
	// nolint:gosec
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type stubAppStaticAssetsResourceManager struct {
	files map[string][]byte
}

func (m *stubAppStaticAssetsResourceManager) Read(ctx context.Context, desc resource.Descriptor, view resource.View) (any, error) {
	p := view.(resource.EffectiveFileView).EffectiveFilePath()
	data, ok := m.files[p]
	if !ok {
		return nil, resource.ErrResourceNotFound
	}
	return data, nil
}

func TestAppStaticAssetsHandler(t *testing.T) {
	Convey("AppStaticAssetsHandler", t, func() {
		hashOf := func(b []byte) string {
			// nolint:gosec
			return fmt.Sprintf("%x", md5.Sum(b))
		}

		css := []byte("body {}")
		cssHash := hashOf(css)
		logo := []byte("logo")
		unknownImage := []byte("unknown image")
		unknownText := []byte("unknown text")

		h := &AppStaticAssetsHandler{
			Resources: &stubAppStaticAssetsResourceManager{
				files: map[string][]byte{
					"static/authgear-light-theme.css": css,
					"static/en/app_logo.png":          logo,
					"static/unknown.png":              unknownImage,
					"static/unknown.txt":              unknownText,
					"authgear.yaml":                   []byte("id: accounts\n"),
					"deno/hook.ts":                    []byte("export default {}\n"),
				},
			},
		}

		serve := func(target string) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodGet, target, nil)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			return rw
		}

		Convey("serves a static asset with its content hash", func() {
			rw := serve("/static/authgear-light-theme." + cssHash + ".css")
			So(rw.Code, ShouldEqual, http.StatusOK)
			So(rw.Body.String(), ShouldEqual, string(css))
		})

		Convey("serves a localized static asset with its content hash", func() {
			rw := serve("/static/en/app_logo." + hashOf(logo) + ".png")
			So(rw.Code, ShouldEqual, http.StatusOK)
			So(rw.Body.String(), ShouldEqual, string(logo))
		})

		Convey("does not serve a file that is not a known static asset", func() {
			for _, target := range []string{
				"/static/unknown." + hashOf(unknownImage) + ".png",
				"/static/unknown." + hashOf(unknownText) + ".txt",
			} {
				rw := serve(target)
				So(rw.Code, ShouldEqual, http.StatusNotFound)
				So(rw.Body.Len(), ShouldEqual, 0)
			}
		})

		Convey("does not serve a static asset whose hash does not match its content", func() {
			for _, target := range []string{
				"/static/authgear-light-theme.deadbeef.css",
				"/static/authgear-light-theme.00000000000000000000000000000000.css",
				"/static/authgear-light-theme.x.css",
				"/static/authgear-light-theme.css",
			} {
				rw := serve(target)
				So(rw.Code, ShouldEqual, http.StatusNotFound)
				So(rw.Body.Len(), ShouldEqual, 0)
			}
		})

		Convey("does not serve files outside the static directory", func() {
			for _, target := range []string{
				"/static/..%2fauthgear.deadbeef.yaml",
				"/static/../authgear.deadbeef.yaml",
				"/static/..%2fdeno%2fhook.deadbeef.ts",
				"/static/foo/..%2f..%2fauthgear.deadbeef.yaml",
				"/authgear.deadbeef.yaml",
			} {
				rw := serve(target)
				So(rw.Code, ShouldEqual, http.StatusNotFound)
				So(rw.Body.Len(), ShouldEqual, 0)
			}
		})
	})
}
