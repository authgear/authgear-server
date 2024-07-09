package dpop

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type Provider struct {
	Clock clock.Clock
}

func (p *Provider) ParseProof(jwtStr string) (*DPoPProof, error) {
	jwtBytes := []byte(jwtStr)
	now := p.Clock.NowUTC()

	hdr, payload, err := jwtutil.SplitWithoutVerify(jwtBytes)
	if err != nil {
		return nil, ErrMalformedJwt
	}

	err = jwt.Validate(payload,
		jwt.WithClock(jwtClock{p.Clock}),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		return nil, ErrInvalidJwt
	}

	// Do not accept a proof issued a long time ago
	// https://datatracker.ietf.org/doc/html/rfc9449#section-11.1
	if payload.IssuedAt().Add(duration.Short).Before(now) {
		return nil, ErrProofExpired
	}

	var key jwk.Key
	if jwkIface, ok := hdr.Get("jwk"); ok {
		var jwkBytes []byte
		jwkBytes, err = json.Marshal(jwkIface)
		if err != nil {
			return nil, ErrInvalidJwk
		}

		var set jwk.Set
		set, err = jwk.Parse(jwkBytes)
		if err != nil {
			return nil, ErrInvalidJwk
		}

		key, ok = set.Key(0)
		if !ok {
			return nil, ErrInvalidJwk
		}
	} else {
		return nil, ErrInvalidJwtNoJwkProvided
	}

	// Verify the signature
	set := jwk.NewSet()
	_ = set.AddKey(key)
	_, err = jws.Verify(jwtBytes, jws.WithKeySet(set))
	if err != nil {
		return nil, ErrInvalidJwtSignature
	}

	getPayloadAsString := func(key string) (string, bool) {
		valInterface, ok := payload.Get(key)
		if !ok {
			return "", false
		}
		valStr, ok := valInterface.(string)
		if !ok {
			return "", false
		}
		return valStr, true
	}

	jti, ok := getPayloadAsString("jti")
	if !ok {
		return nil, ErrInvalidJwtPayload
	}

	htm, ok := getPayloadAsString("htm")
	if !ok {
		return nil, ErrInvalidJwtPayload
	}

	htu, ok := getPayloadAsString("htu")
	if !ok {
		return nil, ErrInvalidJwtPayload
	}

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		panic(err)
	}

	jkt := base64.RawURLEncoding.EncodeToString(thumbprint)

	return &DPoPProof{
		JTI: jti,
		HTM: htm,
		HTU: htu,
		JKT: jkt,
	}, nil
}
