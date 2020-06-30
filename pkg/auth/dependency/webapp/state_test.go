package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

func TestStateRestore(t *testing.T) {
	Convey("StateRestore", t, func() {
		test := func(rForm string, stateForm string, expected string) {
			s := State{Form: stateForm}
			form, err := url.ParseQuery(rForm)
			So(err, ShouldBeNil)
			err = s.Restore(form)
			So(err, ShouldBeNil)
			So(form.Encode(), ShouldEqual, expected)
		}

		test("a=b", "b=c", "a=b&b=c")
		test("a=b", "a=1&b=c", "a=b&b=c")
	})
}

func TestStateClearErrorIfFormChanges(t *testing.T) {
	Convey("StateClearErrorIfFormChanges", t, func() {
		test := func(rForm string, stateForm string, expected bool) {
			s := State{Form: stateForm, Error: &skyerr.APIError{}}
			form, err := url.ParseQuery(rForm)
			So(err, ShouldBeNil)
			change, err := s.ClearErrorIfFormChanges(form)
			So(err, ShouldBeNil)
			So(change, ShouldEqual, expected)
		}

		// Does not clear if the forms are equal.
		test("a=b", "a=b", false)

		// Does not clear if state form has something more.
		test("a=b", "a=b&c=d", false)

		// clear if some param is not equal.
		test("a=42", "a=b", true)
	})
}
