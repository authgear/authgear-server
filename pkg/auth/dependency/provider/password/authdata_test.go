package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthData(t *testing.T) {
	Convey("Test toValidAuthDataMap with different keys", t, func() {
		Convey("should generate authData map by keys: [username, email]", func() {
			keys := []string{"username", "email"}

			authData := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			})

			authData = map[string]string{
				"username": "johndoe",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{
				"username": "johndoe",
			})

			authData = map[string]string{
				"email": "johndoe@example.com",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{
				"email": "johndoe@example.com",
			})

			authData = map[string]string{
				"nickname": "johndoe",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{})
		})
	})

	Convey("Test defaultAuthDataChecker isMatching", t, func() {
		Convey("should match is authData exactly match [username, email]", func() {
			loginIDsKeyWhitelist := []string{"username", "email"}
			authDataChecker := defaultAuthDataChecker{
				loginIDsKeyWhitelist: loginIDsKeyWhitelist,
			}

			authData := map[string]string{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]string{
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]string{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
			authData = map[string]string{
				"role":  "mock_role",
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
		})

		Convey("shouldn't match zero value", func() {
			keys := []string{"username", "email"}
			authData := map[string]string{
				"username": "",
				"email":    "",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{})
			authData = map[string]string{
				"username": "user",
				"email":    "",
			}
			So(toValidAuthDataMap(keys, authData), ShouldResemble, map[string]string{})
		})
	})
}
