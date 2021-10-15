package stdattrs

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestT(t *testing.T) {
	Convey("T", t, func() {
		Convey("WithNameCopiedToGivenName", func() {
			So(T{}.WithNameCopiedToGivenName(), ShouldResemble, T{})
			So(T{
				"name": "John",
			}.WithNameCopiedToGivenName(), ShouldResemble, T{
				"name":       "John",
				"given_name": "John",
			})
			So(T{
				"name":       "John",
				"given_name": "Jonathan",
			}.WithNameCopiedToGivenName(), ShouldResemble, T{
				"name":       "John",
				"given_name": "Jonathan",
			})
		})
	})
}
