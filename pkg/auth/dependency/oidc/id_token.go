package oidc

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/jwkutil"
)

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type UserClaims struct {
	jwt.StandardClaims
}

type IDTokenClaims struct {
	UserClaims
	Nonce string `json:"nonce,omitempty"`
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
	userClaims, err := ti.LoadUserClaims(session)
	if err != nil {
		return "", err
	}

	now := ti.Clock.NowUTC()
	userClaims.StandardClaims.Audience = client.ClientID()
	userClaims.StandardClaims.IssuedAt = now.Unix()
	userClaims.StandardClaims.ExpiresAt = now.Add(IDTokenValidDuration).Unix()

	claims := &IDTokenClaims{
		UserClaims: *userClaims,
		Nonce:      nonce,
	}

	jwk := ti.Secrets.Set.Keys[0]
	var key interface{}
	err = jwk.Raw(&key)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = jwk.KeyID()
	return token.SignedString(key)
}

func (ti *IDTokenIssuer) LoadUserClaims(session auth.AuthSession) (*UserClaims, error) {
	allowProfile := false
	for _, scope := range oauth.SessionScopes(session) {
		if scope == oauth.FullAccessScope {
			allowProfile = true
		}
	}

	claims := &UserClaims{
		StandardClaims: jwt.StandardClaims{
			// TODO(id-token): https://openid.net/specs/openid-connect-core-1_0.html#IDToken
			// Set `aud` to `client_id`.
			// Set `exp` to the expiration time.
			// Set `iat` to NowUTC().
			// Set `acr` to session.ACR.
			// Set `amr` to session.AMR.
			// Define a custom claim to indicate anonymous.
			Issuer:  ti.Endpoints.BaseURL().String(),
			Subject: session.AuthnAttrs().UserID,
		},
	}

	if !allowProfile {
		return claims, nil
	}

	return claims, nil
}
