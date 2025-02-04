package identity

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOAuth_Apple_MergeRawProfileAndClaims(t *testing.T) {
	Convey("OAuth.Apple_MergeRawProfileAndClaims", t, func() {
		existing := &OAuth{
			UserProfile: map[string]interface{}{
				"iss":         "https://appleid.apple.com",
				"auth_time":   1,
				"given_name":  "John",
				"family_name": "Doe",
				"email":       "johndoe@example.com",
			},
			Claims: map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
				"email":       "johndoe@example.com",
			},
		}

		Convey("Just use the incoming if it is THE FIRST TIME authorization", func() {
			rawProfile := map[string]interface{}{
				"iss":         "https://appleid.apple.com",
				"auth_time":   2,
				"given_name":  "Jane",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			}
			claims := map[string]interface{}{
				"given_name":  "Jane",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			}

			existing.Apple_MergeRawProfileAndClaims(rawProfile, claims)
			So(existing.UserProfile, ShouldResemble, map[string]interface{}{
				"iss":         "https://appleid.apple.com",
				"auth_time":   2,
				"given_name":  "Jane",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			})
			So(existing.Claims, ShouldResemble, map[string]interface{}{
				"given_name":  "Jane",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			})
		})

		Convey("Use the existing given_name and family_name", func() {
			rawProfile := map[string]interface{}{
				"iss":       "https://appleid.apple.com",
				"auth_time": 2,
				"email":     "newjohndoe@example.com",
			}
			claims := map[string]interface{}{
				"email": "newjohndoe@example.com",
			}
			existing.Apple_MergeRawProfileAndClaims(rawProfile, claims)
			So(existing.UserProfile, ShouldResemble, map[string]interface{}{
				"iss":         "https://appleid.apple.com",
				"auth_time":   2,
				"given_name":  "John",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			})
			So(existing.Claims, ShouldResemble, map[string]interface{}{
				"given_name":  "John",
				"family_name": "Doe",
				"email":       "newjohndoe@example.com",
			})
		})
	})
}
