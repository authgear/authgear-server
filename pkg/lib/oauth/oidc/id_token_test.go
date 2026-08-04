package oidc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/userinfo"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type stubUserBlockingEventContextUserService struct {
	user *model.User
	err  error
}

func (s *stubUserBlockingEventContextUserService) Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
	return s.user, s.err
}

type stubUserBlockingEventContextIdentityService struct {
	identities []*identity.Info
	err        error
}

func (s *stubUserBlockingEventContextIdentityService) ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error) {
	return s.identities, s.err
}

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
					IsAnonymous:       false,
					IsVerified:        true,
					CanReauthenticate: true,
				},
				EffectiveRoleKeys:   []string{"role-1", "role-3"},
				RecoveryCodeEnabled: true,
			},
			nil,
		)

		mockEventService := NewMockIDTokenIssuerEventService(ctrl)
		mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
			return &event.Event{
				Payload: e,
			}, nil
		}).AnyTimes()
		mockEventService.EXPECT().DispatchEventWithoutTx(gomock.Any(), gomock.Any()).Return(nil)
		mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate).Return(true)

		eventUserCtxProvider := &oauth.UserBlockingEventContextProvider{
			Users:      &stubUserBlockingEventContextUserService{user: &model.User{Meta: model.Meta{ID: "user-id"}}},
			Identities: &stubUserBlockingEventContextIdentityService{},
		}

		issuer := &IDTokenIssuer{
			Secrets: secrets,
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test.authgear.com",
					HTTPProto: "http",
				},
			},
			UserInfoService:           mockUserInfoService,
			Events:                    mockEventService,
			UserBlockingEventContexts: eventUserCtxProvider,
			Clock:                     clock.NewMockClockAtTime(now),
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
		So(roles, ShouldResemble, []any{"role-1", "role-3"})

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
}

func TestIDTokenIssuer_GetUserInfo(t *testing.T) {
	Convey("GetUserInfo", t, func() {
		ctrl := gomock.NewController(t)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		createdAt := now.Add(-1 * time.Hour)
		updatedAt := now.Add(-30 * time.Minute)

		mockUserInfoService := NewMockUserInfoService(ctrl)
		mockUserInfoService.EXPECT().GetUserInfoBearer(gomock.Any(), "user-id").Return(
			&userinfo.UserInfo{
				User: &model.User{
					IsAnonymous:       false,
					IsVerified:        true,
					CanReauthenticate: true,
				},
				EffectiveRoleKeys:   []string{"role-1", "role-3"},
				RecoveryCodeEnabled: true,
				Authenticators: []model.UserInfoAuthenticator{
					{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Type:      model.AuthenticatorTypePassword,
						Kind:      model.AuthenticatorKindPrimary,
					},
					{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Type:      model.AuthenticatorTypeOOBSMS,
						Kind:      model.AuthenticatorKindPrimary,
					},
					{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Type:      model.AuthenticatorTypeOOBEmail,
						Kind:      model.AuthenticatorKindPrimary,
					},
					{
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						Type:      model.AuthenticatorTypeTOTP,
						Kind:      model.AuthenticatorKindPrimary,
					},
				},
				Identities: []model.UserInfoIdentity{
					{CreatedAt: createdAt, UpdatedAt: updatedAt, Type: model.IdentityTypeOAuth, OAuthProviderType: "google", OAuthProviderAlias: "google"},
					{CreatedAt: createdAt, UpdatedAt: updatedAt, Type: model.IdentityTypeLoginID, LoginIDKey: "email", LoginIDType: model.LoginIDKeyTypeEmail},
				},
			},
			nil,
		)

		issuer := &IDTokenIssuer{
			UserInfoService: mockUserInfoService,
		}

		clientConfig := &config.OAuthClientConfig{
			ClientID:        "client-id",
			ApplicationType: config.OAuthClientApplicationTypeSPA,
		}
		client := oauth.ClientClientLike(clientConfig, []string{"openid", "email", oauth.FullUserInfoScope, string(model.ClaimAuthenticators), string(model.ClaimIdentities), string(model.ClaimPhoneNumber), string(model.ClaimEmail)})
		userInfo, err := issuer.GetUserInfo(context.Background(), "user-id", client)
		So(err, ShouldBeNil)

		expectedJSON := `{
  "custom_attributes": null,
  "https://authgear.com/claims/user/authenticators": [
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "password",
      "kind": "primary"
    },
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "oob_otp_sms",
      "kind": "primary"
    },
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "oob_otp_email",
      "kind": "primary"
    },
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "totp",
      "kind": "primary"
    }
  ],
  "https://authgear.com/claims/user/can_reauthenticate": true,
  "https://authgear.com/claims/user/identities": [
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "oauth",
      "oauth_provider_type": "google",
      "oauth_provider_alias": "google"
    },
    {
      "created_at": "2019-12-31T23:00:00Z",
      "updated_at": "2019-12-31T23:30:00Z",
      "type": "login_id",
      "login_id_key": "email",
      "login_id_type": "email"
    }
  ],
  "https://authgear.com/claims/user/is_anonymous": false,
  "https://authgear.com/claims/user/is_verified": true,
  "https://authgear.com/claims/user/recovery_code_enabled": true,
  "https://authgear.com/claims/user/roles": [
    "role-1",
    "role-3"
  ],
  "sub": "user-id",
  "x_web3": null
}`

		userInfoBytes, err := json.MarshalIndent(userInfo, "", "  ")
		So(err, ShouldBeNil)
		So(string(userInfoBytes), ShouldEqual, expectedJSON)
	})
}

func TestGetUserInfo(t *testing.T) {
	Convey("GetUserInfo", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		mockUserInfoService := NewMockUserInfoService(ctrl)
		mockUserInfoService.EXPECT().GetUserInfoBearer(gomock.Any(), "user-id").Return(
			&userinfo.UserInfo{
				User: &model.User{
					IsAnonymous:       false,
					IsVerified:        true,
					CanReauthenticate: true,
					StandardAttributes: map[string]any{
						"email": "test@example.com",
					},
				},
				EffectiveRoleKeys: []string{"role-1", "role-3"},
				Authenticators: []model.UserInfoAuthenticator{
					{
						Type: model.AuthenticatorTypePassword,
						Kind: model.AuthenticatorKindPrimary,
					},
				},
				Identities: []model.UserInfoIdentity{
					{CreatedAt: now, UpdatedAt: now, Type: model.IdentityTypeOAuth, OAuthProviderType: "google", OAuthProviderAlias: "google"},
				},
				RecoveryCodeEnabled: true,
			},
			nil,
		)

		issuer := &IDTokenIssuer{
			UserInfoService: mockUserInfoService,
			Clock:           clock.NewMockClockAtTime(now),
		}

		client := &config.OAuthClientConfig{
			ClientID: "client-id",
		}
		scopes := []string{"openid", "email", "https://authgear.com/scopes/full-userinfo", string(model.ClaimIdentities)}

		clientLike := oauth.ClientClientLike(client, scopes)
		clientLike.PIIAllowedInIDToken = true

		userInfo, err := issuer.GetUserInfo(context.Background(), "user-id", clientLike)
		So(err, ShouldBeNil)

		So(userInfo["sub"], ShouldEqual, "user-id")
		So(userInfo[string(model.ClaimUserIsAnonymous)], ShouldEqual, false)
		So(userInfo[string(model.ClaimUserIsVerified)], ShouldEqual, true)
		So(userInfo[string(model.ClaimUserCanReauthenticate)], ShouldEqual, true)
		So(userInfo[string(model.ClaimAuthgearRoles)], ShouldResemble, []string{"role-1", "role-3"})
		So(userInfo[string(model.ClaimRecoveryCodeEnabled)], ShouldEqual, true)
		So(userInfo["email"], ShouldEqual, "test@example.com")
		So(userInfo[string(model.ClaimAuthenticators)], ShouldResemble, []model.UserInfoAuthenticator{
			{
				Type: model.AuthenticatorTypePassword,
				Kind: model.AuthenticatorKindPrimary,
			},
		})
		So(userInfo[string(model.ClaimIdentities)], ShouldResemble, []model.UserInfoIdentity{
			{CreatedAt: now, UpdatedAt: now, Type: model.IdentityTypeOAuth, OAuthProviderType: "google", OAuthProviderAlias: "google"},
		})
	})
}

func TestIDTokenIssuer_PrepareIDToken_UserBlockingEventContext(t *testing.T) {
	Convey("PrepareIDToken threads UserBlockingEventContext to PrepareBlockingEventWithTx", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		mockUserInfoService := NewMockUserInfoService(ctrl)
		mockUserInfoService.EXPECT().GetUserInfoBearer(gomock.Any(), gomock.Any()).Return(
			&userinfo.UserInfo{User: &model.User{}},
			nil,
		).AnyTimes()

		baseOpts := func() PrepareIDTokenOptions {
			return PrepareIDTokenOptions{
				ClientID:           "client-id",
				AuthenticationInfo: authenticationInfoForUser("user-id"),
				ClientLike:         oauth.ClientClientLike(&config.OAuthClientConfig{ClientID: "client-id"}, []string{"openid"}),
			}
		}

		Convey("WillDeliverBlockingEvent false: the provider is never called and ResolvedUser is nil", func() {
			mockEventService := NewMockIDTokenIssuerEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate).Return(false)
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			issuer := &IDTokenIssuer{
				Secrets:         secretsForTest(),
				BaseURL:         baseURLForTest(),
				UserInfoService: mockUserInfoService,
				Events:          mockEventService,
				Clock:           clock.NewMockClockAtTime(now),
			}

			_, err := issuer.PrepareIDToken(context.Background(), baseOpts())
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldBeNil)
		})

		Convey("WillDeliverBlockingEvent true with opts.UserBlockingEventContext supplied: the provider is not called", func() {
			mockEventService := NewMockIDTokenIssuerEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate).Return(true)
			var capturedPayload *blocking.OIDCIDTokenPreCreateBlockingEventPayload
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedPayload = e.(*blocking.OIDCIDTokenPreCreateBlockingEventPayload)
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			suppliedUser := &model.User{Meta: model.Meta{ID: "user-id"}, IsVerified: true}
			suppliedIdentities := []model.Identity{{Meta: model.Meta{ID: "identity-1"}}}

			issuer := &IDTokenIssuer{
				Secrets:         secretsForTest(),
				BaseURL:         baseURLForTest(),
				UserInfoService: mockUserInfoService,
				Events:          mockEventService,
				// Left nil: it must not be called in this case. Calling it
				// would panic on the nil pointer, failing the test.
				UserBlockingEventContexts: nil,
				Clock:                     clock.NewMockClockAtTime(now),
			}

			opts := baseOpts()
			opts.UserBlockingEventContext = &oauth.UserBlockingEventContext{
				UserID:     "user-id",
				UserModel:  suppliedUser,
				Identities: suppliedIdentities,
			}

			_, err := issuer.PrepareIDToken(context.Background(), opts)
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldEqual, suppliedUser)
			So(capturedPayload.Identities, ShouldResemble, suppliedIdentities)
		})

		Convey("WillDeliverBlockingEvent true with a supplied context whose UserID does not match: the provider is called", func() {
			mockEventService := NewMockIDTokenIssuerEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate).Return(true)
			var capturedPayload *blocking.OIDCIDTokenPreCreateBlockingEventPayload
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedPayload = e.(*blocking.OIDCIDTokenPreCreateBlockingEventPayload)
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			providerUser := &model.User{Meta: model.Meta{ID: "user-id"}, IsVerified: true}
			providerIdentity := &identity.Info{
				ID:        "identity-from-provider",
				Type:      model.IdentityTypeAnonymous,
				Anonymous: &identity.Anonymous{KeyID: "key-id"},
			}

			issuer := &IDTokenIssuer{
				Secrets:         secretsForTest(),
				BaseURL:         baseURLForTest(),
				UserInfoService: mockUserInfoService,
				Events:          mockEventService,
				UserBlockingEventContexts: &oauth.UserBlockingEventContextProvider{
					Users:      &stubUserBlockingEventContextUserService{user: providerUser},
					Identities: &stubUserBlockingEventContextIdentityService{identities: []*identity.Info{providerIdentity}},
				},
				Clock: clock.NewMockClockAtTime(now),
			}

			opts := baseOpts()
			opts.UserBlockingEventContext = &oauth.UserBlockingEventContext{
				UserID:    "other-user",
				UserModel: &model.User{Meta: model.Meta{ID: "other-user"}},
			}

			_, err := issuer.PrepareIDToken(context.Background(), opts)
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldEqual, providerUser)
			So(capturedPayload.Identities, ShouldResemble, []model.Identity{providerIdentity.ToModel()})
		})

		Convey("WillDeliverBlockingEvent true with no supplied context: the provider is called once", func() {
			mockEventService := NewMockIDTokenIssuerEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCIDTokenPreCreate).Return(true)
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					return &event.Event{Payload: e}, nil
				})

			providerUser := &model.User{Meta: model.Meta{ID: "user-id"}}
			callCount := 0
			issuer := &IDTokenIssuer{
				Secrets:         secretsForTest(),
				BaseURL:         baseURLForTest(),
				UserInfoService: mockUserInfoService,
				Events:          mockEventService,
				UserBlockingEventContexts: &oauth.UserBlockingEventContextProvider{
					Users: &countingUserBlockingEventContextUserService{
						stubUserBlockingEventContextUserService: stubUserBlockingEventContextUserService{user: providerUser},
						calls:                                   &callCount,
					},
					Identities: &stubUserBlockingEventContextIdentityService{},
				},
				Clock: clock.NewMockClockAtTime(now),
			}

			_, err := issuer.PrepareIDToken(context.Background(), baseOpts())
			So(err, ShouldBeNil)
			So(callCount, ShouldEqual, 1)
		})
	})
}

type countingUserBlockingEventContextUserService struct {
	stubUserBlockingEventContextUserService
	calls *int
}

func (s *countingUserBlockingEventContextUserService) Get(ctx context.Context, id string, role accesscontrol.Role) (*model.User, error) {
	*s.calls++
	return s.stubUserBlockingEventContextUserService.Get(ctx, id, role)
}

func secretsForTest() *config.OAuthKeyMaterials {
	jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
	if err != nil {
		panic(err)
	}
	jwkKey, _ := jwkSet.Key(0)
	_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
	_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")
	return &config.OAuthKeyMaterials{Set: jwkSet}
}

func baseURLForTest() *endpoints.Endpoints {
	return &endpoints.Endpoints{
		OAuthEndpoints: &endpoints.OAuthEndpoints{
			HTTPHost:  "test.authgear.com",
			HTTPProto: "http",
		},
	}
}

func authenticationInfoForUser(userID string) authenticationinfo.T {
	return authenticationinfo.T{
		UserID:          userID,
		AuthenticatedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}
