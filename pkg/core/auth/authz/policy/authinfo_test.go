package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"

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
			req = req.WithContext(session.WithSession(
				req.Context(),
				&session.Session{},
				&authinfo.AuthInfo{
					ID: "ID",
				},
			))

			err := requireAuthenticated(req)
			So(err, ShouldBeEmpty)
		})
	})
}
