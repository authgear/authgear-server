package webapp

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
