package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/authn"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyDisabledUser(t *testing.T) {
	Convey("Test DenyDisabledUser", t, func() {
		Convey("should not return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)

			err := DenyDisabledUser(req)
			So(err, ShouldBeNil)
		})

		Convey("should return error if user is disabled", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			authninfo := &authn.Info{UserID: "user-id", UserDisabled: true}
			req = req.WithContext(authn.WithAuthn(req.Context(), authninfo, authninfo.User()))

			err := DenyDisabledUser(req)
			So(err, ShouldNotBeNil)
		})

		Convey("should pass if user is not disabled", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			authninfo := &authn.Info{UserID: "user-id", UserDisabled: false}
			req = req.WithContext(authn.WithAuthn(req.Context(), authninfo, authninfo.User()))

			err := DenyDisabledUser(req)
			So(err, ShouldBeNil)
		})

	})
}
