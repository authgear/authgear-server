package password

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/validation"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func newLoginIDKeyConfig(t config.LoginIDKeyType, min int, max int) config.LoginIDKeyConfiguration {
	return config.LoginIDKeyConfiguration{
		Type:    t,
		Minimum: &min,
		Maximum: &max,
	}
}

func TestLoginID(t *testing.T) {
	Convey("Test isValid", t, func() {
		Convey("validate by config: username (0-1), email (0-1)", func() {
			checker := defaultLoginIDChecker{
				loginIDsKeys: map[string]config.LoginIDKeyConfiguration{
					"username": newLoginIDKeyConfig(config.LoginIDKeyTypeRaw, 0, 1),
					"email":    newLoginIDKeyConfig(config.LoginIDKeyType(metadata.Email), 0, 1),
					"phone":    newLoginIDKeyConfig(config.LoginIDKeyType(metadata.Phone), 0, 1),
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
			So(validation.ErrorCauses(checker.validate(loginIDs)), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorEntryAmount,
				Pointer: "",
				Message: "too many login IDs",
				Details: map[string]interface{}{"key": "email", "lte": 1},
			}})

			loginIDs = []LoginID{
				LoginID{Key: "nickname", Value: "johndoe"},
			}
			So(checker.validate(loginIDs), ShouldBeError, "login ID key is not allowed")

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: ""},
			}
			So(validation.ErrorCauses(checker.validate(loginIDs)), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorRequired,
				Pointer: "/0/value",
				Message: "login ID is required",
			}})

			loginIDs = []LoginID{
				LoginID{Key: "phone", Value: "51234567"},
			}
			So(validation.ErrorCauses(checker.validate(loginIDs)), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorStringFormat,
				Pointer: "/0/value",
				Message: "invalid login ID format",
				Details: map[string]interface{}{"format": "phone"},
			}})

			loginIDs = []LoginID{}
			So(validation.ErrorCauses(checker.validate(loginIDs)), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorRequired,
				Pointer: "",
				Message: "login ID is required",
			}})
		})
	})
}
