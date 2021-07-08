package interaction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

type inputA struct{}

func (a inputA) X() string { return "X" }

type inputB struct{ A inputA }

func (b inputB) Y() string          { return "Y" }
func (b inputB) Input() interface{} { return b.A }

func TestInput(t *testing.T) {
	Convey("Input", t, func() {
		var x interface{ X() string }
		var y interface{ Y() string }

		a := inputA{}
		So(interaction.AsInput(a, &x), ShouldBeTrue)
		So(x.X(), ShouldEqual, "X")
		So(interaction.AsInput(a, &y), ShouldBeFalse)

		b := inputB{}
		So(interaction.AsInput(b, &x), ShouldBeTrue)
		So(x.X(), ShouldEqual, "X")
		So(interaction.AsInput(b, &y), ShouldBeTrue)
		So(y.Y(), ShouldEqual, "Y")
	})

	Convey("Input incompatible nil", t, func() {
		var x interface{ X() string }
		var a interface{} = nil
		So(interaction.AsInput(a, &x), ShouldBeFalse)
	})

	Convey("Input compatible nil", t, func() {
		var x interface{ X() string }
		var a interface{ X() string } = nil
		So(interaction.AsInput(a, &x), ShouldBeFalse)
	})
}
