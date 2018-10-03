package policy

import (
	"net/http"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authtoken"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRequireAuthenticated(t *testing.T) {
	Convey("should return error if auth context has no auth info", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{}

		err := RequireAuthenticated(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should return error if token is not valid", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		validSince := time.Date(2017, 10, 1, 0, 0, 0, 0, time.UTC)
		ctx := handler.AuthContext{
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

	Convey("should pass if valid auth info exist", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{
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
}
