package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

func TestSecondary(t *testing.T) {
	Convey("WithSecondaryError", t, func() {
		secondary := errors.WithDetails(errors.New("secondary"), errors.Details{"data1": 123})
		primary := errors.WithDetails(errors.New("primary"), errors.Details{"data2": 456})
		err := errors.WithSecondaryError(primary, secondary)

		So(err, ShouldBeError, "primary")
		So(errors.Is(err, primary), ShouldBeTrue)
		So(errors.Is(err, secondary), ShouldBeFalse)

		details := errors.CollectDetails(err, nil)
		So(details, ShouldContainKey, "data1")
		So(details, ShouldContainKey, "data2")
	})
}
