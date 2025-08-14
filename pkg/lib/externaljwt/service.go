package externaljwt

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
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
