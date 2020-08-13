package errorutil_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestSummary(t *testing.T) {
	Convey("Summary", t, func() {
		err1 := errors.New("err a")
		err2 := fmt.Errorf("err b: %w", err1)
		err3 := errorutil.Wrap(err2, "err c")
		err4 := errorutil.HandledWithMessage(err2, "err d")
		err5 := errorutil.WithSecondaryError(errors.New("err e"), err2)

		So(errorutil.Summary(err1), ShouldEqual, "err a")
		So(errorutil.Summary(err2), ShouldEqual, "err b: err a")
		So(errorutil.Summary(err3), ShouldEqual, "err c: err b: err a")
		So(errorutil.Summary(err4), ShouldEqual, "err d: err b: err a")
		So(errorutil.Summary(err5), ShouldEqual, "(err b: err a) err e")
	})
}
