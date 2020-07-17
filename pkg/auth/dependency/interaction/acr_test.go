package interaction_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/interaction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oidc"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestDeriveACR(t *testing.T) {
	Convey("DeriveACR", t, func() {
		test := func(primary *authenticator.Info, secondary *authenticator.Info, expected string) {
			actual := interaction.DeriveACR(primary, secondary)
			So(actual, ShouldEqual, expected)
		}

		// OAuth
		test(nil, nil, "")

		// password
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypePassword,
		}, nil, "")

		// OOB
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
			},
		}, nil, "")

		// OOB + OOB
		test(&authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelSMS),
			},
		}, &authenticator.Info{
			Type: authn.AuthenticatorTypeOOB,
			Props: map[string]interface{}{
				authenticator.AuthenticatorPropOOBOTPChannelType: string(authn.AuthenticatorOOBChannelEmail),
			},
		}, oidc.ACRMFA)
	})
}
