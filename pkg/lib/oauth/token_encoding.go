package oauth

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type UserClaimsProvider interface {
	PopulateNonPIIUserClaims(token jwt.Token, userID string) error
}

type BaseURLProvider interface {
	BaseURL() *url.URL
}

type AccessTokenEncoding struct {
	Secrets    *config.OAuthKeyMaterials
	Clock      clock.Clock
	UserClaims UserClaimsProvider
	BaseURL    BaseURLProvider
}

func (e *AccessTokenEncoding) EncodeAccessToken(client *config.OAuthClientConfig, grant *AccessGrant, userID string, token string) (string, error) {
	if !client.IssueJWTAccessToken {
		return token, nil
	}

	claims := jwt.New()

	err := e.UserClaims.PopulateNonPIIUserClaims(claims, userID)
	if err != nil {
		return "", err
	}

	_ = claims.Set(jwt.AudienceKey, e.BaseURL.BaseURL().String())
	_ = claims.Set(jwt.IssuedAtKey, grant.CreatedAt.Unix())
	_ = claims.Set(jwt.ExpirationKey, grant.ExpireAt.Unix())
	_ = claims.Set("client_id", client.ClientID)
	// Do not put raw token in JWT access token; JWT payload is not specified
	// to be confidential. Put token hash to allow looking up access grant from
	// verified JWT.
	_ = claims.Set(jwt.JwtIDKey, grant.TokenHash)

	jwk, _ := e.Secrets.Set.Get(0)

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
		jwt.WithAudience(e.BaseURL.BaseURL().String()),
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
