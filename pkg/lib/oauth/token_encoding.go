package oauth

import (
	"context"
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
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

//go:generate go tool mockgen -source=token_encoding.go -destination=token_encoding_mock_test.go -package oauth

type IDTokenIssuer interface {
	Iss() string
	PopulateUserClaimsInIDToken(ctx context.Context, token jwt.Token, userID string, clientLike *ClientLike) error
}

type BaseURLProvider interface {
	Origin() *url.URL
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type AccessTokenEncodingIdentityService interface {
	ListIdentitiesThatHaveStandardAttributes(ctx context.Context, userID string) ([]*identity.Info, error)
}

type AccessTokenEncoding struct {
	Secrets       *config.OAuthKeyMaterials
	Clock         clock.Clock
	IDTokenIssuer IDTokenIssuer
	BaseURL       BaseURLProvider
	Events        EventService
	Identities    AccessTokenEncodingIdentityService
}

type EncodeUserAccessTokenOptions struct {
	OriginalToken      string
	ClientConfig       *config.OAuthClientConfig
	ClientLike         *ClientLike
	AccessGrant        *AccessGrant
	AuthenticationInfo authenticationinfo.T
}

type EncodeClientAccessTokenOptions struct {
	OriginalToken string
	ClientConfig  *config.OAuthClientConfig
	ResourceURI   string
	Scope         string
	CreatedAt     time.Time
	ExpireAt      time.Time
}

func (e *AccessTokenEncoding) EncodeUserAccessToken(ctx context.Context, options EncodeUserAccessTokenOptions) (string, error) {
	if !options.ClientConfig.IssueJWTAccessToken {
		return options.OriginalToken, nil
	}

	claims := jwt.New()

	// iss
	_ = claims.Set(jwt.IssuerKey, e.IDTokenIssuer.Iss())
	// aud
	_ = claims.Set(jwt.AudienceKey, e.BaseURL.Origin().String())
	// iat
	_ = claims.Set(jwt.IssuedAtKey, options.AccessGrant.CreatedAt.Unix())
	// exp
	_ = claims.Set(jwt.ExpirationKey, options.AccessGrant.ExpireAt.Unix())
	// client_id
	_ = claims.Set("client_id", options.ClientConfig.ClientID)

	// auth_time
	_ = claims.Set(string(model.ClaimAuthTime), options.AuthenticationInfo.AuthenticatedAt.Unix())

	// amr
	if amr := options.AuthenticationInfo.AMR; len(amr) > 0 {
		_ = claims.Set(string(model.ClaimAMR), amr)
	}

	// Do not put raw token in JWT access token; JWT payload is not specified
	// to be confidential. Put token hash to allow looking up access grant from
	// verified JWT.
	_ = claims.Set(jwt.JwtIDKey, options.AccessGrant.TokenHash)

	err := e.IDTokenIssuer.PopulateUserClaimsInIDToken(ctx, claims, options.AuthenticationInfo.UserID, options.ClientLike)
	if err != nil {
		return "", err
	}

	forMutation, forBackup, err := jwtutil.PrepareForMutations(claims)
	if err != nil {
		return "", err
	}

	identities, err := e.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, options.AuthenticationInfo.UserID)
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
				ID: options.AuthenticationInfo.UserID,
			},
		},
		Identities: identityModels,
		JWT: blocking.OIDCJWT{
			Payload: forMutation,
		},
	}

	err = e.Events.DispatchEventOnCommit(ctx, eventPayload)
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

func (e *AccessTokenEncoding) EncodeClientAccessToken(ctx context.Context, options EncodeClientAccessTokenOptions) (string, error) {
	if !options.ClientConfig.IssueJWTAccessToken {
		return options.OriginalToken, nil
	}

	claims := jwt.New()

	// iss
	_ = claims.Set(jwt.IssuerKey, e.IDTokenIssuer.Iss())
	// aud
	_ = claims.Set(jwt.AudienceKey, options.ResourceURI)
	// iat
	_ = claims.Set(jwt.IssuedAtKey, options.CreatedAt.Unix())
	// exp
	_ = claims.Set(jwt.ExpirationKey, options.ExpireAt.Unix())
	// client_id
	_ = claims.Set("client_id", fmt.Sprint("client_id_%s", options.ClientConfig.ClientID))
	// scope
	_ = claims.Set("scope", options.Scope)

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
