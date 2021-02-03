package interaction

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func TestSortAuthenticators(t *testing.T) {
	info := func(typ authn.AuthenticatorType, id string) *authenticator.Info {
		return &authenticator.Info{
			ID:   id,
			Type: typ,
		}
	}

	infoDefault := func(typ authn.AuthenticatorType, id string, isDefault bool) *authenticator.Info {
		i := &authenticator.Info{
			ID:        id,
			Type:      typ,
			IsDefault: isDefault,
		}
		return i
	}

	test := func(ais []*authenticator.Info, preferred []authn.AuthenticatorType, expected []*authenticator.Info) {
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
		test([]*authenticator.Info{}, []authn.AuthenticatorType{}, []*authenticator.Info{})

		// Sort singleton
		test([]*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password"),
		}, []authn.AuthenticatorType{}, []*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password"),
		})

		// OTP comes before
		test([]*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password"),
			info(authn.AuthenticatorTypeOOBEmail, "oob"),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			info(authn.AuthenticatorTypeOOBEmail, "oob"),
			info(authn.AuthenticatorTypePassword, "password"),
		})

		// Sort is stable
		test([]*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
			info(authn.AuthenticatorTypeOOBEmail, "oob1"),
			info(authn.AuthenticatorTypeOOBEmail, "oob2"),
			info(authn.AuthenticatorTypeOOBSMS, "oob_sms"),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			info(authn.AuthenticatorTypeOOBEmail, "oob1"),
			info(authn.AuthenticatorTypeOOBEmail, "oob2"),
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
			info(authn.AuthenticatorTypeOOBSMS, "oob_sms"),
		})
	})

	Convey("SortAuthenticators by default", t, func() {
		// Sort singleton
		test([]*authenticator.Info{
			infoDefault(authn.AuthenticatorTypePassword, "password", true),
		}, []authn.AuthenticatorType{}, []*authenticator.Info{
			infoDefault(authn.AuthenticatorTypePassword, "password", true),
		})

		// Default comes first
		test([]*authenticator.Info{
			infoDefault(authn.AuthenticatorTypePassword, "password", true),
			info(authn.AuthenticatorTypeOOBEmail, "oob"),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			infoDefault(authn.AuthenticatorTypePassword, "password", true),
			info(authn.AuthenticatorTypeOOBEmail, "oob"),
		})

		test([]*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
			info(authn.AuthenticatorTypeOOBEmail, "oob1"),
			infoDefault(authn.AuthenticatorTypeOOBEmail, "oob2", true),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOBEmail,
		}, []*authenticator.Info{
			infoDefault(authn.AuthenticatorTypeOOBEmail, "oob2", true),
			info(authn.AuthenticatorTypeOOBEmail, "oob1"),
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
		})
	})
}
