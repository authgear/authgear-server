package password

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func newLoginIDKeyConfig(min int, max int) config.LoginIDKeyConfiguration {
	return config.LoginIDKeyConfiguration{
		Minimum: &min,
		Maximum: &max,
	}
}

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
				LoginID{Key: "email", Value: "johndoe@example.com"},
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "phone", Value: "+85299999999"},
			})

			loginIDs = ParseLoginIDs([]map[string]string{})
			So(loginIDs, ShouldResemble, []LoginID{})
		})
	})

	Convey("Test isValid", t, func() {
		Convey("validate by config: username (0-1), email (0-1)", func() {
			checker := defaultLoginIDChecker{
				loginIDsKeys: map[string]config.LoginIDKeyConfiguration{
					"username": newLoginIDKeyConfig(0, 1),
					"email":    newLoginIDKeyConfig(0, 1),
				},
			}
			var loginIDs []LoginID

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
			}
			So(checker.validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: "johndoe+1@example.com"},
				LoginID{Key: "email", Value: "johndoe+2@example.com"},
			}
			So(checker.validate(loginIDs), ShouldBeError, "InvalidArgument: login ID is not valid")

			loginIDs = []LoginID{
				LoginID{Key: "nickname", Value: "johndoe"},
			}
			So(checker.validate(loginIDs), ShouldBeError, "InvalidArgument: login ID key is not allowed")

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: ""},
			}
			So(checker.validate(loginIDs), ShouldBeError, "InvalidArgument: login ID is empty")

			loginIDs = []LoginID{}
			So(checker.validate(loginIDs), ShouldBeError, "InvalidArgument: no login ID is present")
		})
	})
}
