package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errors"
)

func TestBarrier(t *testing.T) {
	Convey("Handled/HandledWithMessage", t, func() {
		inner := errors.New("error")

		err1 := errors.Handled(inner)
		So(err1, ShouldBeError, "error")
		So(errors.Unwrap(err1), ShouldBeNil)

		err2 := errors.HandledWithMessage(inner, "test")
		So(err2, ShouldBeError, "test")
		So(errors.Unwrap(err2), ShouldBeNil)
	})
}
