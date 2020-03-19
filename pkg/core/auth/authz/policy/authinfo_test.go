package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRequireAuthenticated(t *testing.T) {
	Convey("Test requireAuthenticated", t, func() {
		Convey("should return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)

			err := requireAuthenticated(req)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should pass if valid auth info exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			authninfo := &authn.Info{UserID: "user-id"}
			req = req.WithContext(authn.WithAuthn(req.Context(), authninfo, authninfo.User()))

			err := requireAuthenticated(req)
			So(err, ShouldBeEmpty)
		})
	})
}
