package authgear

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"strings"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/authgear", new(Authgear))
}

func toJavaScript(v interface{}) interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var out interface{}
	err = json.Unmarshal(b, &out)
	if err != nil {
		panic(err)
	}
	return out
}

func toJSONBytes(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

type Authgear struct{}

func (*Authgear) Uuid() string {
	return uuid.Must(uuid.NewRandom()).String()
}

func (*Authgear) GenerateRSAPrivateKeyInPKCS8PEM(bits int) string {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}

	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}
	var buf strings.Builder
	err = pem.Encode(&buf, block)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (*Authgear) JwkKeyFromPKCS8PEM(pkcs8Pem string) interface{} {
	key, err := jwk.ParseKey([]byte(pkcs8Pem), jwk.WithPEM(true))
	if err != nil {
		panic(err)
	}

	return toJavaScript(key)
}

func (*Authgear) JwkPublicKeyFromJWKPrivateKey(jwkValue interface{}) interface{} {
	jwkBytes := toJSONBytes(jwkValue)
	jwkPrivateKey, err := jwk.ParseKey(jwkBytes)
	if err != nil {
		panic(err)
	}
	jwkPublicKey, err := jwkPrivateKey.PublicKey()
	if err != nil {
		panic(err)
	}

	return toJavaScript(jwkPublicKey)
}

func (*Authgear) SignJWT(alg string, jwkValue interface{}, payload interface{}, headers map[string]interface{}) string {
	jwkBytes := toJSONBytes(jwkValue)
	jwkPrivateKey, err := jwk.ParseKey(jwkBytes)
	if err != nil {
		panic(err)
	}

	payloadBytes := toJSONBytes(payload)

	hdr := jws.NewHeaders()
	for k, v := range headers {
		switch k {
		case "jwk":
			jwkKey, err := jwk.ParseKey(toJSONBytes(v))
			if err != nil {
				panic(err)
			}
			err = hdr.Set(k, jwkKey)
		default:
			err = hdr.Set(k, v)
		}
		if err != nil {
			panic(err)
		}
	}

	b, err := jws.Sign(
		payloadBytes,
		jws.WithKey(
			jwa.SignatureAlgorithm(alg),
			jwkPrivateKey,
			jws.WithProtectedHeaders(hdr),
		),
	)
	if err != nil {
		panic(err)
	}

	return string(b)
}
