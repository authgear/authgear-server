package oauth

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
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
		mockIdentityService := NewMockAccessTokenEncodingIdentityService(ctrl)

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
			Events:     mockEventService,
			Identities: mockIdentityService,
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

		mockEventService.EXPECT().PrepareBlockingEventWithTx(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, e event.Payload) (*event.Event, error) {
			return &event.Event{
				Payload: e,
			}, nil
		}).AnyTimes()
		mockEventService.EXPECT().DispatchEventWithoutTx(gomock.Any(), gomock.Any()).Return(nil)
		mockIDTokenIssuer.EXPECT().Iss().Return("http://test1.authgear.com")
		mockIDTokenIssuer.EXPECT().PopulateUserClaimsInIDToken(gomock.Any(), gomock.Any(), "user-id", clientLike).Return(nil)
		mockIdentityService.EXPECT().ListIdentitiesThatHaveStandardAttributes(gomock.Any(), "user-id").Return(nil, nil)

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
