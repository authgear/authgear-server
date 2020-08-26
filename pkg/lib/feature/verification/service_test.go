package verification

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestService(t *testing.T) {

	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		identities := NewMockIdentityService(ctrl)
		authenticators := NewMockAuthenticatorService(ctrl)
		t := true
		f := false
		service := &Service{
			Config: &config.VerificationConfig{
				Claims: &config.VerificationClaimsConfig{
					Email: &config.VerificationClaimConfig{
						Enabled:  &t,
						Required: &f,
					},
					PhoneNumber: &config.VerificationClaimConfig{
						Enabled:  &t,
						Required: &t,
					},
				},
				Criteria: config.VerificationCriteriaAny,
			},
			Identities:     identities,
			Authenticators: authenticators,
		}

		identityLoginID := func(loginIDKey string, loginIDValue string) *identity.Info {
			i := &identity.Info{
				UserID: "user-id",
				ID:     "login-id-" + loginIDValue,
				Type:   authn.IdentityTypeLoginID,
				Claims: map[string]interface{}{
					"test-id":                          "login-id-" + loginIDValue,
					loginIDKey:                         loginIDValue,
					identity.IdentityClaimLoginIDKey:   loginIDKey,
					identity.IdentityClaimLoginIDType:  loginIDKey,
					identity.IdentityClaimLoginIDValue: loginIDValue,
				},
			}
			switch loginIDKey {
			case "email":
				i.Claims[identity.StandardClaimEmail] = loginIDValue
			case "phone":
				i.Claims[identity.StandardClaimPhoneNumber] = loginIDValue
			case "username":
				i.Claims[identity.StandardClaimPreferredUsername] = loginIDValue
			}
			return i
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

		must := func(value Status, err error) Status {
			So(err, ShouldBeNil)
			return value
		}

		Convey("IsIdentityVerifiable", func() {
			So(service.IsIdentityVerifiable(identityOfType(authn.IdentityTypeOAuth)), ShouldBeTrue)
			So(service.IsIdentityVerifiable(identityOfType(authn.IdentityTypeAnonymous)), ShouldBeFalse)
			So(service.IsIdentityVerifiable(identityLoginID("email", "foo@example.com")), ShouldBeTrue)
			So(service.IsIdentityVerifiable(identityLoginID("phone", "+85200000000")), ShouldBeTrue)
			So(service.IsIdentityVerifiable(identityLoginID("username", "bar")), ShouldBeFalse)
		})

		Convey("GetVerificationStatus", func() {
			// TODO: add test for infinite recursion prevention

			So(must(service.GetVerificationStatus(identityOfType(authn.IdentityTypeAnonymous))), ShouldEqual, StatusDisabled)

			identities.EXPECT().ListByUser("user-id").Return([]*identity.Info{identityOfType(authn.IdentityTypeOAuth)}, nil)
			authenticators.EXPECT().List("user-id").Return(nil, nil)
			So(must(service.GetVerificationStatus(identityOfType(authn.IdentityTypeOAuth))), ShouldEqual, StatusVerified)

			identities.EXPECT().ListByUser("user-id").Return([]*identity.Info{identityLoginID("email", "foo@example.com")}, nil)
			authenticators.EXPECT().List("user-id").Return([]*authenticator.Info{{
				ID:     "email",
				Type:   authn.AuthenticatorTypeOOB,
				Claims: map[string]interface{}{authenticator.AuthenticatorClaimOOBOTPEmail: "foo@example.com"},
			}}, nil)
			So(must(service.GetVerificationStatus(identityLoginID("email", "foo@example.com"))), ShouldEqual, StatusVerified)

			identities.EXPECT().ListByUser("user-id").Return([]*identity.Info{identityLoginID("email", "foo@example.com")}, nil)
			authenticators.EXPECT().List("user-id").Return([]*authenticator.Info{{
				ID:     "phone",
				Type:   authn.AuthenticatorTypeOOB,
				Claims: map[string]interface{}{authenticator.AuthenticatorClaimOOBOTPPhone: "+85200000000"},
			}}, nil)
			So(must(service.GetVerificationStatus(identityLoginID("email", "foo@example.com"))), ShouldEqual, StatusPending)

			So(must(service.GetVerificationStatus(identityLoginID("username", "foo"))), ShouldEqual, StatusDisabled)

			identities.EXPECT().ListByUser("user-id").Return([]*identity.Info{identityLoginID("phone", "+85200000000")}, nil)
			authenticators.EXPECT().List("user-id").Return([]*authenticator.Info{
				{ID: "email", Claims: map[string]interface{}{"test-id": "login-id-foo@example.com"}},
			}, nil)
			So(must(service.GetVerificationStatus(identityLoginID("phone", "+85200000000"))), ShouldEqual, StatusRequired)

			identities.EXPECT().ListByUser("user-id").Return([]*identity.Info{
				{
					UserID: "user-id",
					ID:     "login-id",
					Type:   authn.IdentityTypeLoginID,
					Claims: map[string]interface{}{
						identity.StandardClaimEmail: "foo@example.com",
					},
				},
				{
					UserID: "user-id",
					ID:     "oauth",
					Type:   authn.IdentityTypeOAuth,
					Claims: map[string]interface{}{
						identity.StandardClaimEmail: "foo@example.com",
					},
				},
			}, nil)
			authenticators.EXPECT().List("user-id").Return(nil, nil)
			So(must(service.GetVerificationStatus(&identity.Info{
				UserID: "user-id",
				ID:     "login-id",
				Type:   authn.IdentityTypeLoginID,
				Claims: map[string]interface{}{
					identity.StandardClaimEmail: "foo@example.com",
				},
			})), ShouldEqual, StatusVerified)
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
						ID:     "email",
						Type:   authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{authenticator.AuthenticatorClaimOOBOTPEmail: "foo@example.com"},
					}},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("username", "foo"),
					},
					Authenticators: []*authenticator.Info{{
						ID:     "phone",
						Type:   authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{authenticator.AuthenticatorClaimOOBOTPPhone: "+85200000000"},
					}},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityLoginID("username", "foo"),
					},
					Authenticators: []*authenticator.Info{{
						ID:     "email",
						Type:   authn.AuthenticatorTypeOOB,
						Claims: map[string]interface{}{authenticator.AuthenticatorClaimOOBOTPEmail: "foo@example.com"},
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
