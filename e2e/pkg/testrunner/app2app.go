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

// app2appRequestTokenType mirrors app2app.RequestTokenType.
const app2appRequestTokenType = "vnd.authgear.app2app-request"

// app2appKeyID must satisfy app2app.KeyIDFormat.
const app2appKeyID = "e2e-app2app-device-key"

// The app2app device key belongs to the client, and the server learns it from
// the JWT. One key for the whole process keeps it stable across the requests
// that bind it and then use it.
var (
	app2appKeyOnce sync.Once
	app2appKey     jwk.Key
	app2appPubKey  jwk.Key
)

func getApp2AppKey() (private jwk.Key, public jwk.Key) {
	app2appKeyOnce.Do(func() {
		raw, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		app2appKey, err = jwk.FromRaw(raw)
		if err != nil {
			panic(err)
		}
		app2appPubKey, err = app2appKey.PublicKey()
		if err != nil {
			panic(err)
		}
		_ = app2appPubKey.Set(jwk.AlgorithmKey, jwa.ES256)
		_ = app2appPubKey.Set(jwk.KeyIDKey, app2appKeyID)
	})
	return app2appKey, app2appPubKey
}

// GenerateApp2AppJWT returns an app2app request JWT carrying the challenge,
// signed by the device key the test holds.
func GenerateApp2AppJWT(challenge string) (string, error) {
	privateKey, publicKey := getApp2AppKey()

	header := jws.NewHeaders()
	if err := header.Set("typ", app2appRequestTokenType); err != nil {
		return "", err
	}
	if err := header.Set("alg", "ES256"); err != nil {
		return "", err
	}
	// jws.Verify with a key set matches on the protected header's kid, so it
	// must be present alongside the embedded jwk.
	if err := header.Set("kid", app2appKeyID); err != nil {
		return "", err
	}
	if err := header.Set("jwk", publicKey); err != nil {
		return "", err
	}

	payload := jwt.New()
	if err := payload.Set("challenge", challenge); err != nil {
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
