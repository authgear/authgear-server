package interaction

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func TestSortAuthenticators(t *testing.T) {
	info := func(typ model.AuthenticatorType, id string) *authenticator.Info {
		return &authenticator.Info{
			ID:   id,
			Type: typ,
		}
	}

	infoDefault := func(typ model.AuthenticatorType, id string, isDefault bool) *authenticator.Info {
		i := &authenticator.Info{
			ID:        id,
			Type:      typ,
			IsDefault: isDefault,
		}
		return i
	}

	test := func(ais []*authenticator.Info, preferred []model.AuthenticatorType, expected []*authenticator.Info) {
		actual := make([]*authenticator.Info, len(ais))
		copy(actual, ais)
		SortAuthenticators(preferred, actual, func(i int) SortableAuthenticator {
			a := SortableAuthenticatorInfo(*actual[i])
			return &a
		})

		So(actual, ShouldResemble, expected)
	}

	Convey("SortAuthenticators by type", t, func() {
		// Sort nil
		test(nil, nil, []*authenticator.Info{})

		// Sort empty
		test([]*authenticator.Info{}, []model.AuthenticatorType{}, []*authenticator.Info{})

		// Sort singleton
		test([]*authenticator.Info{
			info(model.AuthenticatorTypePassword, "password"),
		}, []model.AuthenticatorType{}, []*authenticator.Info{
			info(model.AuthenticatorTypePassword, "password"),
		})

		// OTP comes before
		test([]*authenticator.Info{
			info(model.AuthenticatorTypePassword, "password"),
			info(model.AuthenticatorTypeOOBEmail, "oob"),
		}, []model.AuthenticatorType{
			model.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			info(model.AuthenticatorTypeOOBEmail, "oob"),
			info(model.AuthenticatorTypePassword, "password"),
		})

		// Sort is stable
		test([]*authenticator.Info{
			info(model.AuthenticatorTypePassword, "password1"),
			info(model.AuthenticatorTypePassword, "password2"),
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			info(model.AuthenticatorTypeOOBEmail, "oob2"),
			info(model.AuthenticatorTypeOOBSMS, "oob_sms"),
		}, []model.AuthenticatorType{
			model.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			info(model.AuthenticatorTypeOOBEmail, "oob2"),
			info(model.AuthenticatorTypePassword, "password1"),
			info(model.AuthenticatorTypePassword, "password2"),
			info(model.AuthenticatorTypeOOBSMS, "oob_sms"),
		})
	})

	Convey("SortAuthenticators by default", t, func() {
		// Sort singleton
		test([]*authenticator.Info{
			infoDefault(model.AuthenticatorTypePassword, "password", true),
		}, []model.AuthenticatorType{}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypePassword, "password", true),
		})

		// Default comes first
		test([]*authenticator.Info{
			infoDefault(model.AuthenticatorTypePassword, "password", true),
			info(model.AuthenticatorTypeOOBEmail, "oob"),
		}, []model.AuthenticatorType{
			model.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypePassword, "password", true),
			info(model.AuthenticatorTypeOOBEmail, "oob"),
		})

		test([]*authenticator.Info{
			info(model.AuthenticatorTypePassword, "password1"),
			info(model.AuthenticatorTypePassword, "password2"),
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			infoDefault(model.AuthenticatorTypeOOBEmail, "oob2", true),
		}, []model.AuthenticatorType{
			model.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypeOOBEmail, "oob2", true),
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			info(model.AuthenticatorTypePassword, "password1"),
			info(model.AuthenticatorTypePassword, "password2"),
		})
	})

	Convey("SortAuthenticators by passkey", t, func() {
		test([]*authenticator.Info{
			infoDefault(model.AuthenticatorTypePasskey, "passkey", true),
		}, []model.AuthenticatorType{}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypePasskey, "passkey", true),
		})

		// non-passkey come BEFORE default
		test([]*authenticator.Info{
			info(model.AuthenticatorTypePasskey, "passkey"),
			infoDefault(model.AuthenticatorTypePassword, "password", true),
		}, []model.AuthenticatorType{}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypePassword, "password", true),
			info(model.AuthenticatorTypePasskey, "passkey"),
		})

		// non-passkey come BEFORE higher rank
		test([]*authenticator.Info{
			info(model.AuthenticatorTypePasskey, "passkey"),
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			infoDefault(model.AuthenticatorTypeOOBEmail, "oob2", true),
		}, []model.AuthenticatorType{
			model.AuthenticatorTypePasskey,
		}, []*authenticator.Info{
			infoDefault(model.AuthenticatorTypeOOBEmail, "oob2", true),
			info(model.AuthenticatorTypeOOBEmail, "oob1"),
			info(model.AuthenticatorTypePasskey, "passkey"),
		})
	})
}
