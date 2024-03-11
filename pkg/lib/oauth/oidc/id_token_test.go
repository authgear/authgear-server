package oidc

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/rand"
	utilsecrets "github.com/authgear/authgear-server/pkg/util/secrets"
)

type FakeSession struct {
	ID   string
	Type session.Type
}

func (s *FakeSession) SessionID() string {
	return s.ID
}

func (s *FakeSession) SessionType() session.Type {
	return s.Type
}

func TestSID(t *testing.T) {
	Convey("EncodeSID and DecodeSID", t, func() {
		s := &FakeSession{
			ID:   "a",
			Type: session.TypeIdentityProvider,
		}
		typ, sessionID, ok := DecodeSID(EncodeSID(s))
		So(typ, ShouldEqual, session.TypeIdentityProvider)
		So(sessionID, ShouldEqual, "a")
		So(ok, ShouldBeTrue)

		s = &FakeSession{
			ID:   "b",
			Type: session.TypeOfflineGrant,
		}
		typ, sessionID, ok = DecodeSID(EncodeSID(s))
		So(typ, ShouldEqual, session.TypeOfflineGrant)
		So(sessionID, ShouldEqual, "b")
		So(ok, ShouldBeTrue)

		s = &FakeSession{
			ID:   "c",
			Type: "nonsense",
		}
		_, _, ok = DecodeSID(EncodeSID(s))
		So(ok, ShouldBeFalse)
	})

	Convey("IssueIDToken and VerifyIDTokenHint", t, func() {
		ctrl := gomock.NewController(t)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		secrets := &config.OAuthKeyMaterials{
			Set: jwk.NewSet(),
		}
		jwkKey := utilsecrets.GenerateRSAKey(now, rand.SecureRand)
		err := secrets.Set.AddKey(jwkKey)
		So(err, ShouldBeNil)

		mockUserProvider := NewMockUserProvider(ctrl)
		mockRolesAndGroupsProvider := NewMockRolesAndGroupsProvider(ctrl)

		mockUserProvider.EXPECT().Get("user-id", gomock.Any()).DoAndReturn(
			func(id string, role accesscontrol.Role) (*model.User, error) {
				return &model.User{
					IsAnonymous:       false,
					IsVerified:        true,
					CanReauthenticate: true,
				}, nil
			})

		mockRolesAndGroupsProvider.EXPECT().ListEffectiveRolesByUserID("user-id").DoAndReturn(
			func(userID string) ([]*model.Role, error) {
				return []*model.Role{
					{
						Key: "role-1",
					},
					{
						Key: "role-3",
					},
				}, nil
			})

		issuer := &IDTokenIssuer{
			Secrets: secrets,
			BaseURL: &endpoints.Endpoints{
				HTTPHost:  "test.authgear.com",
				HTTPProto: "http",
			},
			Users:          mockUserProvider,
			RolesAndGroups: mockRolesAndGroupsProvider,
			Clock:          clock.NewMockClockAtTime(now),
		}

		client := &config.OAuthClientConfig{
			ClientID: "client-id",
		}
		scopes := []string{"openid", "email"}

		offlineGrant := &oauth.OfflineGrant{
			ID:     "offline-grant-id",
			Scopes: scopes,
			Attrs: session.Attrs{
				UserID: "user-id",
			},
		}

		idToken, err := issuer.IssueIDToken(IssueIDTokenOptions{
			ClientID:           "client-id",
			SID:                EncodeSID(offlineGrant),
			AuthenticationInfo: offlineGrant.GetAuthenticationInfo(),
			ClientLike:         oauth.ClientClientLike(client, scopes),
			Nonce:              "nonce-1",
		})
		So(err, ShouldBeNil)

		token, err := issuer.VerifyIDTokenHint(client, idToken)
		So(err, ShouldBeNil)

		// Standard claims
		So(token.Issuer(), ShouldEqual, "http://test.authgear.com")
		So(token.Subject(), ShouldEqual, "user-id")
		So(token.Audience(), ShouldResemble, []string{"client-id"})
		So(token.IssuedAt(), ShouldEqual, now)
		So(token.Expiration().Equal(now.Add(IDTokenValidDuration)), ShouldBeTrue)

		// User claims
		isAnonymous, _ := token.Get(string(model.ClaimUserIsAnonymous))
		isVerified, _ := token.Get(string(model.ClaimUserIsVerified))
		canReauthenticate, _ := token.Get(string(model.ClaimUserCanReauthenticate))
		roles, _ := token.Get(string(model.ClaimAuthgearRoles))

		So(isAnonymous, ShouldEqual, false)
		So(isVerified, ShouldEqual, true)
		So(canReauthenticate, ShouldEqual, true)
		So(roles, ShouldResemble, []interface{}{"role-1", "role-3"})

		// Session claims
		encodedSessionID, _ := token.Get(string(model.ClaimSID))
		_, sessionID, _ := DecodeSID(encodedSessionID.(string))
		So(sessionID, ShouldEqual, offlineGrant.ID)

		// Authz-specific claims
		nonce, _ := token.Get(string("nonce"))
		So(nonce, ShouldEqual, "nonce-1")
	})
}
