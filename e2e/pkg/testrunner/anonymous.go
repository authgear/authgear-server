package testrunner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"sync"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// anonymousRequestTokenType mirrors anonymous.RequestTokenType.
const anonymousRequestTokenType = "vnd.authgear.anonymous-request"

// anonymousKeyID must satisfy anonymous.KeyIDFormat.
const anonymousKeyID = "e2e-anonymous-key"

// The anonymous key identifies the anonymous user the server creates from it,
// so one key for the whole process keeps the same user across requests.
var (
	anonymousKeyOnce sync.Once
	anonymousKey     jwk.Key
	anonymousPubKey  jwk.Key
)

func getAnonymousKey() (private jwk.Key, public jwk.Key) {
	anonymousKeyOnce.Do(func() {
		raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		anonymousKey, err = jwk.FromRaw(raw)
		if err != nil {
			panic(err)
		}
		anonymousPubKey, err = anonymousKey.PublicKey()
		if err != nil {
			panic(err)
		}
		_ = anonymousPubKey.Set(jwk.AlgorithmKey, jwa.ES256)
		_ = anonymousPubKey.Set(jwk.KeyIDKey, anonymousKeyID)
	})
	return anonymousKey, anonymousPubKey
}

// GenerateAnonymousRequestJWT returns an anonymous request JWT carrying the
// challenge, with the public key in the header as the SDKs send it.
func GenerateAnonymousRequestJWT(challenge string) (string, error) {
	privateKey, publicKey := getAnonymousKey()

	header := jws.NewHeaders()
	if err := header.Set("typ", anonymousRequestTokenType); err != nil {
		return "", err
	}
	if err := header.Set("alg", "ES256"); err != nil {
		return "", err
	}
	if err := header.Set("kid", anonymousKeyID); err != nil {
		return "", err
	}
	if err := header.Set("jwk", publicKey); err != nil {
		return "", err
	}

	payload := jwt.New()
	if err := payload.Set("challenge", challenge); err != nil {
		return "", err
	}
	if err := payload.Set("action", "auth"); err != nil {
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
