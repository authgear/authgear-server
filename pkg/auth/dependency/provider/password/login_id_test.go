package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoginID(t *testing.T) {
	Convey("Test ParseLoginIDs", t, func() {
		Convey("parse raw login ID list", func() {
			var loginIDs []LoginID

			loginIDs = ParseLoginIDs([]map[string]string{
				map[string]string{"username": "johndoe"},
			})
			So(loginIDs, ShouldResemble, []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
			})

			loginIDs = ParseLoginIDs([]map[string]string{
				map[string]string{"username": "johndoe"},
				map[string]string{"email": "johndoe@example.com"},
			})
			So(loginIDs, ShouldResemble, []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "email", Value: "johndoe@example.com"},
			})

			loginIDs = ParseLoginIDs([]map[string]string{
				map[string]string{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
				map[string]string{"phone": "+85299999999"},
			})
			So(loginIDs, ShouldResemble, []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "email", Value: "johndoe@example.com"},
				LoginID{Key: "phone", Value: "+85299999999"},
			})

			loginIDs = ParseLoginIDs([]map[string]string{})
			So(loginIDs, ShouldResemble, []LoginID{})
		})
	})

	Convey("Test isValid", t, func() {
		Convey("validate by loginIDsKeyWhitelist: [username, email]", func() {
			keys := []string{"username", "email"}
			checker := defaultLoginIDChecker{
				loginIDsKeyWhitelist: keys,
			}
			var loginIDs []LoginID

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.isValid(loginIDs), ShouldBeTrue)

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
			}
			So(checker.isValid(loginIDs), ShouldBeTrue)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.isValid(loginIDs), ShouldBeTrue)

			loginIDs = []LoginID{
				LoginID{Key: "nickname", Value: "johndoe"},
			}
			So(checker.isValid(loginIDs), ShouldBeFalse)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: ""},
			}
			So(checker.isValid(loginIDs), ShouldBeFalse)

			loginIDs = []LoginID{}
			So(checker.isValid(loginIDs), ShouldBeFalse)
		})
	})
}
