package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func newLoginIDKeyConfig(key string, t config.LoginIDKeyType, maxLength int) config.LoginIDKeyConfig {
	return config.LoginIDKeyConfig{
		Key:       key,
		Type:      t,
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
			CaseSensitive:                 newFalse(),
			BlockPlusSign:                 newFalse(),
			IgnoreDotSign:                 newFalse(),
			DomainBlocklistEnabled:        newFalse(),
			DomainAllowlistEnabled:        newFalse(),
			BlockFreeEmailProviderDomains: newFalse(),
		},
		Username: &config.LoginIDUsernameConfig{
			BlockReservedUsernames: newFalse(),
			ExcludeKeywordsEnabled: newFalse(),
			ASCIIOnly:              newFalse(),
			CaseSensitive:          newFalse(),
		},
	}

}

func TestLoginIDChecker(t *testing.T) {
	Convey("LoginIDChecker.Validate", t, func() {
		Convey("Validate by config: username (0-1), email (0-1)", func() {
			keysConfig := []config.LoginIDKeyConfig{
				newLoginIDKeyConfig("username", config.LoginIDKeyTypeUsername, 10),
				newLoginIDKeyConfig("email", config.LoginIDKeyTypeEmail, 30),
				newLoginIDKeyConfig("phone", config.LoginIDKeyTypePhone, 12),
			}
			typesConfig := newLoginIDTypesConfig()
			cfg := &config.LoginIDConfig{
				Types: typesConfig,
				Keys:  keysConfig,
			}
			checker := &Checker{
				Config: cfg,
				TypeCheckerFactory: &TypeCheckerFactory{
					Config:    cfg,
					Resources: resource.NewManager(resource.DefaultRegistry, nil),
				},
			}
			options := CheckerOptions{
				EmailByPassBlocklistAllowlist: false,
			}
			var loginID Spec

			loginID = Spec{Key: "username", Type: config.LoginIDKeyTypeUsername, Value: "johndoe"}

			So(checker.ValidateOne(loginID, options), ShouldBeNil)
			loginID = Spec{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: "johndoe@example.com"}
			So(checker.ValidateOne(loginID, options), ShouldBeNil)

			loginID = Spec{Key: "nickname", Type: "", Value: "johndoe"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n<root>: login ID key is not allowed")

			loginID = Spec{Key: "username", Type: config.LoginIDKeyTypeUsername, Value: "foobarexample"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n<root>: maxLength\n  map[actual:13 expected:10]")

			loginID = Spec{Key: "email", Type: config.LoginIDKeyTypeEmail, Value: ""}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n<root>: required")

			loginID = Spec{Key: "phone", Type: config.LoginIDKeyTypePhone, Value: "51234567"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n<root>: format\n  map[format:phone]")
		})
	})
}
