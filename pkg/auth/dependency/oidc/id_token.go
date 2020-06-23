package oidc

import (
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/config"
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
	OIDCConfig config.OIDCConfiguration
	Endpoints  EndpointsProvider
	Users      UserProvider
	Clock      clock.Clock
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = 5 * time.Minute

func (ti *IDTokenIssuer) IssueIDToken(client config.OAuthClientConfiguration, session auth.AuthSession, nonce string) (string, error) {
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

	key := ti.OIDCConfig.Keys[0]
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = key.KID
	return token.SignedString(privKey)
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
