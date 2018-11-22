package password

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSingupHandler(t *testing.T) {
	Convey("Test ToValidAuthDataList with multiple keys", t, func() {
		Convey("should generate authData list according to keys", func() {
			keys := [][]string{[]string{"username"}, []string{"email"}}
			authData := map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(ToValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
				},
				map[string]interface{}{
					"email": "johndoe@example.com",
				},
			})

			keys = [][]string{[]string{"username", "email"}, []string{"username", "phone"}}
			authData = map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(ToValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"username": "johndoe",
					"email":    "johndoe@example.com",
				},
			})

			authData = map[string]interface{}{
				"username": "johndoe",
			}
			So(ToValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{})

			keys = [][]string{[]string{"username", "email"}, []string{"email"}}
			authData = map[string]interface{}{
				"username": "johndoe",
				"email":    "johndoe@example.com",
			}
			So(ToValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
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
			So(ToValidAuthDataList(keys, authData), ShouldResemble, []map[string]interface{}{
				map[string]interface{}{
					"nickname": "johndoe",
				},
			})
		})
	})
}
