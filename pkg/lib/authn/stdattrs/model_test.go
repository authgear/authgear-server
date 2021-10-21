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

		Convey("FormattedName", func() {
			So(T{
				"name": "John Doe",
			}.FormattedName(), ShouldEqual, "John Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
			}.FormattedName(), ShouldEqual, "John Doe")
			So(T{
				"given_name":  "John",
				"middle_name": "William",
				"family_name": "Doe",
			}.FormattedName(), ShouldEqual, "John William Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedName(), ShouldEqual, "John Doe (Johnny)")
		})

		Convey("FormattedNames", func() {
			So(T{}.FormattedNames(), ShouldEqual, "")
			So(T{
				"nickname": "John",
			}.FormattedNames(), ShouldEqual, "John")
			So(T{
				"given_name":  "John",
				"middle_name": "William",
				"family_name": "Doe",
			}.FormattedNames(), ShouldEqual, "John William Doe")
			So(T{
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)")
			So(T{
				"name": "John Doe",
			}.FormattedNames(), ShouldEqual, "John Doe")
			So(T{
				"name":     "John Doe",
				"nickname": "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)")
			So(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
			}.FormattedNames(), ShouldEqual, "John Doe\nJohn Doe")
			So(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
				"nickname":    "Johnny",
			}.FormattedNames(), ShouldEqual, "John Doe (Johnny)\nJohn Doe")
		})

		Convey("EndUserAccountID", func() {
			So(T{}.EndUserAccountID(), ShouldEqual, "")
			So(T{
				"email": "user@example.com",
			}.EndUserAccountID(), ShouldEqual, "user@example.com")
			So(T{
				"preferred_username": "user",
			}.EndUserAccountID(), ShouldEqual, "user")
			So(T{
				"phone_number": "+85298765432",
			}.EndUserAccountID(), ShouldEqual, "+85298765432")
			So(T{
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			}.EndUserAccountID(), ShouldEqual, "user")
			So(T{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			}.EndUserAccountID(), ShouldEqual, "user@example.com")
		})

		Convey("Clone", func() {
			a := T{
				"address": map[string]interface{}{
					"street_address": "a",
				},
			}
			b := a.Clone()
			b["address"].(map[string]interface{})["street_address"] = "b"

			So(b, ShouldResemble, T{
				"address": map[string]interface{}{
					"street_address": "b",
				},
			})
			So(a, ShouldResemble, T{
				"address": map[string]interface{}{
					"street_address": "a",
				},
			})
		})

		Convey("Tidy", func() {
			a := T{
				"name":    "John Doe",
				"address": map[string]interface{}{},
			}
			So(a.Tidy(), ShouldResemble, T{
				"name": "John Doe",
			})
		})

		Convey("MergedWithJSONPointer", func() {
			test := func(original T, ptrs map[string]interface{}, expected T) {
				actual, err := original.MergedWithJSONPointer(ptrs)
				So(err, ShouldBeNil)
				So(actual, ShouldResemble, expected)
			}

			test(T{
				"name":        "John Doe",
				"given_name":  "John",
				"family_name": "Doe",
				"address": map[string]interface{}{
					"street_address": "Some street",
				},
			}, map[string]interface{}{
				"/given_name":             "",
				"/family_name":            "Lee",
				"/middle_name":            "William",
				"/address/street_address": "",
			}, T{
				"name":        "John Doe",
				"middle_name": "William",
				"family_name": "Lee",
			})
		})
	})
}
