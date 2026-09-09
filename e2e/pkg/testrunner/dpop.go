package testrunner

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// The DPoP proof is signed by a key the client holds, which the server learns
// from the proof itself, so a key generated here is all a test needs. One key
// for the whole process keeps the thumbprint stable, which is what lets a test
// bind a session to it and then present further proofs for that session.
var (
	dpopKeyOnce sync.Once
	dpopKey     jwk.Key
	dpopPubKey  jwk.Key
)

func getDPoPKey() (private jwk.Key, public jwk.Key) {
	dpopKeyOnce.Do(func() {
		raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		dpopKey, err = jwk.FromRaw(raw)
		if err != nil {
			panic(err)
		}
		dpopPubKey, err = dpopKey.PublicKey()
		if err != nil {
			panic(err)
		}
		_ = dpopPubKey.Set(jwk.AlgorithmKey, jwa.ES256)
	})
	return dpopKey, dpopPubKey
}

// GenerateDPoPProof returns a DPoP proof JWT bound to the given method and URI.
//
// jti identifies the proof. Two proofs generated with the same jti are distinct
// JWTs, since ECDSA signing is randomised, but the server must still treat the
// second as a replay: RFC 9449 section 11.1 deduplicates on jti, not on bytes.
func GenerateDPoPProof(htm string, htu string, jti string) (string, error) {
	privateKey, publicKey := getDPoPKey()

	header := jws.NewHeaders()
	if err := header.Set("typ", "dpop+jwt"); err != nil {
		return "", err
	}
	if err := header.Set("alg", "ES256"); err != nil {
		return "", err
	}
	if err := header.Set("jwk", publicKey); err != nil {
		return "", err
	}

	payload := jwt.New()
	if err := payload.Set("jti", jti); err != nil {
		return "", err
	}
	if err := payload.Set("htm", htm); err != nil {
		return "", err
	}
	if err := payload.Set("htu", htu); err != nil {
		return "", err
	}
	if err := payload.Set("iat", time.Now().UTC().Unix()); err != nil {
		return "", err
	}

	payloadBytes, err := jwt.NewSerializer().Serialize(payload)
	if err != nil {
		return "", err
	}

	signed, err := jws.Sign(payloadBytes, jws.WithKey(jwa.ES256, privateKey, jws.WithProtectedHeaders(header)))
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// GenerateDPoPJKT returns the thumbprint of the key used by GenerateDPoPProof,
// for the dpop_jkt authorization request parameter.
func GenerateDPoPJKT() (string, error) {
	_, publicKey := getDPoPKey()
	thumbprint, err := publicKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", fmt.Errorf("failed to compute jkt: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(thumbprint), nil
}
