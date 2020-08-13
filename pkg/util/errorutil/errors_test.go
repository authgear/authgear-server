package errorutil_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestNew(t *testing.T) {
	Convey("New", t, func() {
		err := errorutil.New("test")
		So(err, ShouldBeError, "test")
	})
	Convey("Newf", t, func() {
		inner := errorutil.New("test")
		err := errorutil.Newf("error %d: %w", 1, inner)
		So(err, ShouldBeError, "error 1: test")
		So(errorutil.Is(err, inner), ShouldBeTrue)
	})
}
