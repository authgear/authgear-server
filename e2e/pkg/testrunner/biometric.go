package testrunner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// biometricRequestTokenType mirrors biometric.RequestTokenType.
const biometricRequestTokenType = "vnd.authgear.biometric-request"

// biometricKeyID must satisfy biometric.KeyIDFormat.
const biometricKeyID = "e2e-biometric-key"

// The biometric key identifies the identity the server creates at setup, so
// one key for the whole process keeps it stable across the requests that set
// it up and then authenticate with it.
var (
	biometricKeyOnce sync.Once
	biometricKey     jwk.Key
	biometricPubKey  jwk.Key
)

func getBiometricKey() (private jwk.Key, public jwk.Key) {
	biometricKeyOnce.Do(func() {
		raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		biometricKey, err = jwk.FromRaw(raw)
		if err != nil {
			panic(err)
		}
		biometricPubKey, err = biometricKey.PublicKey()
		if err != nil {
			panic(err)
		}
		_ = biometricPubKey.Set(jwk.AlgorithmKey, jwa.ES256)
		_ = biometricPubKey.Set(jwk.KeyIDKey, biometricKeyID)
	})
	return biometricKey, biometricPubKey
}

// GenerateBiometricRequestJWT returns a biometric request JWT carrying the
// challenge. A setup request embeds the public key in the header; an
// authenticate request names it by key ID, as the SDKs do.
func GenerateBiometricRequestJWT(challenge string, action string) (string, error) {
	privateKey, publicKey := getBiometricKey()

	header := jws.NewHeaders()
	if err := header.Set("typ", biometricRequestTokenType); err != nil {
		return "", err
	}
	if err := header.Set("alg", "ES256"); err != nil {
		return "", err
	}
	if err := header.Set("kid", biometricKeyID); err != nil {
		return "", err
	}
	switch action {
	case "setup":
		if err := header.Set("jwk", publicKey); err != nil {
			return "", err
		}
	case "authenticate":
	default:
		return "", fmt.Errorf("unknown biometric action: %v", action)
	}

	payload := jwt.New()
	if err := payload.Set("challenge", challenge); err != nil {
		return "", err
	}
	if err := payload.Set("action", action); err != nil {
		return "", err
	}
	if err := payload.Set("device_info", map[string]any{
		"android": map[string]any{
			"Build": map[string]any{
				"MANUFACTURER": "Google",
				"MODEL":        "Pixel",
			},
		},
	}); err != nil {
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
