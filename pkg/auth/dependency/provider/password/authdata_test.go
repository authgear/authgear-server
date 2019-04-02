package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLoginID(t *testing.T) {
	Convey("Test toValidLoginIDMap with different keys", t, func() {
		Convey("should generate loginID map by keys: [username, email]", func() {
			keys := []string{"username", "email"}

			loginID := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			})

			loginID = map[string]string{
				"username": "johndoe",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{
				"username": "johndoe",
			})

			loginID = map[string]string{
				"email": "johndoe@example.com",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{
				"email": "johndoe@example.com",
			})

			loginID = map[string]string{
				"nickname": "johndoe",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{})
		})
	})

	Convey("Test defaultLoginIDChecker isMatching", t, func() {
		Convey("should match is loginID exactly match [username, email]", func() {
			loginIDsKeyWhitelist := []string{"username", "email"}
			loginIDChecker := defaultLoginIDChecker{
				loginIDsKeyWhitelist: loginIDsKeyWhitelist,
			}

			loginID := map[string]string{
				"username": "mock_username",
			}
			So(loginIDChecker.isMatching(loginID), ShouldBeTrue)
			loginID = map[string]string{
				"email": "mock_email@example.com",
			}
			So(loginIDChecker.isMatching(loginID), ShouldBeTrue)
			loginID = map[string]string{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(loginIDChecker.isMatching(loginID), ShouldBeFalse)
			loginID = map[string]string{
				"role":  "mock_role",
				"email": "mock_email@example.com",
			}
			So(loginIDChecker.isMatching(loginID), ShouldBeFalse)
		})

		Convey("shouldn't match zero value", func() {
			keys := []string{"username", "email"}
			loginID := map[string]string{
				"username": "",
				"email":    "",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{})
			loginID = map[string]string{
				"username": "user",
				"email":    "",
			}
			So(toValidLoginIDMap(keys, loginID), ShouldResemble, map[string]string{})
		})
	})
}
