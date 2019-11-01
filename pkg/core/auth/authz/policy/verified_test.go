package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyNotVerifiedUser(t *testing.T) {
	Convey("Test denyNotVerifiedUser", t, func() {
		Convey("should not return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := denyNotVerifiedUser(req, ctx)
			So(err, ShouldBeNil)
		})

		Convey("should return error if user is not verified", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:       "ID",
					Disabled: true,
				},
			}

			err := denyNotVerifiedUser(req, ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("should pass if user is verified", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID:       "ID",
					Verified: true,
				},
			}

			err := denyNotVerifiedUser(req, ctx)
			So(err, ShouldBeNil)
		})

	})
}
