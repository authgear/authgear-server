package sso

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIFrameHandler(t *testing.T) {
	Convey("Test IFrameHandler", t, func() {
		ih := &IFrameHandler{}
		ih.IFrameHTMLProvider = sso.NewIFrameHTMLProvider(
			"https://api.example.com",
		)

		Convey("should use provided endpoint", func() {
			req, _ := http.NewRequest("GET", "", nil)
			resp := httptest.NewRecorder()
			ih.ServeHTTP(resp, req)

			apiEndpointPattern := `"https://api.example.com/_auth/sso/config"`
			matched, err := regexp.MatchString(apiEndpointPattern, resp.Body.String())

			So(err, ShouldBeNil)
			So(matched, ShouldBeTrue)
		})
	})
}
