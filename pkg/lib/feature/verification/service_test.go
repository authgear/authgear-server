package verification

import (
	"context"
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
				LoginID: &identity.LoginID{
					LoginID: loginIDValue,
					Claims:  map[string]interface{}{},
				},
			}
			switch loginIDKey {
			case "email":
				i.LoginID.LoginIDType = model.LoginIDKeyTypeEmail
				i.LoginID.Claims[string(model.ClaimEmail)] = loginIDValue
			case "phone":
				i.LoginID.LoginIDType = model.LoginIDKeyTypePhone
				i.LoginID.Claims[string(model.ClaimPhoneNumber)] = loginIDValue
			case "username":
				i.LoginID.LoginIDType = model.LoginIDKeyTypeUsername
				i.LoginID.Claims[string(model.ClaimPreferredUsername)] = loginIDValue
			}
			return i
		}

		identityAnonymous := func() *identity.Info {
			return &identity.Info{
				UserID:    "user-id",
				ID:        string(model.IdentityTypeAnonymous),
				Type:      model.IdentityTypeAnonymous,
				Anonymous: &identity.Anonymous{},
			}
		}
		identityOAuth := func(claims map[string]interface{}) *identity.Info {
			return &identity.Info{
				UserID: "user-id",
				ID:     string(model.IdentityTypeOAuth),
				Type:   model.IdentityTypeOAuth,
				OAuth: &identity.OAuth{
					Claims: claims,
				},
			}
		}

		mustBool := func(value bool, err error) bool {
			So(err, ShouldBeNil)
			return value
		}

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
						identityAnonymous(),
					},
					AnyResult: false, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityOAuth(nil),
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
						identityOAuth(nil),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: true,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOAuth(map[string]interface{}{"email": "bar@example.com"}),
					},
					Claims: []*Claim{
						verifiedClaim("user-id", "email", "foo@example.com"),
					},
					AnyResult: true, AllResult: false,
				},
				{
					Identities: []*identity.Info{
						identityLoginID("email", "foo@example.com"),
						identityOAuth(map[string]interface{}{"email": "bar@example.com"}),
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
					ctx := context.Background()
					service.Config.Criteria = config.VerificationCriteriaAny
					claimStore.EXPECT().ListByUserIDs(ctx, []string{"user-id"}).Return(c.Claims, nil).MaxTimes(1)
					So(mustBool(service.IsUserVerified(ctx, c.Identities)), ShouldEqual, c.AnyResult)

					service.Config.Criteria = config.VerificationCriteriaAll
					claimStore.EXPECT().ListByUserIDs(ctx, []string{"user-id"}).Return(c.Claims, nil).MaxTimes(1)
					So(mustBool(service.IsUserVerified(ctx, c.Identities)), ShouldEqual, c.AllResult)
				})
			}
		})
	})
}
