package loginid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func newLoginIDKeyConfig(key string, t model.LoginIDKeyType, maxLength int) config.LoginIDKeyConfig {
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
				newLoginIDKeyConfig("username", model.LoginIDKeyTypeUsername, 10),
				newLoginIDKeyConfig("email", model.LoginIDKeyTypeEmail, 30),
				newLoginIDKeyConfig("phone", model.LoginIDKeyTypePhone, 12),
			}
			typesConfig := newLoginIDTypesConfig()
			loginIDConfig := &config.LoginIDConfig{
				Types: typesConfig,
				Keys:  keysConfig,
			}
			uiConfig := &config.UIConfig{}
			checker := &Checker{
				Config: loginIDConfig,
				TypeCheckerFactory: &TypeCheckerFactory{
					LoginIDConfig: loginIDConfig,
					UIConfig:      uiConfig,
					Resources:     resource.NewManager(resource.DefaultRegistry, nil),
				},
			}
			options := CheckerOptions{
				EmailByPassBlocklistAllowlist: false,
			}
			var loginID identity.LoginIDSpec

			loginID = identity.LoginIDSpec{Key: "username", Type: model.LoginIDKeyTypeUsername, Value: "johndoe"}

			So(checker.ValidateOne(loginID, options), ShouldBeNil)
			loginID = identity.LoginIDSpec{Key: "email", Type: model.LoginIDKeyTypeEmail, Value: "johndoe@example.com"}
			So(checker.ValidateOne(loginID, options), ShouldBeNil)

			loginID = identity.LoginIDSpec{Key: "nickname", Type: "", Value: "johndoe"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n/login_id: login ID key is not allowed")

			loginID = identity.LoginIDSpec{Key: "username", Type: model.LoginIDKeyTypeUsername, Value: "foobarexample"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n/login_id: maxLength\n  map[actual:13 expected:10]")

			loginID = identity.LoginIDSpec{Key: "email", Type: model.LoginIDKeyTypeEmail, Value: ""}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n/login_id: required")

			loginID = identity.LoginIDSpec{Key: "phone", Type: model.LoginIDKeyTypePhone, Value: "51234567"}
			So(checker.ValidateOne(loginID, options), ShouldBeError, "invalid login ID:\n/login_id: format\n  map[format:phone]")
		})
	})
}
