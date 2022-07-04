package verification

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestService(t *testing.T) {

	Convey("Service", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		claimStore := NewMockClaimStore(ctrl)
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
			ClaimStore: claimStore,
		}

		verifiedClaim := func(userID string, name string, value string) *Claim {
			return &Claim{
				ID:     "claim-id",
				UserID: userID,
				Name:   name,
				Value:  value,
			}
		}

		identityLoginID := func(loginIDKey string, loginIDValue string) *identity.Info {
			i := &identity.Info{
				UserID: "user-id",
				ID:     "login-id-" + loginIDValue,
				Type:   model.IdentityTypeLoginID,
				Claims: map[string]interface{}{},
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

		identityOfType := func(t model.IdentityType, claims map[string]interface{}) *identity.Info {
			return &identity.Info{
				UserID: "user-id",
				ID:     string(t),
				Type:   t,
				Claims: claims,
			}
		}

		mustBool := func(value bool, err error) bool {
			So(err, ShouldBeNil)
			return value
		}

		Convey("IsClaimVerifiable", func() {
			So(service.IsClaimVerifiable("email"), ShouldBeTrue)
			So(service.IsClaimVerifiable("phone_number"), ShouldBeTrue)
			So(service.IsClaimVerifiable("username"), ShouldBeFalse)
		})

		Convey("IsVerified", func() {
			cases := []struct {
				Identities []*identity.Info
				Claims     []*Claim
				AnyResult  bool
				AllResult  bool
			}{
				{
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityOfType(model.IdentityTypeAnonymous, nil),
					},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityOfType(model.IdentityTypeOAuth, nil),
					},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOfType(model.IdentityTypeOAuth, nil),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOfType(model.IdentityTypeOAuth, map[string]interface{}{"email": "bar@example.com"}),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOfType(model.IdentityTypeOAuth, map[string]interface{}{"email": "bar@example.com"}),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
						verifiedClaim("user-id", "email", "bar@example.com"),
					},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("username", "foo"),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "phone", "+85212345678"),
					},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityLoginID("username", "foo"),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: true,
				},
			}

			for i, c := range cases {
				Convey(fmt.Sprintf("case %d", i), func() {
					service.Config.Criteria = config.VerificationCriteriaAny
					claimStore.EXPECT().ListByUser("user-id").Return(c.Claims, nil).MaxTimes(1)
					So(mustBool(service.IsUserVerified(c.Identities)), ShouldEqual, c.AnyResult)

					service.Config.Criteria = config.VerificationCriteriaAll
					claimStore.EXPECT().ListByUser("user-id").Return(c.Claims, nil).MaxTimes(1)
					So(mustBool(service.IsUserVerified(c.Identities)), ShouldEqual, c.AllResult)
				})
			}
		})
	})
}
