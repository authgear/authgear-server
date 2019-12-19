package inject

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInjectContext(t *testing.T) {
	Convey("resolving dependencies", t, func() {
		type Dep struct{ int }

		ctxRoot := WithInject(context.Background())
		ctxRequest := WithInject(ctxRoot)
		depFactory := func() interface{} { return &Dep{} }

		Convey("resolve singleton dependencies", func() {
			depRoot := Singleton(ctxRoot, "dep", depFactory)
			depRequest := Singleton(ctxRequest, "dep", depFactory)

			dep1 := depRoot()
			So(dep1, ShouldHaveSameTypeAs, &Dep{})
			dep2 := depRequest()
			So(dep2, ShouldEqual, dep1)
		})

		Convey("resolve scoped dependencies", func() {
			depRoot := Scoped(ctxRoot, "dep", depFactory)
			depRequest := Scoped(ctxRequest, "dep", depFactory)

			dep1 := depRoot()
			dep2 := depRoot()
			So(dep1, ShouldHaveSameTypeAs, &Dep{})
			So(dep2, ShouldEqual, dep1)
			dep3 := depRequest()
			dep4 := depRequest()
			So(dep3, ShouldHaveSameTypeAs, &Dep{})
			So(dep4, ShouldEqual, dep3)
			So(dep3, ShouldNotEqual, dep1)
		})

		Convey("resolve transient dependencies", func() {
			depRoot := Transient(ctxRoot, "dep", depFactory)
			depRequest := Transient(ctxRequest, "dep", depFactory)

			dep1 := depRoot()
			dep2 := depRoot()
			So(dep1, ShouldHaveSameTypeAs, &Dep{})
			So(dep2, ShouldNotEqual, dep1)
			dep3 := depRequest()
			dep4 := depRequest()
			So(dep3, ShouldHaveSameTypeAs, &Dep{})
			So(dep4, ShouldNotEqual, dep3)
			So(dep3, ShouldNotEqual, dep1)
		})
	})
}
