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
)

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type UserClaims struct {
	jwt.StandardClaims
	User      *model.User `json:"skygear_user,omitempty"`
	SessionID string      `json:"skygear_session_id,omitempty"`
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
	keys, err := ti.Secrets.Decode()
	if err != nil {
		return nil, err
	}

	jwks := &jwk.Set{}
	for _, key := range keys.Keys {
		k, err := key.Materialize()
		if err != nil {
			return nil, err
		}
		k, err = jwk.GetPublicKey(k)
		if err != nil {
			return nil, err
		}
		key, err = jwk.New(k)
		if err != nil {
			return nil, err
		}

		jwks.Keys = append(jwks.Keys, key)
	}
	return jwks, nil
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

	keys, err := ti.Secrets.Decode()
	if err != nil {
		panic("oidc: invalid key materials: " + err.Error())
	}

	jwk := keys.Keys[0]
	key, err := jwk.Materialize()
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
			Issuer:  ti.Endpoints.BaseURL().String(),
			Subject: session.AuthnAttrs().UserID,
		},
	}

	if !allowProfile {
		return claims, nil
	}

	user, err := ti.Users.Get(session.AuthnAttrs().UserID)
	if err != nil {
		return nil, err
	}

	claims.User = user
	claims.SessionID = session.SessionID()

	return claims, nil
}
