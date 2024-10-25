package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCSPJoin(t *testing.T) {
	Convey("CSPJoin", t, func() {
		So(CSPJoin([]string{"a", "b"}), ShouldResemble, "a; b")
	})
}

func TestStaticCSPHeader(t *testing.T) {
	Convey("StaticCSPHeader", t, func() {
		middleware := StaticCSPHeader{}
		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := middleware.Handle(dummy)
			return h
		}

		Convey("csp directives", func() {
			middleware.CSPDirectives = CSPDirectives{
				CSPDirective{CSPDirectiveNameDefaultSrc, CSPSources{CSPSourceSelf}},
				CSPDirective{CSPDirectiveNameObjectSrc, CSPSources{CSPSourceNone}},
				CSPDirective{CSPDirectiveNameBaseURI, CSPSources{CSPSourceNone}},
				CSPDirective{CSPDirectiveNameScriptSrc, CSPSources{CSPSourceSelf}},
				CSPDirective{CSPDirectiveNameBlockAllMixedContent, CSPSources{}},
				CSPDirective{CSPDirectiveNameFrameAncestors, CSPSources{CSPSourceNone}},
			}
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			makeHandler().ServeHTTP(w, r)

			So(w.Result().Header.Get("Content-Security-Policy"), ShouldEqual, "default-src 'self'; object-src 'none'; base-uri 'none'; script-src 'self'; block-all-mixed-content; frame-ancestors 'none'")
		})
	})
}
