package policy

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAllOfPolicy(t *testing.T) {
	Convey("Test AllOfPolicy", t, func() {
		Convey("should pass if all pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AllOf(
				Everybody{allow: true},
				Everybody{allow: true},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if one of them return error", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AllOf(
				Everybody{allow: true},
				Everybody{allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})

	})
}
