package errorutil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestErrorf(t *testing.T) {
	Convey("Wrap", t, func() {
		inner := errorutil.New("inner")
		err := errorutil.Wrap(inner, "test")
		So(err, ShouldBeError, "test")
		So(errorutil.Is(err, inner), ShouldBeTrue)
	})
	Convey("Wrapf", t, func() {
		inner := errorutil.New("inner")
		err := errorutil.Wrapf(inner, "err %d", 1)
		So(err, ShouldBeError, "err 1")
		So(errorutil.Is(err, inner), ShouldBeTrue)
	})
}
