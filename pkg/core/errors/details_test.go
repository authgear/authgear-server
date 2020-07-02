package errors_test

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/errors"
)

func TestDetails(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.WithDetails(err1, errors.Details{"data": 123})
	err3 := errors.HandledWithMessage(err2, "error 2")
	err := errors.WithDetails(err3, errors.Details{"data": 456, "value": errors.SafeDetail.Value("test")})
	Convey("WithDetails/CollectDetails", t, func() {
		So(err, ShouldBeError, "error 2")
		details := errors.CollectDetails(err, nil)
		So(details, ShouldResemble, errors.Details{
			"data":  456,
			"value": errors.SafeDetail.Value("test"),
		})
	})
	Convey("FilterDetails/GetSafeDetails", t, func() {
		details := errors.GetSafeDetails(err)
		So(details, ShouldResemble, errors.Details{
			"value": "test",
		})
	})
	Convey("DetailTaggedValue", t, func() {
		So(err, ShouldBeError, "error 2")
		details := errors.CollectDetails(err, nil)
		data, _ := json.Marshal(details)
		So(string(data), ShouldEqual, `{"data":456,"value":"[detail: safe]"}`)
	})
}
