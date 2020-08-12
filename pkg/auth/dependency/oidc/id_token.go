package oidc

import (
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/auth/model"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
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

	claims.Set(jwt.AudienceKey, client.ClientID())
	claims.Set(jwt.IssuedAtKey, now.Unix())
	claims.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
	for key, value := range session.AuthnAttrs().Claims {
		claims.Set(string(key), value)
	}
	if nonce != "" {
		claims.Set("nonce", nonce)
	}

	jwk := ti.Secrets.Set.Keys[0]

	signed, err := jwtutil.Sign(claims, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (ti *IDTokenIssuer) LoadUserClaims(session auth.AuthSession) (jwt.Token, error) {
	user, err := ti.Users.Get(session.AuthnAttrs().UserID)
	if err != nil {
		return nil, err
	}

	claims := jwt.New()
	claims.Set(jwt.IssuerKey, ti.Endpoints.BaseURL().String())
	claims.Set(jwt.SubjectKey, session.AuthnAttrs().UserID)
	claims.Set(string(authn.ClaimUserIsAnonymous), user.IsAnonymous)
	claims.Set(string(authn.ClaimUserIsVerified), user.IsVerified)

	return claims, nil
}
