package inject

import (
	"context"
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
	. "github.com/smartystreets/goconvey/convey"
)

type contextKey string

var (
	configurationKey = contextKey("configuration")
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

func (s dmap) Provide(name string, req *http.Request) interface{} {
	conf := req.Context().Value(configurationKey).(config.TenantConfiguration)
	switch name {
	case "str":
		return "string"
	case "int":
		return 1
	case "store":
		return store{conf.AppName}
	case "istore":
		return &store{conf.AppName}
	default:
		return nil
	}
}

func TestInjectDependency(t *testing.T) {
	conf := config.TenantConfiguration{
		AppName: "TestApp",
	}

	req, _ := http.NewRequest("POST", "", nil)
	req = req.WithContext(context.WithValue(req.Context(), configurationKey, conf))

	Convey("Test injectDependency", t, func() {
		Convey("should inject simple type", func() {
			type targetStruct struct {
				Str string `dependency:"str"`
				Int int    `dependency:"int"`
			}

			target := targetStruct{}
			injectDependency(&target, dmap{}, req)
			So(target.Str, ShouldEqual, "string")
			So(target.Int, ShouldEqual, 1)
		})

		Convey("should inject interface", func() {
			type targetStruct struct {
				Store istore `dependency:"istore"`
			}

			target := targetStruct{}
			injectDependency(&target, dmap{}, req)
			So(target.Store, ShouldImplement, (*istore)(nil))
			So(target.Store.get(), ShouldEqual, "TestApp")
		})

		Convey("should inject struct", func() {
			type targetStruct struct {
				Store store `dependency:"store"`
			}

			target := targetStruct{}
			injectDependency(&target, dmap{}, req)
			So(target.Store, ShouldHaveSameTypeAs, store{})
			So(target.Store.get(), ShouldEqual, "TestApp")
		})

		Convey("should not inject to with field without tag", func() {
			type targetStruct struct {
				Str string `dependency:"str"`
				str string
			}

			target := targetStruct{}
			injectDependency(&target, dmap{}, req)
			So(target.Str, ShouldEqual, "string")
			So(target.str, ShouldBeEmpty)
		})

		Convey("should panic if field type is wrong", func() {
			type targetStruct struct {
				Str int `dependency:"str"`
			}

			target := targetStruct{}
			So(func() {
				injectDependency(&target, dmap{}, req)
			}, ShouldPanic)
		})

		Convey("should return error dependency name is wrong", func() {
			type targetStruct struct {
				Str int `dependency:"i_am_your_father"`
			}

			target := targetStruct{}
			err := injectDependency(&target, dmap{}, req)
			errResponse := err.(skyerr.Error)
			So(errResponse.Code(), ShouldEqual, skyerr.InvalidArgument)
		})
	})
}
