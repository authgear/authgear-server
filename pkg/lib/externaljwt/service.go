package externaljwt

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
)

type jwtClock struct {
	clock.Clock
}

func (j *jwtClock) Now() time.Time {
	return j.Clock.NowUTC()
}

type Service struct {
	ExternalJWTConfig *config.ExternalJWTConfig
	JWKSCache         *jwk.Cache
	Clock             clock.Clock
}

func (s *Service) VerifyExternalJWT(ctx context.Context, rawToken string) (jwt.Token, error) {
	token, err := jwt.ParseString(rawToken, jwt.WithValidate(false))
	if err != nil {
		return nil, ErrInvalidExternalJWT.New("failed to parse JWT")
	}

	issuer := token.Issuer()
	if issuer == "" {
		return nil, ErrInvalidExternalJWT.New("JWT is missing issuer (iss) claim")
	}

	var issuerConfig *config.ExternalJWTIssuerConfig
	for _, c := range s.ExternalJWTConfig.Issuers {
		c := c
		if c.Iss == issuer {
			issuerConfig = &c
			break
		}
	}

	if issuerConfig == nil {
		return nil, ErrInvalidExternalJWT.New("unknown external jwt issuer")
	}

	jwksURI := issuerConfig.JWKSURI

	// Register JWKS URI with the cache
	err = s.JWKSCache.Register(jwksURI)
	if err != nil {
		return nil, err
	}

	// Fetch JWKS from cache
	keySet, err := s.JWKSCache.Get(ctx, jwksURI)
	if err != nil {
		return nil, ErrFailedToFetchJWKS.New("failed to fetch JWKS from " + jwksURI)
	}

	// Verify the token
	// We already parsed the token once to get the issuer, now parse again with verification
	verifiedToken, err := jwt.ParseString(
		rawToken,
		jwt.WithKeySet(keySet),
		jwt.WithIssuer(issuer),
		jwt.WithAudience(issuerConfig.Aud),
		jwt.WithClock(&jwtClock{s.Clock}),
		jwt.WithAcceptableSkew(5*time.Minute), // Allow for clock skew
	)
	if err != nil {
		return nil, ErrInvalidExternalJWT.New("failed to verify jwt signature")
	}

	return verifiedToken, nil
}

func (s *Service) ConstructLoginIDSpec(
	identification model.AuthenticationFlowIdentification,
	token jwt.Token,
) (*identity.Spec, error) {
	var claimValue string
	var loginIDKey string
	var loginIDKeyType model.LoginIDKeyType

	switch identification {
	case model.AuthenticationFlowIdentificationEmail:
		email, ok := token.Get(string(stdattrs.Email))
		if !ok {
			return nil, ErrInvalidJWTClaim.New("email claim not found in JWT")
		}
		claimValue = email.(string)
		loginIDKey = string(model.LoginIDKeyTypeEmail)
		loginIDKeyType = model.LoginIDKeyTypeEmail
	case model.AuthenticationFlowIdentificationPhone:
		phoneNumber, ok := token.Get(string(stdattrs.PhoneNumber))
		if !ok {
			return nil, ErrInvalidJWTClaim.New("phone_number claim not found in JWT")
		}
		claimValue = phoneNumber.(string)
		loginIDKey = string(model.LoginIDKeyTypePhone)
		loginIDKeyType = model.LoginIDKeyTypePhone
	case model.AuthenticationFlowIdentificationUsername:
		username, ok := token.Get(string(stdattrs.PreferredUsername))
		if !ok {
			return nil, ErrInvalidJWTClaim.New("preferred_username claim not found in JWT")
		}
		claimValue = username.(string)
		loginIDKey = string(model.LoginIDKeyTypeUsername)
		loginIDKeyType = model.LoginIDKeyTypeUsername
	default:
		// This should not happen.
		panic("unexpected identification method: " + identification)
	}

	spec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Key:   loginIDKey,
			Type:  loginIDKeyType,
			Value: stringutil.NewUserInputString(claimValue),
		},
	}
	return spec, nil
}
