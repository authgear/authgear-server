package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoginID(t *testing.T) {
	Convey("Test isValid", t, func() {
		Convey("validate by loginIDsKeyWhitelist: [username, email]", func() {
			keys := []string{"username", "email"}
			checker := defaultLoginIDChecker{
				loginIDsKeyWhitelist: keys,
			}

			loginID := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(checker.isValid(loginID), ShouldBeTrue)

			loginID = map[string]string{
				"username": "johndoe",
			}
			So(checker.isValid(loginID), ShouldBeTrue)

			loginID = map[string]string{
				"email": "johndoe@example.com",
			}
			So(checker.isValid(loginID), ShouldBeTrue)

			loginID = map[string]string{
				"nickname": "johndoe",
			}
			So(checker.isValid(loginID), ShouldBeFalse)

			loginID = map[string]string{
				"email": "",
			}
			So(checker.isValid(loginID), ShouldBeFalse)

			loginID = map[string]string{}
			So(checker.isValid(loginID), ShouldBeFalse)
		})
	})
}
