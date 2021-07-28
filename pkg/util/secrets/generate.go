package secrets

import (
	"crypto/rsa"
	mathrand "math/rand"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

// Alphabet is an alphabet for secret.
// Secret is not intended to be manually entered by human.
const Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"

func GenerateSecret(length int, rng *mathrand.Rand) string {
	return rand.StringWithAlphabet(length, Alphabet, rng)
}

func GenerateOctetKey(rng *mathrand.Rand) jwk.Key {
	key := []byte(GenerateSecret(32, rng))

	jwkKey, err := jwk.New(key)
	if err != nil {
		panic(err)
	}

	_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
	_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = jwkKey.Set(jwk.AlgorithmKey, "HS256")

	return jwkKey
}

func GenerateRSAKey(rng *mathrand.Rand) jwk.Key {
	privateKey, err := rsa.GenerateKey(rng, 2048)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(privateKey)
	if err != nil {
		panic(err)
	}
	_ = jwkKey.Set(jwk.KeyIDKey, uuid.New())
	_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = jwkKey.Set(jwk.AlgorithmKey, "RS256")

	return jwkKey
}
