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
				Everybody{Allow: true},
				Everybody{Allow: true},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if one of them return error", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AllOf(
				Everybody{Allow: true},
				Everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})
	})

	Convey("Test AnyOfPolicy", t, func() {
		Convey("should pass if all pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				Everybody{Allow: true},
				Everybody{Allow: true},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should pass if one of them pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				Everybody{Allow: true},
				Everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if none of them pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				Everybody{Allow: false},
				Everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})
	})
}
