package oauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
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

func TestAccessToken(t *testing.T) {
	Convey("EncodeAccessToken and DecodeAccessToken", t, func() {
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

		mockIDTokenIssuer := NewMockIDTokenIssuer(ctrl)
		mockEventService := NewMockEventService(ctrl)

		encoding := &AccessTokenEncoding{
			Secrets:       secrets,
			Clock:         clock.NewMockClockAtTime(now),
			IDTokenIssuer: mockIDTokenIssuer,
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test1.authgear.com",
					HTTPProto: "http",
				},
			},
			Events: mockEventService,
			UserBlockingEventContexts: &UserBlockingEventContextProvider{
				Users:      &stubUserBlockingEventContextUserService{user: &model.User{Meta: model.Meta{ID: "user-id"}}},
				Identities: &stubUserBlockingEventContextIdentityService{},
			},
		}

		client := &config.OAuthClientConfig{
			IssueJWTAccessToken: true,
			ClientID:            "client-id",
			AccessTokenLifetime: 3600,
		}
		var noScopes []string
		clientLike := ClientClientLike(client, noScopes)

		accessGrant := &AccessGrant{
			CreatedAt: now,
			ExpireAt:  now.Add(client.AccessTokenLifetime.Duration()),
			TokenHash: "token-hash",
			Scopes:    []string{"openid", "email"},
		}

		mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
			return &event.Event{
				Payload: e,
			}, nil
		}).AnyTimes()
		mockEventService.EXPECT().DispatchEventWithoutTx(gomock.Any(), gomock.Any()).Return(nil)
		mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate).Return(true).AnyTimes()
		mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
		mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)

		ctx := context.Background()
		options := EncodeUserAccessTokenOptions{
			OriginalToken: "token",
			ClientConfig:  client,
			ClientLike:    clientLike,
			AccessGrant:   accessGrant,
			AuthenticationInfo: authenticationinfo.T{
				UserID: "user-id",
				// AMR
				// AuthenticatedAt
			},
		}

		preparation, err := encoding.PrepareUserAccessToken(ctx, options)
		So(err, ShouldBeNil)

		tokenResult, err := encoding.MakeUserAccessTokenFromPreparationResult(ctx, MakeUserAccessTokenFromPreparationOptions{
			PreparationResult: preparation,
		})
		So(err, ShouldBeNil)

		_, _, err = encoding.DecodeAccessToken(tokenResult.Token)
		So(err, ShouldBeNil)

		// Peek token payload
		keys, err := jwk.PublicSetOf(encoding.Secrets.Set)
		So(err, ShouldBeNil)

		decodedToken, _ := jwt.ParseString(tokenResult.Token, jwt.WithKeySet(keys), jwt.WithValidate(false))
		So(err, ShouldBeNil)

		clientID, _ := decodedToken.Get("client_id")
		idKey, _ := decodedToken.Get(jwt.JwtIDKey)
		scope, _ := decodedToken.Get("scope")

		So(decodedToken.Issuer(), ShouldEqual, "http://test1.authgear.com")
		So(decodedToken.Audience(), ShouldResemble, []string{"http://test1.authgear.com"})
		So(decodedToken.IssuedAt(), ShouldEqual, accessGrant.CreatedAt)
		So(decodedToken.Expiration(), ShouldEqual, accessGrant.ExpireAt)
		So(clientID, ShouldEqual, "client-id")
		So(scope, ShouldEqual, "openid email")
		So(idKey, ShouldEqual, "token-hash")
	})
}

func TestDecodeAccessTokenResourceBound(t *testing.T) {
	Convey("DecodeAccessToken with a resource-bound aud", t, func() {
		// jwt.ParseString validates exp/iat against the real wall clock at
		// parse time (before DecodeAccessToken's own jwt.Validate call with
		// the injected mock clock ever runs), so a fixed past date would be
		// seen as already-expired regardless of what this test intends —
		// use a real-time-relative base instead.
		now := time.Now().UTC().Truncate(time.Second)

		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)
		jwkKey, _ := jwkSet.Key(0)
		_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
		_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

		secrets := &config.OAuthKeyMaterials{Set: jwkSet}

		encoding := &AccessTokenEncoding{
			Secrets: secrets,
			Clock:   clock.NewMockClockAtTime(now),
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test1.authgear.com",
					HTTPProto: "http",
				},
			},
		}

		sign := func(signingKey jwk.Key, aud string, exp time.Time) string {
			claims := jwt.New()
			_ = claims.Set(jwt.JwtIDKey, "token-hash")
			_ = claims.Set(jwt.IssuerKey, "http://test1.authgear.com")
			_ = claims.Set(jwt.AudienceKey, []string{aud})
			_ = claims.Set(jwt.IssuedAtKey, now.Unix())
			_ = claims.Set(jwt.ExpirationKey, exp.Unix())

			hdr := jws.NewHeaders()
			_ = hdr.Set("typ", "at+jwt")
			signed, err := jwtutil.SignWithHeader(claims, hdr, jwa.RS256, signingKey)
			So(err, ShouldBeNil)
			return string(signed)
		}

		Convey("a JWT with aud=[resource_uri] (no project endpoint) decodes successfully", func() {
			token := sign(jwkKey, "https://api.example.com/orders", now.Add(time.Hour))
			jti, isHash, err := encoding.DecodeAccessToken(token)
			So(err, ShouldBeNil)
			So(isHash, ShouldBeTrue)
			So(jti, ShouldEqual, "token-hash")
		})

		// jwt.ParseString rejects an expired/badly-signed token outright, and
		// DecodeAccessToken's "Invalid JWT string" branch treats any parse
		// failure as "not a JWT" rather than surfacing an error — the token
		// falls back to being treated as opaque (and will then fail to match
		// any AccessGrant downstream in oauth.Resolver, since its raw string
		// was never issued as an opaque token). So the observable contract
		// here is isHash=false, err=nil, not a returned error.
		Convey("an expired resource-bound JWT falls back to being treated as opaque", func() {
			token := sign(jwkKey, "https://api.example.com/orders", now.Add(-time.Hour))
			_, isHash, err := encoding.DecodeAccessToken(token)
			So(err, ShouldBeNil)
			So(isHash, ShouldBeFalse)
		})

		Convey("a resource-bound JWT signed by a foreign key falls back to being treated as opaque", func() {
			foreignRSAKey, err := rsa.GenerateKey(rand.Reader, 2048)
			So(err, ShouldBeNil)
			foreignKey, err := jwk.FromRaw(foreignRSAKey)
			So(err, ShouldBeNil)
			_ = foreignKey.Set(jwk.KeyIDKey, uuid.New())
			_ = foreignKey.Set(jwk.AlgorithmKey, "RS256")

			token := sign(foreignKey, "https://api.example.com/orders", now.Add(time.Hour))
			_, isHash, err := encoding.DecodeAccessToken(token)
			So(err, ShouldBeNil)
			So(isHash, ShouldBeFalse)
		})
	})
}

func TestClientAccessToken(t *testing.T) {
	Convey("EncodeClientAccessToken", t, func() {
		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)
		jwkKey, _ := jwkSet.Key(0)
		_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
		_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

		secrets := &config.OAuthKeyMaterials{
			Set: jwkSet,
		}

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIDTokenIssuer := NewMockIDTokenIssuer(mockCtrl)
		mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")

		encoding := &AccessTokenEncoding{
			Secrets:       secrets,
			Clock:         clock.NewMockClockAtTime(now),
			IDTokenIssuer: mockIDTokenIssuer,
			BaseURL: &endpoints.Endpoints{
				OAuthEndpoints: &endpoints.OAuthEndpoints{
					HTTPHost:  "test1.authgear.com",
					HTTPProto: "http",
				},
			},
		}

		client := &config.OAuthClientConfig{
			IssueJWTAccessToken: true,
			ClientID:            "client-id",
			AccessTokenLifetime: 3600,
		}
		resourceURI := "https://api.example.com/"
		scope := "read write"
		createdAt := now
		expireAt := now.Add(client.AccessTokenLifetime.Duration())
		originalToken := "opaque-token" // #nosec G101

		options := EncodeClientAccessTokenOptions{
			OriginalToken: originalToken,
			ClientConfig:  client,
			ResourceURI:   resourceURI,
			Scope:         scope,
			CreatedAt:     createdAt,
			ExpireAt:      expireAt,
		}

		accessToken, err := encoding.EncodeClientAccessToken(context.Background(), options)
		So(err, ShouldBeNil)

		// Peek token payload
		keys, err := jwk.PublicSetOf(encoding.Secrets.Set)
		So(err, ShouldBeNil)

		decodedToken, err := jwt.ParseString(accessToken, jwt.WithKeySet(keys), jwt.WithValidate(false))
		So(err, ShouldBeNil)

		clientID, _ := decodedToken.Get("client_id")
		scopeClaim, _ := decodedToken.Get("scope")
		sub, _ := decodedToken.Get("sub")
		aud := decodedToken.Audience()
		iss := decodedToken.Issuer()
		iat := decodedToken.IssuedAt()
		exp := decodedToken.Expiration()

		So(clientID, ShouldEqual, "client-id")
		So(scopeClaim, ShouldEqual, scope)
		So(sub, ShouldEqual, "client_id_client-id")
		So(aud, ShouldResemble, []string{resourceURI})
		So(iss, ShouldEqual, "http://test1.authgear.com")
		So(iat, ShouldEqual, createdAt)
		So(exp, ShouldEqual, expireAt)
	})
}

func TestAccessTokenEncoding_PrepareUserAccessToken_UserBlockingEventContext(t *testing.T) {
	Convey("PrepareUserAccessToken threads UserBlockingEventContext to PrepareBlockingEventWithTx", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		jwkSet, err := jwk.Parse([]byte(PrivateKeyPEM), jwk.WithPEM(true))
		So(err, ShouldBeNil)
		jwkKey, _ := jwkSet.Key(0)
		_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
		_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")
		secrets := &config.OAuthKeyMaterials{Set: jwkSet}

		client := &config.OAuthClientConfig{
			IssueJWTAccessToken: true,
			ClientID:            "client-id",
			AccessTokenLifetime: 3600,
		}
		clientLike := ClientClientLike(client, []string{"openid"})
		accessGrant := &AccessGrant{
			CreatedAt: now,
			ExpireAt:  now.Add(client.AccessTokenLifetime.Duration()),
			TokenHash: "token-hash",
			Scopes:    []string{"openid"},
		}

		baseURL := &endpoints.Endpoints{
			OAuthEndpoints: &endpoints.OAuthEndpoints{
				HTTPHost:  "test1.authgear.com",
				HTTPProto: "http",
			},
		}

		baseOptions := func() EncodeUserAccessTokenOptions {
			return EncodeUserAccessTokenOptions{
				OriginalToken:      "token",
				ClientConfig:       client,
				ClientLike:         clientLike,
				AccessGrant:        accessGrant,
				AuthenticationInfo: authenticationinfo.T{UserID: "user-id"},
			}
		}

		Convey("WillDeliverBlockingEvent false: the provider is never called and ResolvedUser is nil", func() {
			mockIDTokenIssuer := NewMockIDTokenIssuer(ctrl)
			mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
			mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)

			mockEventService := NewMockEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate).Return(false)
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			encoding := &AccessTokenEncoding{
				Secrets:       secrets,
				Clock:         clock.NewMockClockAtTime(now),
				IDTokenIssuer: mockIDTokenIssuer,
				BaseURL:       baseURL,
				Events:        mockEventService,
			}

			_, err := encoding.PrepareUserAccessToken(context.Background(), baseOptions())
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldBeNil)
		})

		Convey("WillDeliverBlockingEvent true with opts.UserBlockingEventContext supplied: the provider is not called", func() {
			mockIDTokenIssuer := NewMockIDTokenIssuer(ctrl)
			mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
			mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)

			mockEventService := NewMockEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate).Return(true)
			var capturedPayload *blocking.OIDCJWTPreCreateBlockingEventPayload
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedPayload = e.(*blocking.OIDCJWTPreCreateBlockingEventPayload)
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			suppliedUser := &model.User{Meta: model.Meta{ID: "user-id"}, IsVerified: true}
			suppliedIdentities := []model.Identity{{Meta: model.Meta{ID: "identity-1"}}}

			encoding := &AccessTokenEncoding{
				Secrets:       secrets,
				Clock:         clock.NewMockClockAtTime(now),
				IDTokenIssuer: mockIDTokenIssuer,
				BaseURL:       baseURL,
				Events:        mockEventService,
				// Left nil: it must not be called in this case. Calling it
				// would panic on the nil pointer, failing the test.
				UserBlockingEventContexts: nil,
			}

			options := baseOptions()
			options.UserBlockingEventContext = &UserBlockingEventContext{
				UserID:     "user-id",
				UserModel:  suppliedUser,
				Identities: suppliedIdentities,
			}

			_, err := encoding.PrepareUserAccessToken(context.Background(), options)
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldEqual, suppliedUser)
			So(capturedPayload.Identities, ShouldResemble, suppliedIdentities)
		})

		Convey("WillDeliverBlockingEvent true with a supplied context whose UserID does not match: the provider is called", func() {
			mockIDTokenIssuer := NewMockIDTokenIssuer(ctrl)
			mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
			mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)

			mockEventService := NewMockEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate).Return(true)
			var capturedPayload *blocking.OIDCJWTPreCreateBlockingEventPayload
			var capturedOpts event.PrepareBlockingEventOptions
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					capturedPayload = e.(*blocking.OIDCJWTPreCreateBlockingEventPayload)
					capturedOpts = opts
					return &event.Event{Payload: e}, nil
				})

			providerUser := &model.User{Meta: model.Meta{ID: "user-id"}, IsVerified: true}
			providerIdentity := &identity.Info{
				ID:        "identity-from-provider",
				Type:      model.IdentityTypeAnonymous,
				Anonymous: &identity.Anonymous{KeyID: "key-id"},
			}

			encoding := &AccessTokenEncoding{
				Secrets:       secrets,
				Clock:         clock.NewMockClockAtTime(now),
				IDTokenIssuer: mockIDTokenIssuer,
				BaseURL:       baseURL,
				Events:        mockEventService,
				UserBlockingEventContexts: &UserBlockingEventContextProvider{
					Users:      &stubUserBlockingEventContextUserService{user: providerUser},
					Identities: &stubUserBlockingEventContextIdentityService{identities: []*identity.Info{providerIdentity}},
				},
			}

			options := baseOptions()
			options.UserBlockingEventContext = &UserBlockingEventContext{
				UserID:    "other-user",
				UserModel: &model.User{Meta: model.Meta{ID: "other-user"}},
			}

			_, err := encoding.PrepareUserAccessToken(context.Background(), options)
			So(err, ShouldBeNil)
			So(capturedOpts.ResolvedUser, ShouldEqual, providerUser)
			So(capturedPayload.Identities, ShouldResemble, []model.Identity{providerIdentity.ToModel()})
		})

		Convey("WillDeliverBlockingEvent true with no supplied context: the provider is called once", func() {
			mockIDTokenIssuer := NewMockIDTokenIssuer(ctrl)
			mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
			mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)

			mockEventService := NewMockEventService(ctrl)
			mockEventService.EXPECT().WillDeliverBlockingEvent(blocking.OIDCJWTPreCreate).Return(true)
			mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, e event.Payload, opts event.PrepareBlockingEventOptions) (*event.Event, error) {
					return &event.Event{Payload: e}, nil
				})

			providerUser := &model.User{Meta: model.Meta{ID: "user-id"}}
			callCount := 0
			encoding := &AccessTokenEncoding{
				Secrets:       secrets,
				Clock:         clock.NewMockClockAtTime(now),
				IDTokenIssuer: mockIDTokenIssuer,
				BaseURL:       baseURL,
				Events:        mockEventService,
				UserBlockingEventContexts: &UserBlockingEventContextProvider{
					Users: &countingUserBlockingEventContextUserService{
						stubUserBlockingEventContextUserService: stubUserBlockingEventContextUserService{user: providerUser},
						calls:                                   &callCount,
					},
					Identities: &stubUserBlockingEventContextIdentityService{},
				},
			}

			_, err := encoding.PrepareUserAccessToken(context.Background(), baseOptions())
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
