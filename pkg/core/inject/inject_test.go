package inject

import (
	"context"
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

func (s dmap) Provide(name string, ctx context.Context, configuration config.TenantConfiguration) interface{} {
	switch name {
	case "str":
		return "string"
	case "int":
		return 1
	case "store":
		return store{configuration.AppName}
	case "istore":
		return &store{configuration.AppName}
	default:
		return nil
	}
}

func TestInjectDependency(t *testing.T) {
	ctx := context.Background()
	config := config.TenantConfiguration{
		AppName: "TestApp",
	}

	Convey("inject to target struct", t, func() {
		type targetStruct struct {
			Str string `dependency:"str"`
			Int int    `dependency:"int"`
		}

		target := targetStruct{}
		injectDependency(&target, dmap{}, ctx, config)
		So(target.Str, ShouldEqual, "string")
		So(target.Int, ShouldEqual, 1)
	})

	Convey("inject to target struct with interface", t, func() {
		type targetStruct struct {
			Store istore `dependency:"istore"`
		}

		target := targetStruct{}
		injectDependency(&target, dmap{}, ctx, config)
		So(target.Store, ShouldImplement, (*istore)(nil))
		So(target.Store.get(), ShouldEqual, "TestApp")
	})

	Convey("inject tot target struct with struct", t, func() {
		type targetStruct struct {
			Store store `dependency:"store"`
		}

		target := targetStruct{}
		injectDependency(&target, dmap{}, ctx, config)
		So(target.Store, ShouldHaveSameTypeAs, store{})
		So(target.Store.get(), ShouldEqual, "TestApp")
	})

	Convey("inject to target struct mixed with field without tag", t, func() {
		type targetStruct struct {
			Str string `dependency:"str"`
			str string
		}

		target := targetStruct{}
		injectDependency(&target, dmap{}, ctx, config)
		So(target.Str, ShouldEqual, "string")
		So(target.str, ShouldBeEmpty)
	})

	Convey("inject to target with wrong type", t, func() {
		type targetStruct struct {
			Str int `dependency:"str"`
		}

		target := targetStruct{}
		So(func() {
			injectDependency(&target, dmap{}, ctx, config)
		}, ShouldPanic)
	})

	Convey("inject to target with non existing dependency name", t, func() {
		type targetStruct struct {
			Str int `dependency:"i_am_your_father"`
		}

		target := targetStruct{}
		So(func() {
			injectDependency(&target, dmap{}, ctx, config)
		}, ShouldPanic)
	})
}
