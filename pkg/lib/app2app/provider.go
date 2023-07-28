package app2app

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type Provider struct {
	Clock clock.Clock
}

func (p *Provider) ParseTokenUnverified(requestJWT string) (t *Request, err error) {
	compact := []byte(requestJWT)

	hdr, jwtToken, err := jwtutil.SplitWithoutVerify(compact)
	if err != nil {
		err = fmt.Errorf("invalid app2app JWT: %w", err)
		return
	}

	err = jwt.Validate(jwtToken,
		jwt.WithClock(jwtClock{p.Clock}),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		err = fmt.Errorf("invalid app2app JWT: %w", err)
		return
	}

	var key jwk.Key
	var keyID string
	if jwkIface, ok := hdr.Get("jwk"); ok {
		var jwkBytes []byte
		jwkBytes, err = json.Marshal(jwkIface)
		if err != nil {
			err = fmt.Errorf("invalid app2app JWK: %w", err)
			return
		}

		var set jwk.Set
		set, err = jwk.Parse(jwkBytes)
		if err != nil {
			err = fmt.Errorf("invalid app2app JWK: %w", err)
			return
		}

		key, ok = set.Get(0)
		if !ok {
			err = fmt.Errorf("empty app2app JWK set")
			return
		}
		keyID = key.KeyID()

		// The client does include alg in the JWK.
		// Fix it by copying alg in the header.
		if key.Algorithm() == "" {
			_ = key.Set(jws.AlgorithmKey, hdr.Algorithm())
		}
	} else {
		err = errors.New("no app2app key provided")
		return
	}

	if !KeyIDFormat.MatchString(keyID) {
		err = errors.New("invalid app2app key ID format")
		return
	}

	typ := hdr.Type()
	if typ != RequestTokenType {
		err = errors.New("invalid app2app JWT type")
		return
	}

	token, err := jws.ParseString(requestJWT)
	if err != nil {
		err = fmt.Errorf("invalid app2app JWT: %w", err)
		return
	}

	var tokenPayload Request
	err = json.Unmarshal(token.Payload(), &tokenPayload)
	if err != nil {
		err = fmt.Errorf("invalid app2app JWT payload: %w", err)
		return
	}

	tokenPayload.Key = key
	t = &tokenPayload
	return
}

func (p *Provider) ParseToken(requestJWT string, key jwk.Key) (*Request, error) {

	set := jwk.NewSet()
	_ = set.Add(key)

	payload, err := jws.VerifySet([]byte(requestJWT), set)
	if err != nil {
		return nil, fmt.Errorf("invalid app2app JWT: %w", err)
	}

	var tokenPayload Request
	err = json.Unmarshal(payload, &tokenPayload)
	if err != nil {
		return nil, fmt.Errorf("invalid app2app JWT payload: %w", err)
	}

	tokenPayload.Key = key
	return &tokenPayload, nil
}
