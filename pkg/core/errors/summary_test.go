package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

func TestSummary(t *testing.T) {
	Convey("Summary", t, func() {
		err1 := errors.New("err a")
		err2 := errors.Newf("err b: %w", err1)
		err3 := errors.Wrap(err2, "err c")
		err4 := errors.HandledWithMessage(err2, "err d")
		err5 := errors.WithSecondaryError(errors.New("err e"), err2)

		So(errors.Summary(err1), ShouldEqual, "err a")
		So(errors.Summary(err2), ShouldEqual, "err b: err a")
		So(errors.Summary(err3), ShouldEqual, "err c: err b: err a")
		So(errors.Summary(err4), ShouldEqual, "err d: err b: err a")
		So(errors.Summary(err5), ShouldEqual, "(err b: err a) err e")
	})
}
