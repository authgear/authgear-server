package service

import (
	"encoding/json"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

type TokenService struct {
	Clock clock.Clock
}

func (t *TokenService) GenerateShortLivedAdminAPIToken(
	appID string,
	keyID string,
	privateKeyPEM string) (string, error) {

	jwkSet, err := jwk.Parse([]byte(privateKeyPEM), jwk.WithPEM(true))
	if err != nil {
		return "", err
	}

	key, _ := jwkSet.Key(0)
	_ = key.Set("kid", keyID)

	now := t.Clock.NowUTC()
	payload := jwt.New()
	_ = payload.Set(jwt.AudienceKey, appID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())

	// The alg MUST be RS256.
	alg := jwa.RS256
	hdr := jws.NewHeaders()
	_ = hdr.Set("typ", "JWT")

	buf, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	token, err := jws.Sign(buf, jws.WithKey(alg, key, jws.WithProtectedHeaders(hdr)))
	if err != nil {
		return "", err
	}

	return string(token), nil
}
