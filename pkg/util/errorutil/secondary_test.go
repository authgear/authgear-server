package errorutil_test

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestSecondary(t *testing.T) {
	Convey("WithSecondaryError", t, func() {
		secondary := errorutil.WithDetails(errors.New("secondary"), errorutil.Details{"data1": 123})
		primary := errorutil.WithDetails(errors.New("primary"), errorutil.Details{"data2": 456})
		err := errorutil.WithSecondaryError(primary, secondary)

		So(err, ShouldBeError, "primary")
		So(errorutil.Is(err, primary), ShouldBeTrue)
		So(errorutil.Is(err, secondary), ShouldBeFalse)

		details := errorutil.CollectDetails(err, nil)
		So(details, ShouldContainKey, "data1")
		So(details, ShouldContainKey, "data2")
	})
}
