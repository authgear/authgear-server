package errorutil_test

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestBarrier(t *testing.T) {
	Convey("Handled/HandledWithMessage", t, func() {
		inner := errors.New("error")

		err1 := errorutil.Handled(inner)
		So(err1, ShouldBeError, "error")
		So(errorutil.Unwrap(err1), ShouldBeNil)

		err2 := errorutil.HandledWithMessage(inner, "test")
		So(err2, ShouldBeError, "test")
		So(errorutil.Unwrap(err2), ShouldBeNil)
	})
}
