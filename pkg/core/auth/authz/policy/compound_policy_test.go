package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	authntesting "github.com/skygeario/skygear-server/pkg/core/authn/testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompoundPolicy(t *testing.T) {
	Convey("Test RequireValidUserOrMasterKey", t, func() {
		Convey("should pass if valid user exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			req = authntesting.WithAuthn().ToRequest(req)

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
