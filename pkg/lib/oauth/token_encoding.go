package oauth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

//go:generate mockgen -source=token_encoding.go -destination=token_encoding_mock_test.go -package oauth

type IDTokenIssuer interface {
	Iss() string
	PopulateUserClaimsInIDToken(token jwt.Token, userID string, clientLike *ClientLike) error
}

type BaseURLProvider interface {
	Origin() *url.URL
}

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type AccessTokenEncodingIdentityService interface {
	ListIdentitiesThatHaveStandardAttributes(userID string) ([]*identity.Info, error)
}

type AccessTokenEncoding struct {
	Secrets       *config.OAuthKeyMaterials
	Clock         clock.Clock
	IDTokenIssuer IDTokenIssuer
	BaseURL       BaseURLProvider
	Events        EventService
	Identities    AccessTokenEncodingIdentityService
}

func (e *AccessTokenEncoding) EncodeAccessToken(client *config.OAuthClientConfig, clientLike *ClientLike, grant *AccessGrant, userID string, token string) (string, error) {
	if !client.IssueJWTAccessToken {
		return token, nil
	}

	claims := jwt.New()

	err := e.IDTokenIssuer.PopulateUserClaimsInIDToken(claims, userID, clientLike)
	if err != nil {
		return "", err
	}

	_ = claims.Set(jwt.IssuerKey, e.IDTokenIssuer.Iss())
	_ = claims.Set(jwt.AudienceKey, e.BaseURL.Origin().String())
	_ = claims.Set(jwt.IssuedAtKey, grant.CreatedAt.Unix())
	_ = claims.Set(jwt.ExpirationKey, grant.ExpireAt.Unix())
	_ = claims.Set("client_id", client.ClientID)
	// Do not put raw token in JWT access token; JWT payload is not specified
	// to be confidential. Put token hash to allow looking up access grant from
	// verified JWT.
	_ = claims.Set(jwt.JwtIDKey, grant.TokenHash)

	forMutation, forBackup, err := jwtutil.PrepareForMutations(claims)
	if err != nil {
		return "", err
	}

	identities, err := e.Identities.ListIdentitiesThatHaveStandardAttributes(userID)
	if err != nil {
		return "", err
	}

	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}

	eventPayload := &blocking.OIDCJWTPreCreateBlockingEventPayload{
		UserRef: model.UserRef{
			Meta: model.Meta{
				ID: userID,
			},
		},
		Identities: identityModels,
		JWT: blocking.OIDCJWT{
			Payload: forMutation,
		},
	}

	err = e.Events.DispatchEventOnCommit(eventPayload)
	if err != nil {
		return "", err
	}

	claims, err = jwtutil.ApplyMutations(
		eventPayload.JWT.Payload,
		forBackup,
	)
	if err != nil {
		return "", err
	}

	jwk, _ := e.Secrets.Set.Key(0)

	hdr := jws.NewHeaders()
	_ = hdr.Set("typ", "at+jwt")

	signed, err := jwtutil.SignWithHeader(claims, hdr, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (e *AccessTokenEncoding) DecodeAccessToken(encodedToken string) (tok string, isHash bool, err error) {
	// Check for JWT common prefix.
	if !strings.HasPrefix(encodedToken, "eyJ") {
		return encodedToken, false, nil
	}

	keys, err := jwk.PublicSetOf(e.Secrets.Set)
	if err != nil {
		return "", false, err
	}

	token, err := jwt.ParseString(encodedToken, jwt.WithKeySet(keys))
	if err != nil {
		// Invalid JWT string: assume opaque tokens.
		return encodedToken, false, nil
	}

	err = jwt.Validate(token,
		jwt.WithClock(&jwtClock{e.Clock}),
		jwt.WithAudience(e.BaseURL.Origin().String()),
	)
	if err != nil {
		return "", false, err
	}

	return token.JwtID(), true, nil
}

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

func EncodeRefreshToken(token string, grantID string) string {
	return fmt.Sprintf("%s.%s", grantID, token)
}

func DecodeRefreshToken(encodedToken string) (token string, grantID string, err error) {
	parts := strings.SplitN(encodedToken, ".", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid refresh token")
	}

	return parts[1], parts[0], nil
}
