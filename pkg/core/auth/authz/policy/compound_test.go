package policy

import (
	"errors"
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	. "github.com/smartystreets/goconvey/convey"
)

type everybody struct {
	Allow bool
}

func (p everybody) IsAllowed(r *http.Request, ctx auth.ContextGetter) error {
	if !p.Allow {
		return errors.New("denied")
	}

	return nil
}

func TestAllOfPolicy(t *testing.T) {
	Convey("Test AllOfPolicy", t, func() {
		Convey("should pass if all pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AllOf(
				everybody{Allow: true},
				everybody{Allow: true},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if one of them return error", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AllOf(
				everybody{Allow: true},
				everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})
	})

	Convey("Test AnyOfPolicy", t, func() {
		Convey("should pass if all pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				everybody{Allow: true},
				everybody{Allow: true},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should pass if one of them pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				everybody{Allow: true},
				everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldBeEmpty)
		})

		Convey("should return error if none of them pass", func() {
			req, _ := http.NewRequest("POST", "/", nil)
			ctx := MemoryContextGetter{}

			err := AnyOf(
				everybody{Allow: false},
				everybody{Allow: false},
			).IsAllowed(req, ctx)
			So(err, ShouldNotBeEmpty)
		})
	})
}
