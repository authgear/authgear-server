package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAuthData(t *testing.T) {
	Convey("Test toValidAuthDataList with different keys", t, func() {
		Convey("should generate authData list by keys: [[username], [email]]", func() {
			keys := [][]string{[]string{"username"}, []string{"email"}}

			authData := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
				},
				map[string]string{
					"email": "johndoe@example.com",
				},
			})

			authData = map[string]string{
				"username": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
				},
			})

			authData = map[string]string{
				"email": "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"email": "johndoe@example.com",
				},
			})

			authData = map[string]string{
				"nickname": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{})
		})

		Convey("should generate authData list by keys: [[username, email], [username, phone]]", func() {
			keys := [][]string{[]string{"username", "email"}, []string{"username", "phone"}}

			authData := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
			})

			authData = map[string]string{
				"username": "johndoe",
				"phone":    "123456",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
					"phone":    "123456",
				},
			})

			authData = map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
				"phone":    "123456",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
				map[string]string{
					"username": "johndoe",
					"phone":    "123456",
				},
			})

			authData = map[string]string{
				"username": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{})
		})

		Convey("should generate authData list by keys: [[username, email], [email]]", func() {
			keys := [][]string{[]string{"username", "email"}, []string{"email"}}

			authData := map[string]string{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
				map[string]string{
					"email": "johndoe@example.com",
				},
			})

			keys = [][]string{[]string{"username", "email"}, []string{"nickname"}}
			authData = map[string]string{
				"username": "johndoe",
				"nickname": "johndoe",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
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
		})

		Convey("should match is authData exactly match [\"username\", \"email\"]]", func() {
			authRecordKeys := [][]string{
				[]string{"username", "email"},
			}
			authDataChecker := defaultAuthDataChecker{
				authRecordKeys: authRecordKeys,
			}

			authData := map[string]string{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]string{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
			authData = map[string]string{
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

			authData := map[string]string{
				"username": "mock_username",
				"email":    "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
			authData = map[string]string{
				"username": "mock_username",
			}
			So(authDataChecker.isMatching(authData), ShouldBeFalse)
			authData = map[string]string{
				"email": "mock_email@example.com",
			}
			So(authDataChecker.isMatching(authData), ShouldBeTrue)
		})

		Convey("shouldn't match zero value", func() {
			keys := [][]string{
				[]string{"username"},
				[]string{"email"},
			}
			authData := map[string]string{
				"username": "",
				"email":    "",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{})
			authData = map[string]string{
				"username": "user",
				"email":    "",
			}
			So(toValidAuthDataList(keys, authData), ShouldResemble, []map[string]string{
				map[string]string{
					"username": "user",
				},
			})
		})
	})
}
