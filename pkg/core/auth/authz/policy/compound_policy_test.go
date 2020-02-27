package policy

import (
	"net/http"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCompoundPolicy(t *testing.T) {
	Convey("Test RequireValidUserOrMasterKey", t, func() {
		Convey("should pass if valid user exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID: "ID",
				},
				mSession: &auth.Session{
					AccessTokenCreatedAt: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			err := RequireValidUserOrMasterKey.IsAllowed(req, ctx)
			So(err, ShouldBeNil)
		})

		Convey("should pass if master key is used", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAccessKey: model.AccessKey{
					Type:     model.MasterAccessKeyType,
					ClientID: "",
				},
			}

			err := RequireValidUserOrMasterKey.IsAllowed(req, ctx)
			So(err, ShouldBeNil)
		})

		Convey("should fail if no user", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := RequireValidUserOrMasterKey.IsAllowed(req, ctx)
			So(err, ShouldBeError, "authentication required")
		})
	})
}
