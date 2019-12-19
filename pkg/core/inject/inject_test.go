package inject

import (
	"context"
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

type istore interface {
	get() string
}

type store struct {
	name string
}

func (s store) get() string {
	return s.name
}

type dmap struct{}

// Provide provides dependency instance by name
// nolint: golint
func (s dmap) Provide(
	dependencyName string,
	request *http.Request,
	ctx context.Context,
	requestID string,
	tConfig config.TenantConfiguration,
) interface{} {
	switch dependencyName {
	case "str":
		return "string"
	case "int":
		return 1
	case "store":
		return store{tConfig.AppName}
	case "istore":
		return &store{tConfig.AppName}
	default:
		return nil
	}
}

func TestInjectDependency(t *testing.T) {
	conf := config.TenantConfiguration{
		AppName: "TestApp",
	}

	req, _ := http.NewRequest("POST", "", nil)
	req = req.WithContext(config.WithTenantConfig(req.Context(), &conf))

	Convey("Test injectDependency", t, func() {
		Convey("should inject simple type", func() {
			type targetStruct struct {
				Str string `dependency:"str"`
				Int int    `dependency:"int"`
			}

			target := targetStruct{}
			DefaultRequestInject(&target, dmap{}, req)
			So(target.Str, ShouldEqual, "string")
			So(target.Int, ShouldEqual, 1)
		})

		Convey("should inject interface", func() {
			type targetStruct struct {
				Store istore `dependency:"istore"`
			}

			target := targetStruct{}
			DefaultRequestInject(&target, dmap{}, req)
			So(target.Store, ShouldImplement, (*istore)(nil))
			So(target.Store.get(), ShouldEqual, "TestApp")
		})

		Convey("should inject struct", func() {
			type targetStruct struct {
				Store store `dependency:"store"`
			}

			target := targetStruct{}
			DefaultRequestInject(&target, dmap{}, req)
			So(target.Store, ShouldHaveSameTypeAs, store{})
			So(target.Store.get(), ShouldEqual, "TestApp")
		})

		Convey("should not inject to with field without tag", func() {
			type targetStruct struct {
				Str string `dependency:"str"`
				str string
			}

			target := targetStruct{}
			DefaultRequestInject(&target, dmap{}, req)
			So(target.Str, ShouldEqual, "string")
			So(target.str, ShouldBeEmpty)
		})

		Convey("should panic if field type is wrong", func() {
			type targetStruct struct {
				Str int `dependency:"str"`
			}

			target := targetStruct{}
			So(func() {
				DefaultRequestInject(&target, dmap{}, req)
			}, ShouldPanic)
		})

		Convey("should return error dependency name is wrong", func() {
			type targetStruct struct {
				Str int `dependency:"i_am_your_father"`
			}

			target := targetStruct{}
			err := DefaultRequestInject(&target, dmap{}, req)
			So(err, ShouldBeError)
		})
	})
}
