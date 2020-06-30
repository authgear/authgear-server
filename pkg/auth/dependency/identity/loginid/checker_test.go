package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/core/auth/metadata"
)

func newLoginIDKeyConfig(key string, t config.LoginIDKeyType, max int) config.LoginIDKeyConfig {
	return config.LoginIDKeyConfig{
		Key:     key,
		Type:    t,
		Maximum: &max,
	}
}

func newLoginIDTypesConfig() *config.LoginIDTypesConfig {
	newFalse := func() *bool {
		t := false
		return &t
	}
	return &config.LoginIDTypesConfig{
		Email: &config.LoginIDEmailConfig{
			CaseSensitive: newFalse(),
			BlockPlusSign: newFalse(),
			IgnoreDotSign: newFalse(),
		},
		Username: &config.LoginIDUsernameConfig{
			BlockReservedUsernames: newFalse(),
			ExcludedKeywords:       []string{},
			ASCIIOnly:              newFalse(),
			CaseSensitive:          newFalse(),
		},
	}

}

func TestLoginIDChecker(t *testing.T) {
	Convey("LoginIDChecker.Validate", t, func() {
		Convey("Validate by config: username (0-1), email (0-1)", func() {
			reservedNameChecker, _ := NewReservedNameChecker("../../../../../reserved_name.txt")
			keysConfig := []config.LoginIDKeyConfig{
				newLoginIDKeyConfig("username", config.LoginIDKeyTypeRaw, 1),
				newLoginIDKeyConfig("email", config.LoginIDKeyType(metadata.Email), 1),
				newLoginIDKeyConfig("phone", config.LoginIDKeyType(metadata.Phone), 1),
			}
			typesConfig := newLoginIDTypesConfig()
			cfg := &config.LoginIDConfig{
				Types: typesConfig,
				Keys:  keysConfig,
			}
			checker := &Checker{
				Config: cfg,
				TypeCheckerFactory: &TypeCheckerFactory{
					Config:              cfg,
					ReservedNameChecker: reservedNameChecker,
				},
			}
			var loginIDs []LoginID

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "username", Value: "johndoe"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: "johndoe@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: "johndoe+1@example.com"},
				LoginID{Key: "email", Value: "johndoe+2@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n<root>: too many login IDs")

			loginIDs = []LoginID{
				LoginID{Key: "nickname", Value: "johndoe"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: login ID key is not allowed")

			loginIDs = []LoginID{
				LoginID{Key: "email", Value: ""},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: required")

			loginIDs = []LoginID{
				LoginID{Key: "phone", Value: "51234567"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: format\n  map[format:phone]")

			loginIDs = []LoginID{}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n<root>: required")
		})
	})
}
