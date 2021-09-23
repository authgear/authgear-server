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

		Convey("NonIdentityAware", func() {
			So(T{
				"a":     "b",
				"name":  "John Doe",
				"email": "louischan@oursky.com",
			}.NonIdentityAware(), ShouldResemble, T{
				"name": "John Doe",
			})
		})

		Convey("MergedWith", func() {
			So(T{
				"a":    "b",
				"keep": "this",
			}.MergedWith(T{
				"a":   "c",
				"new": "key",
			}), ShouldResemble, T{
				"keep": "this",
				"new":  "key",
				"a":    "c",
			})
		})
	})
}
