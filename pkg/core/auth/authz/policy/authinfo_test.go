package policy

import (
	"net/http"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"
	"github.com/skygeario/skygear-server/pkg/core/handler/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequireAuthenticated(t *testing.T) {
	Convey("Test RequireAuthenticated", t, func() {
		Convey("should return error if auth context has no auth info", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := context.AuthContext{}

			err := RequireAuthenticated(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should return error if token is not valid", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			validSince := time.Date(2017, 10, 1, 0, 0, 0, 0, time.UTC)
			ctx := context.AuthContext{
				AuthInfo: &authinfo.AuthInfo{
					ID:              "ID",
					TokenValidSince: &validSince,
				},
				Token: &authtoken.Token{
					IssuedAt: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			err := RequireAuthenticated(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

		Convey("should pass if valid auth info exist", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := context.AuthContext{
				AuthInfo: &authinfo.AuthInfo{
					ID: "ID",
				},
				Token: &authtoken.Token{
					IssuedAt: time.Date(2016, 10, 1, 0, 0, 0, 0, time.UTC),
				},
			}

			err := RequireAuthenticated(req, ctx)
			So(err, ShouldBeEmpty)
		})
	})
}
