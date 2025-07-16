package slogutil

import (
	"log/slog"
	"testing"

	"github.com/jba/slog/withsupport"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLinearizeGroupOrAttrs(t *testing.T) {
	Convey("LinearizeGroupOrAttrs", t, func() {
		var root *withsupport.GroupOrAttrs

		Convey("should linearize nil", func() {
			attrs := LinearizeGroupOrAttrs(root)
			So(len(attrs), ShouldEqual, 0)
		})

		Convey("should linearize groups and attrs", func() {
			g := root.WithGroup("a").WithAttrs([]slog.Attr{
				slog.String("b", "b"),
			}).WithGroup("c").WithAttrs([]slog.Attr{
				slog.String("d", "d"),
				slog.String("e", "e"),
			})

			attrs := LinearizeGroupOrAttrs(g)
			So(len(attrs), ShouldEqual, 3)
			So(attrs[0].String(), ShouldEqual, "a=[b=b]")
			So(attrs[1].String(), ShouldEqual, "a=[c=[d=d]]")
			So(attrs[2].String(), ShouldEqual, "a=[c=[e=e]]")
		})
	})
}
