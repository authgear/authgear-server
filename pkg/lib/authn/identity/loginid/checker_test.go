package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func newLoginIDKeyConfig(key string, t config.LoginIDKeyType, maxAmount int, maxLength int) config.LoginIDKeyConfig {
	return config.LoginIDKeyConfig{
		Key:       key,
		Type:      t,
		MaxAmount: &maxAmount,
		MaxLength: &maxLength,
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
				newLoginIDKeyConfig("username", config.LoginIDKeyTypeUsername, 1, 10),
				newLoginIDKeyConfig("email", config.LoginIDKeyTypeEmail, 1, 30),
				newLoginIDKeyConfig("phone", config.LoginIDKeyTypePhone, 1, 12),
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
			var loginIDs []Spec

			loginIDs = []Spec{
				{Key: "username", Type: config.LoginIDKeyTypeUsername, Value: "johndoe"},
				{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: "johndoe@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []Spec{
				{Key: "username", Type: config.LoginIDKeyTypeUsername, Value: "johndoe"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []Spec{
				{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: "johndoe@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeNil)

			loginIDs = []Spec{
				{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: "johndoe+1@example.com"},
				{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: "johndoe+2@example.com"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n<root>: too many login IDs")

			loginIDs = []Spec{
				{Key: "nickname", Type: "", Value: "johndoe"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: login ID key is not allowed")

			loginIDs = []Spec{
				{Key: "username", Type: config.LoginIDKeyTypeUsername, Value: "foobarexample"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: maxLength\n  map[actual:13 expected:10]")

			loginIDs = []Spec{
				{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: ""},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: required")

			loginIDs = []Spec{
				{Key: "phone", Type: config.LoginIDKeyTypePhone, Value: "51234567"},
			}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n/0: format\n  map[format:phone]")

			loginIDs = []Spec{}
			So(checker.Validate(loginIDs), ShouldBeError, "invalid login IDs:\n<root>: required")
		})
	})
}
