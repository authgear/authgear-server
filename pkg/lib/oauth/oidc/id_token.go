package oidc

import (
	"net/url"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type BaseURLProvider interface {
	BaseURL() *url.URL
}

type IDTokenIssuer struct {
	Secrets *config.OAuthKeyMaterials
	BaseURL BaseURLProvider
	Users   UserProvider
	Clock   clock.Clock
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = duration.Short

func (ti *IDTokenIssuer) GetPublicKeySet() (jwk.Set, error) {
	return jwk.PublicSetOf(ti.Secrets.Set)
}

func (ti *IDTokenIssuer) IssueIDToken(client *config.OAuthClientConfig, s session.Session, nonce string) (string, error) {
	claims, err := ti.LoadUserClaims(s.SessionAttrs().UserID)
	if err != nil {
		return "", err
	}

	now := ti.Clock.NowUTC()

	_ = claims.Set(jwt.AudienceKey, client.ClientID)
	_ = claims.Set(jwt.IssuedAtKey, now.Unix())
	_ = claims.Set(jwt.ExpirationKey, now.Add(IDTokenValidDuration).Unix())
	for key, value := range s.SessionAttrs().Claims {
		_ = claims.Set(string(key), value)
	}
	if nonce != "" {
		_ = claims.Set("nonce", nonce)
	}

	jwk, _ := ti.Secrets.Set.Get(0)

	signed, err := jwtutil.Sign(claims, jwa.RS256, jwk)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

func (ti *IDTokenIssuer) LoadUserClaims(userID string) (jwt.Token, error) {
	user, err := ti.Users.Get(userID)
	if err != nil {
		return nil, err
	}

	claims := jwt.New()
	_ = claims.Set(jwt.IssuerKey, ti.BaseURL.BaseURL().String())
	_ = claims.Set(jwt.SubjectKey, userID)
	_ = claims.Set(string(authn.ClaimUserIsAnonymous), user.IsAnonymous)
	_ = claims.Set(string(authn.ClaimUserIsVerified), user.IsVerified)

	return claims, nil
}
