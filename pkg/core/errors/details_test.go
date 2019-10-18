package errors_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func TestDetails(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.WithDetails(err1, errors.Details{"data": 123})
	err3 := errors.HandledWithMessage(err2, "error 2")
	err := errors.WithDetails(err3, errors.Details{"data": 456, "value": errors.SafeString("test")})
	Convey("WithDetails/CollectDetails", t, func() {
		So(err, ShouldBeError, "error 2")
		details := errors.CollectDetails(err, nil)
		So(details, ShouldResemble, errors.Details{
			"data":  456,
			"value": errors.SafeString("test"),
		})
	})
	Convey("FilterDetails/GetSafeDetails", t, func() {
		details := errors.GetSafeDetails(err)
		So(details, ShouldResemble, errors.Details{
			"value": errors.SafeString("test"),
		})
	})
}
