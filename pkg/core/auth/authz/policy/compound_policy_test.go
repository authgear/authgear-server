package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompoundPolicy(t *testing.T) {
	Convey("Test RequireValidUserOrMasterKey", t, func() {
		Convey("should pass if valid user exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			req = req.WithContext(session.WithSession(
				req.Context(),
				&session.Session{},
				&authinfo.AuthInfo{
					ID: "ID",
				},
			))

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
