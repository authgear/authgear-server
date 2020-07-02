package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

func TestNew(t *testing.T) {
	Convey("New", t, func() {
		err := errors.New("test")
		So(err, ShouldBeError, "test")
	})
	Convey("Newf", t, func() {
		inner := errors.New("test")
		err := errors.Newf("error %d: %w", 1, inner)
		So(err, ShouldBeError, "error 1: test")
		So(errors.Is(err, inner), ShouldBeTrue)
	})
}
