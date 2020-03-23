package oidc

import (
	gotime "time"

	"github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type IDToken struct {
	jwt.StandardClaims
	Nonce string `json:"nonce,omitempty"`
}

type IDTokenIssuer struct {
	OIDCConfig config.OIDCConfiguration
	URLPrefix  urlprefix.Provider
	Time       time.Provider
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = 5 * gotime.Minute

func (ti *IDTokenIssuer) IssueIDToken(client config.OAuthClientConfiguration, userID string, nonce string) (string, error) {
	now := ti.Time.NowUTC()
	token := &IDToken{
		StandardClaims: jwt.StandardClaims{
			Issuer:    ti.URLPrefix.Value().String(),
			Audience:  client.ClientID(),
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(IDTokenValidDuration).Unix(),
			Subject:   userID,
		},
		Nonce: nonce,
	}

	key := ti.OIDCConfig.Keys[0]
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return "", err
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodRS256, token)
	jwt.Header["kid"] = key.KID
	return jwt.SignedString(privKey)
}
