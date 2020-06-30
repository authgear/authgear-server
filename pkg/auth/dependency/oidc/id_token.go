package oidc

import (
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/dependency/oauth"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/clock"
	"github.com/authgear/authgear-server/pkg/jwkutil"
	"github.com/authgear/authgear-server/pkg/jwtutil"
)

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type IDTokenIssuer struct {
	Secrets   *config.OIDCKeyMaterials
	Endpoints EndpointsProvider
	Users     UserProvider
	Clock     clock.Clock
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = 5 * time.Minute

func (ti *IDTokenIssuer) GetPublicKeySet() (*jwk.Set, error) {
	return jwkutil.PublicKeySet(&ti.Secrets.Set)
}

func (ti *IDTokenIssuer) IssueIDToken(client config.OAuthClientConfig, session auth.AuthSession, nonce string) (string, error) {
	claims, err := ti.LoadUserClaims(session)
	if err != nil {
		return "", err
	}

	now := ti.Clock.NowUTC()

	// TODO(id-token): https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	// Set `aud` to `client_id`.
	// Set `acr` to session.ACR.
	// Set `amr` to session.AMR.
	claims.Set(jwt.AudienceKey, client.ClientID())
	claims.Set(jwt.IssuedAtKey, now.Unix())
	claims.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
	claims.Set("nonce", nonce)

	jwk := ti.Secrets.Set.Keys[0]

	signed, err := jwtutil.Sign(claims, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (ti *IDTokenIssuer) LoadUserClaims(session auth.AuthSession) (jwt.Token, error) {
	allowProfile := false
	for _, scope := range oauth.SessionScopes(session) {
		if scope == oauth.FullAccessScope {
			allowProfile = true
		}
	}

	// TODO(user-info): https://openid.net/specs/openid-connect-core-1_0.html#IDToken
	// Set `exp` to the expiration time.
	// Set `iat` to NowUTC().
	// Define a custom claim to indicate anonymous.
	claims := jwt.New()
	claims.Set(jwt.IssuerKey, ti.Endpoints.BaseURL().String())
	claims.Set(jwt.SubjectKey, session.AuthnAttrs().UserID)

	if !allowProfile {
		return claims, nil
	}

	return claims, nil
}
