package oidc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const PrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC89eQDeH8icj6j
1DUHTXKyhFkYOVrOVLA4xflDwqAuw5IrJQNgIjTsBZXrR1rh4BSBsjoE0ToH+/Da
MfyAicQpv7QPI4pM8a/a3SY+rlr4j4LzFtchUvBMcGbSZZqKINBtxpAsFLPGFnwF
NrxXIwrxE79cgY+g1KcmF8twqDmmash6fMoOeU8MTa8Q9Z7wTzhySeeZlBVFtvJp
79Wqe75dtp0pe6E6ujavVjPifj2Msdl9RW7KhJsttgGhMGR2Jp07nAIBT150qX0G
3gu0G5ILbgxcrhYZYK5fk/u6MQ0sAyXwS+fmppsPmYw6UVYlS2UGnaJlCE7Ml0e2
yyEyrbmnAgMBAAECggEARX7NsDUV1O5deVVnd1sVjvA78DvP2Miu0wKErVYcIXbO
AE4pkqah/hgDzjc9BouqHxUUX4cvp5YSO71cl02TtqMJrvOsPqY4ve7NzQnE7Vui
lpLU5i2hsQs51bGGh7yPy3/WsE+g2n6UeDpsREPgF0/i9ju0PjtXihwAN1u3cCt9
t9CsSGliHqQX9uO7o92yN+aROKEbw3x3gKpRJ/Gv3fQcVR01cXvaBrtdEb6kEVEB
WBlCA0kmRc/H7jVYGcWqalLDjj99Pox47PLUigyJsNxJmMD881Ihah4zEQMpX7pW
eRuyISTAA+i0MXO8+bypE6trglF8YQH6JTcVLTz70QKBgQDngFYD0gAqB41vMpmQ
TGSr16qs63Q9QD0Ot9ZkSvYY745HvK7syLq6FLZl5Qz/f45NQ99BQtkGDZjE8sn9
W4V7/yA8xzNP+xmvdqsoOAcIO4j8W34dA6gS+z4h2u98LpqV9Q6ehbrZCPB8/MSn
1QTnbINGw1ZCxfj6olN7ppaZaQKBgQDQ9RIHxHVhDHWpa83Fgf0oDve2UXD5YDsZ
Axu6cYQOGCM7h0WxwDViIUuieWortYvGq8K1IlqfaDWlo5BHRXozmakMpJ4K8sBW
F8TWn7PYw9cPH2XuZZHPnPiYkkhe0SoAifa3tk4bgyR5txOjdCr5L3ZFWfM7Vmkp
hL2M7JTIjwKBgDXyShkJzs/8gpDvEan2o18IGtXA6I19cr0DSgqFDWQyLs24wmqb
PCgwu3BzN9wyNU78CgKDOV+Xu4npqfhIY4rJoRGIugRhV1L0LF5q7/iTJxDnoTPR
rlD+CzSIeFZP5eYb/RQjxa7dzmzR2mHh2gqz1sOesXNN/v8o5Jtj7qRBAoGAEibH
yy7wt156th3sQRT6pckvEYJfmvoWCCUx+m8z9nl4TgqBLmCxAnY7+MAtTeC2ZKq0
/kEeuCw4RMxBkz9gzyyw960xIWhW9uOXsMEswU651tF2bFAca3mKSs6iRMJMsMFL
Ukge3tr0hzI1HYTQ2taZooqey2/FMNscECrY/dcCgYEAuhBfEof+DCeuLmKgvks+
Idv5Ky2ZIR49L8VxCy7K+BXhr2vnKX6itlVDQVVpNIphdLHXQK6CNr8Ko5WinZHu
gouLseU4p4zh8vYZcgPyqlLEdkygMCN0b0+HVaBTs0jlLGbvTC0Oiz69umYMe+5g
eZDnqWNf7mYPdP5mO5iTtMw=
-----END PRIVATE KEY-----
`

func TestIDTokenIssuer(t *testing.T) {
	Convey("PrepareIDToken, MakeIDTokenFromPreparationResult, and VerifyIDToken", t, func() {
		ctrl := gomock.NewController(t)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)
		jwkKey, _ := jwkSet.Key(0)
		_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
		_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

		secrets := &config.OAuthKeyMaterials{
			Set: jwkSet,
		}

		mockUserInfoService := NewMockUserInfoService(ctrl)
		mockUserInfoService.EXPECT().GetUserInfoBearer(gomock.Any(), "user-id").Return(
			&userinfo.UserInfo{
				User: &model.User{
					IsAnonymous:        false,
					IsVerified:         true,
					CanReauthenticate:  true,
					HasPrimaryPassword: true,
				},
				EffectiveRoleKeys: []string{"role-1", "role-3"},
			},
			nil,
		)

		mockEventService := NewMockIDTokenIssuerEventService(ctrl)
		mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, e event.Payload) (*event.Event, error) {
			return &event.Event{
				Payload: e,
			}, nil
		}).AnyTimes()
		mockEventService.EXPECT().DispatchEventWithoutTx(gomock.Any(), gomock.Any()).Return(nil)

		mockIdentityService := NewMockIDTokenIssuerIdentityService(ctrl)
		mockIdentityService.EXPECT().ListIdentitiesThatHaveStandardAttributes(gomock.Any(), "user-id").Return(nil, nil)

		issuer := &IDTokenIssuer{
			Secrets: secrets,
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test.authgear.com",
					HTTPProto: "http",
				},
			},
			UserInfoService: mockUserInfoService,
			Events:          mockEventService,
			Identities:      mockIdentityService,
			Clock:           clock.NewMockClockAtTime(now),
		}

		client := &config.OAuthClientConfig{
			ClientID: "client-id",
		}
		scopes := []string{"openid", "email"}
		refreshToken := oauth.OfflineGrantRefreshToken{
			ClientID: client.ClientID,
			Scopes:   scopes,
		}

		testDeviceSecretHash := "devicesecrethash"

		offlineGrant := &oauth.OfflineGrant{
			ID:            "offline-grant-id",
			RefreshTokens: []oauth.OfflineGrantRefreshToken{refreshToken},
			Attrs: session.Attrs{
				UserID: "user-id",
			},
			DeviceSecretHash: testDeviceSecretHash,
		}

		ctx := context.Background()
		preparationResult, err := issuer.PrepareIDToken(ctx, PrepareIDTokenOptions{
			ClientID:           "client-id",
			SID:                oauth.EncodeSID(offlineGrant),
			AuthenticationInfo: offlineGrant.GetAuthenticationInfo(),
			ClientLike:         oauth.ClientClientLike(client, scopes),
			Nonce:              "nonce-1",
			DeviceSecretHash:   testDeviceSecretHash,
			IdentitySpecs: []*identity.Spec{
				{
					Type: model.IdentityTypeOAuth,
					OAuth: &identity.OAuthSpec{
						IncludeIdentityAttributesInIDToken: true,
						ProviderAlias:                      "google",
						ProviderID:                         oauthrelyingparty.NewProviderID("google", make(map[string]any)),
						SubjectID:                          "google-user-id",
						RawProfile: map[string]any{
							"google_specific_field": 42,
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)

		idToken, err := issuer.MakeIDTokenFromPreparationResult(ctx, MakeIDTokenFromPreparationResultOptions{
			PreparationResult: preparationResult,
		})
		So(err, ShouldBeNil)

		token, err := issuer.VerifyIDToken(idToken)
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
		_, sessionID, _ := oauth.DecodeSID(encodedSessionID.(string))
		So(sessionID, ShouldEqual, offlineGrant.ID)

		ds_hash, _ := token.Get(string(model.ClaimDeviceSecretHash))
		So(ds_hash, ShouldEqual, offlineGrant.DeviceSecretHash)

		// Authz-specific claims
		nonce, _ := token.Get(string("nonce"))
		So(nonce, ShouldEqual, "nonce-1")

		// Authgear-specific claims
		oauthUsed, ok := token.Get("https://authgear.com/claims/oauth/asserted")
		So(ok, ShouldBeTrue)
		So(oauthUsed, ShouldResemble, []any{
			map[string]any{
				"https://authgear.com/claims/oauth/profile": map[string]any{
					"google_specific_field": float64(42),
				},
				"https://authgear.com/claims/oauth/provider_alias": "google",
				"https://authgear.com/claims/oauth/provider_type":  "google",
				"https://authgear.com/claims/oauth/subject_id":     "google-user-id",
			},
		},
		)
	})

	Convey("GetUserInfo", t, func() {
		ctrl := gomock.NewController(t)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)
		jwkKey, _ := jwkSet.Key(0)
		_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
		_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

		secrets := &config.OAuthKeyMaterials{
			Set: jwkSet,
		}

		mockUserInfoService := NewMockUserInfoService(ctrl)
		mockUserInfoService.EXPECT().GetUserInfoBearer(gomock.Any(), "user-id").Return(
			&userinfo.UserInfo{
				User: &model.User{
					IsAnonymous:        false,
					IsVerified:         true,
					CanReauthenticate:  true,
					HasPrimaryPassword: true,
				},
				EffectiveRoleKeys: []string{"role-1", "role-3"},
			},
			nil,
		)

		issuer := &IDTokenIssuer{
			Secrets: secrets,
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test.authgear.com",
					HTTPProto: "http",
				},
			},
			UserInfoService: mockUserInfoService,
			Clock:           clock.NewMockClockAtTime(now),
		}

		client := &config.OAuthClientConfig{
			ClientID: "client-id",
		}
		scopes := []string{"openid", "email"}

		ctx := context.Background()
		userInfoClaims, err := issuer.GetUserInfo(ctx, "user-id", oauth.ClientClientLike(client, scopes))
		So(err, ShouldBeNil)

		So(userInfoClaims[string(model.ClaimUserHasPrimaryPassword)], ShouldEqual, true)
		So(userInfoClaims[string(model.ClaimUserIsAnonymous)], ShouldEqual, false)
		So(userInfoClaims[string(model.ClaimUserIsVerified)], ShouldEqual, true)
		So(userInfoClaims[string(model.ClaimUserCanReauthenticate)], ShouldEqual, true)
		So(userInfoClaims[string(model.ClaimAuthgearRoles)], ShouldResemble, []string{"role-1", "role-3"})
		So(userInfoClaims["sub"], ShouldEqual, "user-id")
	})
}
