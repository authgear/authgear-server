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
		err3 := fmt.Errorf("err c: %w", err2)
		err4 := fmt.Errorf("err d: %w", err2)
		err5 := errorutil.WithSecondaryError(errors.New("err e"), err2)

		So(errorutil.Summary(err1), ShouldEqual, "err a")
		So(errorutil.Summary(err2), ShouldEqual, "err b: err a")
		So(errorutil.Summary(err3), ShouldEqual, "err c: err b: err a")
		So(errorutil.Summary(err4), ShouldEqual, "err d: err b: err a")
		So(errorutil.Summary(err5), ShouldEqual, "(err b: err a) err e")
	})

	Convey("Summary works well with errors.Join", t, func() {
		rootCause := errors.New("root cause")
		invalidCredentials := fmt.Errorf("invalid credentials: %w", rootCause)
		rollbackError := errors.New("rollback error")
		e := errorutil.WithSecondaryError(invalidCredentials, rollbackError)
		apiError := errors.New("api error")
		err := errors.Join(apiError, e)

		So(errorutil.Summary(err), ShouldEqual, "api error: (rollback error) invalid credentials: root cause")
	})
}
