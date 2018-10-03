package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAllOfPolicy(t *testing.T) {
	Convey("should pass if all pass", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{}

		err := AllOf(
			Everybody{allow: true},
			Everybody{allow: true},
		).IsAllowed(req, ctx)
		So(err, ShouldBeEmpty)
	})

	Convey("should return error if one of them return error", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := handler.AuthContext{}

		err := AllOf(
			Everybody{allow: true},
			Everybody{allow: false},
		).IsAllowed(req, ctx)
		So(err, ShouldNotBeEmpty)
	})
}
