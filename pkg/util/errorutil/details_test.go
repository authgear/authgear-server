package errorutil_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func TestDetails(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errorutil.WithDetails(err1, errorutil.Details{"data": 123})
	err3 := fmt.Errorf("error 2: %w", err2)
	err := errorutil.WithDetails(err3, errorutil.Details{"data": 456, "value": errorutil.SafeDetail.Value("test")})
	Convey("WithDetails/CollectDetails", t, func() {
		So(err, ShouldBeError, "error 2: error 1")
		details := errorutil.CollectDetails(err, nil)
		So(details, ShouldResemble, errorutil.Details{
			"data":  456,
			"value": errorutil.SafeDetail.Value("test"),
		})
	})
	Convey("FilterDetails/GetSafeDetails", t, func() {
		details := errorutil.GetSafeDetails(err)
		So(details, ShouldResemble, errorutil.Details{
			"value": "test",
		})
	})
	Convey("DetailTaggedValue", t, func() {
		So(err, ShouldBeError, "error 2: error 1")
		details := errorutil.CollectDetails(err, nil)
		data, _ := json.Marshal(details)
		So(string(data), ShouldEqual, `{"data":456,"value":"[detail: safe]"}`)
	})

	Convey("WithDetails/CollectDetails works well with errors.Join", t, func() {
		rootCause := errors.New("root cause")
		rootCauseWithDetails := errorutil.WithDetails(rootCause, errorutil.Details{"a": "a", "b": "b"})
		invalidCredentials := errors.New("invalid credentials")
		err := errors.Join(errorutil.WithDetails(invalidCredentials, errorutil.Details{"a": "aa"}), rootCauseWithDetails)

		So(errors.Is(err, rootCause), ShouldBeTrue)
		So(errors.Is(err, rootCauseWithDetails), ShouldBeTrue)
		So(errors.Is(err, invalidCredentials), ShouldBeTrue)
		So(errors.Is(err, err), ShouldBeTrue)
		details := errorutil.CollectDetails(err, nil)
		So(details, ShouldResemble, errorutil.Details{
			"a": "aa",
			"b": "b",
		})
	})
}
