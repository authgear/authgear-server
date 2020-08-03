package verification

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func TestService(t *testing.T) {

	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identities := NewMockIdentityService(ctrl)
		authenticators := NewMockAuthenticatorService(ctrl)
		t := true
		service := &Service{
			Config: &config.VerificationConfig{
				Criteria: config.VerificationCriteriaAny,
			},
			LoginID: &config.LoginIDConfig{
				Keys: []config.LoginIDKeyConfig{{
					Key:          "email",
					Type:         "email",
					Verification: &config.VerificationLoginIDKeyConfig{Enabled: &t},
				}},
			},
			Identities:     identities,
			Authenticators: authenticators,
		}

		identityLoginID := func(loginIDKey string, loginIDValue string) *identity.Info {
			return &identity.Info{
				UserID: "user-id",
				ID:     "login-id-" + loginIDValue,
				Type:   authn.IdentityTypeLoginID,
				Claims: map[string]interface{}{
					"test-id":                          "login-id-" + loginIDValue,
					identity.IdentityClaimLoginIDKey:   loginIDKey,
					identity.IdentityClaimLoginIDValue: loginIDValue,
				},
			}
		}

		identityOfType := func(t authn.IdentityType) *identity.Info {
			return &identity.Info{
				UserID: "user-id",
				ID:     string(t),
				Type:   t,
				Claims: map[string]interface{}{
					"test-id": string(t),
				},
			}
		}

		must := func(value bool, err error) bool {
			So(err, ShouldBeNil)
			return value
		}

		Convey("IsIdentityVerifiable", func() {
			So(service.IsIdentityVerifiable(identityOfType(authn.IdentityTypeOAuth)), ShouldBeTrue)
			So(service.IsIdentityVerifiable(identityOfType(authn.IdentityTypeAnonymous)), ShouldBeFalse)
			So(service.IsIdentityVerifiable(identityLoginID("email", "foo@example.com")), ShouldBeTrue)
			So(service.IsIdentityVerifiable(identityLoginID("phone", "+85200000000")), ShouldBeFalse)
			So(service.IsIdentityVerifiable(identityLoginID("username", "bar")), ShouldBeFalse)
		})

		Convey("IsIdentityVerified", func() {
			So(must(service.IsIdentityVerified(identityOfType(authn.IdentityTypeAnonymous))), ShouldBeFalse)

			authenticators.EXPECT().List("user-id").Return(nil, nil)
			So(must(service.IsIdentityVerified(identityOfType(authn.IdentityTypeOAuth))), ShouldBeTrue)

			authenticators.EXPECT().List("user-id").Return([]*authenticator.Info{{
				ID:    "email",
				Type:  authn.AuthenticatorTypeOOB,
				Props: map[string]interface{}{authenticator.AuthenticatorPropOOBOTPEmail: "foo@example.com"},
			}}, nil)
			So(must(service.IsIdentityVerified(identityLoginID("email", "foo@example.com"))), ShouldBeTrue)

			authenticators.EXPECT().List("user-id").Return([]*authenticator.Info{{
				ID:    "phone",
				Type:  authn.AuthenticatorTypeOOB,
				Props: map[string]interface{}{authenticator.AuthenticatorPropOOBOTPPhone: "+85200000000"},
			}}, nil)
			So(must(service.IsIdentityVerified(identityLoginID("email", "foo@example.com"))), ShouldBeFalse)

			So(must(service.IsIdentityVerified(identityLoginID("phone", "+85200000000"))), ShouldBeFalse)
			So(must(service.IsIdentityVerified(identityLoginID("username", "bar"))), ShouldBeFalse)
		})

		Convey("IsVerified", func() {
			cases := []struct {
				Identities     []*identity.Info
				Authenticators []*authenticator.Info
				AnyResult      bool
				AllResult      bool
			}{
				{
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityOfType(authn.IdentityTypeAnonymous),
					},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityOfType(authn.IdentityTypeOAuth),
					},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOfType(authn.IdentityTypeOAuth),
					},
					AnyResult: true, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOfType(authn.IdentityTypeOAuth),
					},
					Authenticators: []*authenticator.Info{{
						ID:    "email",
						Type:  authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{authenticator.AuthenticatorPropOOBOTPEmail: "foo@example.com"},
					}},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("phone", "+85200000000"),
					},
					Authenticators: []*authenticator.Info{{
						ID:    "phone",
						Type:  authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{authenticator.AuthenticatorPropOOBOTPPhone: "+85200000000"},
					}},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityLoginID("phone", "+85200000000"),
					},
					Authenticators: []*authenticator.Info{{
						ID:    "email",
						Type:  authn.AuthenticatorTypeOOB,
						Props: map[string]interface{}{authenticator.AuthenticatorPropOOBOTPEmail: "foo@example.com"},
					}},
					AnyResult: true, AllResult: true,
				},
			}

			for i, c := range cases {
				Convey(fmt.Sprintf("case %d", i), func() {
					service.Config.Criteria = config.VerificationCriteriaAny
					So(service.IsVerified(c.Identities, c.Authenticators), ShouldEqual, c.AnyResult)

					service.Config.Criteria = config.VerificationCriteriaAll
					So(service.IsVerified(c.Identities, c.Authenticators), ShouldEqual, c.AllResult)
				})
			}
		})
	})
}
