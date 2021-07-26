package secrets

import (
	"crypto/rsa"
	"encoding/base32"
	"io"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func GenerateKeyID(rand io.Reader) string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(id)
}

func GenerateOctetKey(rand io.Reader) jwk.Key {
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	jwkKey, err := jwk.New(key)
	if err != nil {
		panic(err)
	}

	id := make([]byte, 16)
	_, err = rand.Read(id)
	if err != nil {
		panic(err)
	}

	_ = jwkKey.Set(jwk.KeyIDKey, GenerateKeyID(rand))
	_ = jwkKey.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = jwkKey.Set(jwk.AlgorithmKey, "HS256")

	return jwkKey
}

func GenerateRSAKey(rand io.Reader) jwk.Key {
	privateKey, err := rsa.GenerateKey(rand, 2048)
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
