package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/authn"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompoundPolicy(t *testing.T) {
	Convey("Test RequireValidUserOrMasterKey", t, func() {
		Convey("should pass if valid user exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			authninfo := &authn.Info{UserID: "user-id"}
			req = req.WithContext(authn.WithAuthn(req.Context(), authninfo, authninfo.User()))

			err := RequireValidUserOrMasterKey.IsAllowed(req)
			So(err, ShouldBeNil)
		})

		Convey("should pass if master key is used", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			req = req.WithContext(auth.WithAccessKey(req.Context(), auth.AccessKey{
				IsMasterKey: true,
			}))

			err := RequireValidUserOrMasterKey.IsAllowed(req)
			So(err, ShouldBeNil)
		})

		Convey("should fail if no user", func() {
			req, _ := http.NewRequest("POST", "/", nil)

			err := RequireValidUserOrMasterKey.IsAllowed(req)
			So(err, ShouldBeError, "authentication required")
		})
	})
}
