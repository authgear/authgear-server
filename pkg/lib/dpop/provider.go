package dpop

import (
	"crypto"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type Provider struct {
	Clock     clock.Clock
	HTTPHost  httputil.HTTPHost
	HTTPProto httputil.HTTPProto
}

const (
	DPoPJWTTyp = "dpop+jwt"
)

func (p *Provider) ParseProof(jwtStr string) (*DPoPProof, error) {
	jwtBytes := []byte(jwtStr)

	hdr, payload, err := jwtutil.SplitWithoutVerify(jwtBytes)
	if err != nil {
		return nil, ErrMalformedJwt
	}

	jwk, proof, err := p.validateProofJWT(hdr, payload)
	if err != nil {
		return nil, err
	}

	_, err = jws.Verify(jwtBytes, jws.WithKey(hdr.Algorithm(), jwk))
	if err != nil {
		return nil, ErrInvalidJwtSignature
	}

	return proof, nil
}

func (p *Provider) validateProofJWT(header jws.Headers, payload jwt.Token) (jwk.Key, *DPoPProof, error) {
	now := p.Clock.NowUTC()
	err := jwt.Validate(payload,
		jwt.WithClock(jwtClock{p.Clock}),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		return nil, nil, ErrInvalidJwt
	}

	// Do not accept a proof issued a long time ago
	// https://datatracker.ietf.org/doc/html/rfc9449#section-11.1
	if payload.IssuedAt().Add(duration.Short).Before(now) {
		return nil, nil, ErrProofExpired
	}

	var key jwk.Key
	if jwkIface, ok := header.Get("jwk"); ok {
		var jwkBytes []byte
		jwkBytes, err = json.Marshal(jwkIface)
		if err != nil {
			return nil, nil, ErrInvalidJwk
		}

		var set jwk.Set
		set, err = jwk.Parse(jwkBytes)
		if err != nil {
			return nil, nil, ErrInvalidJwk
		}

		key, ok = set.Key(0)
		if !ok {
			return nil, nil, ErrInvalidJwk
		}
	} else {
		return nil, nil, ErrInvalidJwk
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

	if header.Type() != DPoPJWTTyp {
		return nil, nil, ErrInvalidJwtType
	}

	jti, ok := getPayloadAsString("jti")
	if !ok {
		return nil, nil, ErrInvalidJwtPayload
	}

	htm, ok := getPayloadAsString("htm")
	if !ok {
		return nil, nil, ErrInvalidJwtPayload
	}

	htu, ok := getPayloadAsString("htu")
	if !ok {
		return nil, nil, ErrInvalidJwtPayload
	}

	htuURI, err := url.Parse(htu)
	if err != nil {
		return nil, nil, ErrInvalidHTU
	}

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		panic(err)
	}

	jkt := base64.RawURLEncoding.EncodeToString(thumbprint)

	return key, &DPoPProof{
		JTI: jti,
		HTM: htm,
		HTU: htuURI,
		JKT: jkt,
	}, nil
}

func (p *Provider) CompareHTU(proof *DPoPProof, req *http.Request) error {
	// req.URL does not have scheme and host, compare using HTTPProto and HTTPHost
	if proof.HTU.Scheme != string(p.HTTPProto) {
		return ErrUnmatchedURI
	}
	if proof.HTU.Host != string(p.HTTPHost) {
		return ErrUnmatchedURI
	}
	if proof.HTU.Path != req.URL.Path {
		return ErrUnmatchedURI
	}
	return nil
}

func (p *Provider) CompareHTM(proof *DPoPProof, requestMethod string) error {
	if strings.ToLower(proof.HTM) != strings.ToLower(requestMethod) {
		return ErrUnmatchedMethod
	}
	return nil
}
