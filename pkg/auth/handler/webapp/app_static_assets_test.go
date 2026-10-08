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

	"github.com/authgear/authgear-server/pkg/lib/web"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

type stubAppStaticAssetsResourceManager struct {
	files map[string][]byte
}

func (m *stubAppStaticAssetsResourceManager) Resolve(path string) (resource.Descriptor, bool) {
	if _, ok := m.files[path]; !ok {
		return nil, false
	}
	return web.AuthgearLightThemeCSS, true
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
		css := []byte("body {}")
		// nolint:gosec
		cssHash := fmt.Sprintf("%x", md5.Sum(css))

		h := &AppStaticAssetsHandler{
			Resources: &stubAppStaticAssetsResourceManager{
				files: map[string][]byte{
					"static/authgear-light-theme.css": css,
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
