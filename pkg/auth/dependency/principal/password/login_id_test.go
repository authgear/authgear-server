package password

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/validation"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"

	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func newLoginIDKeyConfig(key string, t config.LoginIDKeyType, max int) config.LoginIDKeyConfiguration {
	return config.LoginIDKeyConfiguration{
		Key:     key,
		Type:    t,
		Maximum: &max,
	}
}

func newLoginIDTypesConfig() *config.LoginIDTypesConfiguration {
	newFalse := func() *bool {
		t := false
		return &t
	}
	return &config.LoginIDTypesConfiguration{
		Email: &config.LoginIDTypeEmailConfiguration{
			CaseSensitive: newFalse(),
			BlockPlusSign: newFalse(),
			IgnoreDotSign: newFalse(),
		},
		Username: &config.LoginIDTypeUsernameConfiguration{
			BlockReservedUsernames: newFalse(),
			ExcludedKeywords:       []string{},
			ASCIIOnly:              newFalse(),
			CaseSensitive:          newFalse(),
		},
	}

}

func TestLoginID(t *testing.T) {
	Convey("Test isValid", t, func() {
		Convey("validate by config: username (0-1), email (0-1)", func() {
			reversedNameChecker, _ := NewReservedNameChecker("../../../../../reserved_name.txt")
			checker := newDefaultLoginIDChecker(
				[]config.LoginIDKeyConfiguration{
					newLoginIDKeyConfig("username", config.LoginIDKeyTypeRaw, 1),
					newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 1),
					newLoginIDKeyConfig("phone", config.LoginIDKeyType(metadata.Phone), 1),
				},
				newLoginIDTypesConfig(),
				reversedNameChecker,
			)
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
			So(validation.ErrorCauses(checker.validate(loginIDs)), ShouldResemble, []validation.ErrorCause{{
				Kind:    validation.ErrorGeneral,
				Pointer: "/0/key",
				Message: "login ID key is not allowed",
			}})

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
