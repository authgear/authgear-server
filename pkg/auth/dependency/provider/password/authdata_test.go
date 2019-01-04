package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthData(t *testing.T) {
	Convey("Test toValidAuthDataList with different keys", t, func() {
		Convey("should generate authData list by keys: [[username], [email]]", func() {
			keys := [][]string{[]string{"username"}, []string{"email"}}

			authData := map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
				},
				map[string]interface{}{
					"email": "johndoe@example.com",
				},
			})

			authData = map[string]interface{}{
				"username": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
				},
			})

			authData = map[string]interface{}{
				"email": "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"email": "johndoe@example.com",
				},
			})

			authData = map[string]interface{}{
				"nickname": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{})
		})

		Convey("should generate authData list by keys: [[username, email], [username, phone]]", func() {
			keys := [][]string{[]string{"username", "email"}, []string{"username", "phone"}}

			authData := map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
			})

			authData = map[string]interface{}{
				"username": "johndoe",
				"phone":    "123456",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
					"phone":    "123456",
				},
			})

			authData = map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
				"phone":    "123456",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
				map[string]interface{}{
					"username": "johndoe",
					"phone":    "123456",
				},
			})

			authData = map[string]interface{}{
				"username": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{})
		})

		Convey("should generate authData list by keys: [[username, email], [email]]", func() {
			keys := [][]string{[]string{"username", "email"}, []string{"email"}}

			authData := map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
				map[string]interface{}{
					"email": "johndoe@example.com",
				},
			})

			keys = [][]string{[]string{"username", "email"}, []string{"nickname"}}
			authData = map[string]interface{}{
				"username": "johndoe",
				"nickname": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"nickname": "johndoe",
				},
			})
		})
	})

	Convey("Test defaultAuthDataChecker isMatching", t, func() {
		Convey("should match is authData exactly match [\"username\"], [\"email\"]]", func() {
			authRecordKeys := [][]string{
				[]string{"username"},
				[]string{"email"},
			}
			authDataChecker := defaultAuthDataChecker{
				authRecordKeys: authRecordKeys,
			}

			authData := map[string]interface{}{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]interface{}{
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]interface{}{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
		})

		Convey("should match is authData exactly match [\"username\", \"email\"]]", func() {
			authRecordKeys := [][]string{
				[]string{"username", "email"},
			}
			authDataChecker := defaultAuthDataChecker{
				authRecordKeys: authRecordKeys,
			}

			authData := map[string]interface{}{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]interface{}{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
			authData = map[string]interface{}{
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
		})

		Convey("should match is authData exactly match [\"username\", \"email\"], [\"email\"]]", func() {
			authRecordKeys := [][]string{
				[]string{"username", "email"},
				[]string{"email"},
			}
			authDataChecker := defaultAuthDataChecker{
				authRecordKeys: authRecordKeys,
			}

			authData := map[string]interface{}{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]interface{}{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
			authData = map[string]interface{}{
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
		})

		Convey("shouldn't match zero value", func() {
			keys := [][]string{
				[]string{"username"},
				[]string{"email"},
			}
			authData := map[string]interface{}{
				"username": "",
				"email":    "",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{})
			authData = map[string]interface{}{
				"username": "user",
				"email":    "",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "user",
				},
			})
		})
	})
}
