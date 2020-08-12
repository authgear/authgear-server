package newinteraction

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestSortAuthenticators(t *testing.T) {
	info := func(typ authn.AuthenticatorType, id string) *authenticator.Info {
		return &authenticator.Info{
			ID:   id,
			Type: typ,
		}
	}

	test := func(ais []*authenticator.Info, preferred []authn.AuthenticatorType, expected []*authenticator.Info) {
		actual := make([]*authenticator.Info, len(ais))
		copy(actual, ais)
		SortAuthenticators(preferred, actual, func(i int) authn.AuthenticatorType {
			return actual[i].Type
		})

		So(actual, ShouldResemble, expected)
	}

	Convey("SortAuthenticators", t, func() {
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
			info(authn.AuthenticatorTypeOOB, "oob"),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOB,
		}, []*authenticator.Info{
			info(authn.AuthenticatorTypeOOB, "oob"),
			info(authn.AuthenticatorTypePassword, "password"),
		})

		// Sort is stable
		test([]*authenticator.Info{
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
			info(authn.AuthenticatorTypeOOB, "oob1"),
			info(authn.AuthenticatorTypeOOB, "oob2"),
		}, []authn.AuthenticatorType{
			authn.AuthenticatorTypeOOB,
		}, []*authenticator.Info{
			info(authn.AuthenticatorTypeOOB, "oob1"),
			info(authn.AuthenticatorTypeOOB, "oob2"),
			info(authn.AuthenticatorTypePassword, "password1"),
			info(authn.AuthenticatorTypePassword, "password2"),
		})
	})
}
