package model

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEndUserAccountID(t *testing.T) {
	Convey("EndUserAccountID", t, func() {
		So((&User{}).EndUserAccountID(), ShouldEqual, "")
		So((&User{
			StandardAttributes: map[string]interface{}{
				"email": "user@example.com",
			},
		}).EndUserAccountID(), ShouldEqual, "user@example.com")
		So((&User{
			StandardAttributes: map[string]interface{}{
				"preferred_username": "user",
			},
		}).EndUserAccountID(), ShouldEqual, "user")
		So((&User{
			StandardAttributes: map[string]interface{}{
				"phone_number": "+85298765432",
			},
		}).EndUserAccountID(), ShouldEqual, "+85298765432")
		So((&User{
			StandardAttributes: map[string]interface{}{
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
		}).EndUserAccountID(), ShouldEqual, "user")
		So((&User{
			StandardAttributes: map[string]interface{}{
				"email":              "user@example.com",
				"preferred_username": "user",
				"phone_number":       "+85298765432",
			},
		}).EndUserAccountID(), ShouldEqual, "user@example.com")
	})
}
