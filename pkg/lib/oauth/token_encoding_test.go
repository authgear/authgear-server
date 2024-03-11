package oauth

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/endpoints"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/rand"
	utilsecrets "github.com/authgear/authgear-server/pkg/util/secrets"
)

func TestAccessToken(t *testing.T) {
	Convey("EncodeAccessToken and DecodeAccessToken", t, func() {
		ctrl := gomock.NewController(t)

		now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

		secrets := &config.OAuthKeyMaterials{
			Set: jwk.NewSet(),
		}
		jwkKey := utilsecrets.GenerateRSAKey(now, rand.SecureRand)
		err := secrets.Set.AddKey(jwkKey)
		So(err, ShouldBeNil)

		mockUserClaimsProvider := NewMockUserClaimsProvider(ctrl)
		mockEventService := NewMockEventService(ctrl)

		mockUserClaimsProvider.EXPECT().PopulateNonPIIUserClaims(gomock.Any(), "user-id").DoAndReturn(
			func(token jwt.Token, userID string) error {
				return nil
			})
		mockEventService.EXPECT().DispatchEventOnCommit(gomock.Any()).Return(nil)

		encoding := &AccessTokenEncoding{
			Secrets:    secrets,
			Clock:      clock.NewMockClockAtTime(now),
			UserClaims: mockUserClaimsProvider,
			BaseURL: &endpoints.Endpoints{
				HTTPHost:  "test1.authgear.com",
				HTTPProto: "http",
			},
			Events: mockEventService,
		}

		client := &config.OAuthClientConfig{
			IssueJWTAccessToken: true,
			ClientID:            "client-id",
			AccessTokenLifetime: 3600,
		}

		accessGrant := &AccessGrant{
			CreatedAt: now,
			ExpireAt:  now.Add(client.AccessTokenLifetime.Duration()),
			TokenHash: "token-hash",
		}

		accessToken, err := encoding.EncodeAccessToken(client, accessGrant, "user-id", "token")
		So(err, ShouldBeNil)

		_, _, err = encoding.DecodeAccessToken(accessToken)
		So(err, ShouldBeNil)

		// Peek token payload
		keys, err := jwk.PublicSetOf(encoding.Secrets.Set)
		So(err, ShouldBeNil)

		decodedToken, _ := jwt.ParseString(accessToken, jwt.WithKeySet(keys), jwt.WithValidate(false))
		So(err, ShouldBeNil)

		clientID, _ := decodedToken.Get("client_id")
		idKey, _ := decodedToken.Get(jwt.JwtIDKey)

		So(decodedToken.Audience(), ShouldResemble, []string{"http://test1.authgear.com"})
		So(decodedToken.IssuedAt(), ShouldEqual, accessGrant.CreatedAt)
		So(decodedToken.Expiration(), ShouldEqual, accessGrant.ExpireAt)
		So(clientID, ShouldEqual, "client-id")
		So(idKey, ShouldEqual, "token-hash")
	})
}
