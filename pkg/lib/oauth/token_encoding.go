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
	PrepareBlockingEventWithTx(ctx context.Context, payload event.BlockingPayload) (e *event.Event, err error)
	DispatchEventWithoutTx(ctx context.Context, e *event.Event) (err error)
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

type PrepareUserAccessTokenResult interface {
	prepareUserAccessTokenResult()
}

type prepareUserAccessTokenResultOpaque struct {
	OriginalToken string
	ClientConfig  *config.OAuthClientConfig
}

func (r *prepareUserAccessTokenResultOpaque) prepareUserAccessTokenResult() {}

type prepareUserAccessTokenResultJWT struct {
	Event        *event.Event
	ForBackup    map[string]interface{}
	ClientConfig *config.OAuthClientConfig
}

func (r *prepareUserAccessTokenResultJWT) prepareUserAccessTokenResult() {}

func (e *AccessTokenEncoding) PrepareUserAccessToken(ctx context.Context, options EncodeUserAccessTokenOptions) (PrepareUserAccessTokenResult, error) {
	if !options.ClientConfig.IssueJWTAccessToken {
		return &prepareUserAccessTokenResultOpaque{
			OriginalToken: options.OriginalToken,
			ClientConfig:  options.ClientConfig,
		}, nil
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
	// scope
	_ = claims.Set("scope", strings.Join(options.AccessGrant.Scopes, " "))

	// auth_time
	_ = claims.Set(string(model.ClaimAuthTime), options.AuthenticationInfo.AuthenticatedAt.Unix())

	// amr
	if amr := options.AuthenticationInfo.AMR; len(amr) > 0 {
		_ = claims.Set(string(model.ClaimAMR), amr)
	}

	// jti
	// Do not put raw token in JWT access token; JWT payload is not specified
	// to be confidential. Put token hash to allow looking up access grant from
	// verified JWT.
	_ = claims.Set(jwt.JwtIDKey, options.AccessGrant.TokenHash)

	err := e.IDTokenIssuer.PopulateUserClaimsInIDToken(ctx, claims, options.AuthenticationInfo.UserID, options.ClientLike)
	if err != nil {
		return nil, err
	}

	forMutation, forBackup, err := jwtutil.PrepareForMutations(claims)
	if err != nil {
		return nil, err
	}

	identities, err := e.Identities.ListIdentitiesThatHaveStandardAttributes(ctx, options.AuthenticationInfo.UserID)
	if err != nil {
		return nil, err
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

	event, err := e.Events.PrepareBlockingEventWithTx(ctx, eventPayload)
	if err != nil {
		return nil, err
	}

	return &prepareUserAccessTokenResultJWT{
		Event:        event,
		ForBackup:    forBackup,
		ClientConfig: options.ClientConfig,
	}, nil
}

type MakeUserAccessTokenFromPreparationOptions struct {
	PreparationResult PrepareUserAccessTokenResult
}

func (e *AccessTokenEncoding) MakeUserAccessTokenFromPreparationResult(
	ctx context.Context,
	options MakeUserAccessTokenFromPreparationOptions,
) (*IssueAccessGrantResult, error) {
	if options.PreparationResult == nil {
		panic(fmt.Errorf("options.PreparationResult must be non-nil"))
	}

	switch v := options.PreparationResult.(type) {
	case *prepareUserAccessTokenResultOpaque:
		return &IssueAccessGrantResult{
			Token:     v.OriginalToken,
			TokenType: "Bearer",
			ExpiresIn: int(v.ClientConfig.AccessTokenLifetime),
		}, nil
	case *prepareUserAccessTokenResultJWT:
		err := e.Events.DispatchEventWithoutTx(ctx, v.Event)
		if err != nil {
			return nil, err
		}

		eventPayload := v.Event.Payload.(*blocking.OIDCJWTPreCreateBlockingEventPayload)

		claims, err := jwtutil.ApplyMutations(
			eventPayload.JWT.Payload,
			v.ForBackup,
		)
		if err != nil {
			return nil, err
		}

		jwk, _ := e.Secrets.Set.Key(0)

		hdr := jws.NewHeaders()
		_ = hdr.Set("typ", "at+jwt")

		signed, err := jwtutil.SignWithHeader(claims, hdr, jwa.RS256, jwk)
		if err != nil {
			return nil, err
		}

		return &IssueAccessGrantResult{
			Token:     string(signed),
			TokenType: "Bearer",
			ExpiresIn: int(v.ClientConfig.AccessTokenLifetime),
		}, nil
	default:
		panic(fmt.Errorf("unexpected PreparationResult: %T", options.PreparationResult))
	}
}

func (e *AccessTokenEncoding) EncodeClientAccessToken(ctx context.Context, options EncodeClientAccessTokenOptions) (string, error) {
	if !options.ClientConfig.IssueJWTAccessToken {
		return options.OriginalToken, nil
	}

	claims := jwt.New()

	// jti
	_ = claims.Set(jwt.JwtIDKey, HashToken(options.OriginalToken))
	// iss
	_ = claims.Set(jwt.IssuerKey, e.IDTokenIssuer.Iss())
	// aud
	_ = claims.Set(jwt.AudienceKey, options.ResourceURI)
	// iat
	_ = claims.Set(jwt.IssuedAtKey, options.CreatedAt.Unix())
	// exp
	_ = claims.Set(jwt.ExpirationKey, options.ExpireAt.Unix())
	// client_id
	_ = claims.Set("client_id", options.ClientConfig.ClientID)
	// scope
	_ = claims.Set("scope", options.Scope)
	// sub
	_ = claims.Set("sub", fmt.Sprintf("client_id_%s", options.ClientConfig.ClientID))

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
