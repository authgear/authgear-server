package policy

import (
	"net/http"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequireAuthenticated(t *testing.T) {
	Convey("Test requireAuthenticated", t, func() {
		Convey("should return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := requireAuthenticated(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should pass if valid auth info exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{
				mAuthInfo: &authinfo.AuthInfo{
					ID: "ID",
				},
				mSession: &auth.Session{
					AccessTokenCreatedAt: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			err := requireAuthenticated(req, ctx)
			So(err, ShouldBeEmpty)
		})
	})
}
