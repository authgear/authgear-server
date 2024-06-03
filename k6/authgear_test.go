package authgear

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestUUID(t *testing.T) {
	var a Authgear

	uuid := a.Uuid()
	if len(uuid) != 36 {
		t.Errorf("expected uuid to be 36 characters long")
	}
}

func TestKeyGeneration(t *testing.T) {
	var a Authgear

	pkcs8pem := a.GenerateRSAPrivateKeyInPKCS8PEM(128)
	jwkValue := a.JwkKeyFromPKCS8PEM(pkcs8pem)
	jwkValueObj := jwkValue.(map[string]interface{})
	if jwkValueObj["kty"].(string) != "RSA" {
		t.Errorf("expected a RSA private key: %v", jwkValue)
	}

	pubKey := a.JwkPublicKeyFromJWKPrivateKey(jwkValue)
	pubKeyObj := pubKey.(map[string]interface{})
	if pubKeyObj["kty"].(string) != "RSA" {
		t.Errorf("expected a RSA public key: %v", pubKeyObj)
	}
}

func TestSignJWT(t *testing.T) {
	var a Authgear

	alg := "RS256"
	pkcs8pem := a.GenerateRSAPrivateKeyInPKCS8PEM(1024)
	jwkValue := a.JwkKeyFromPKCS8PEM(pkcs8pem)
	pubKey := a.JwkPublicKeyFromJWKPrivateKey(jwkValue)
	payload := map[string]interface{}{
		"hello": "world",
	}
	headers := map[string]interface{}{
		"jwk": pubKey,
	}
	jwt := a.SignJWT(alg, jwkValue, payload, headers)

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		t.Errorf("unexpected JWT token: %v", jwt)
	}

	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Errorf("failed to decode JWT headers: %v", err)
	}

	var hdr map[string]interface{}
	err = json.Unmarshal(hdrBytes, &hdr)
	if err != nil {
		t.Errorf("failed to parse JWT headers: %v", err)
	}

	_, ok := hdr["jwk"]
	if !ok {
		t.Errorf("expected jwk to be present in header")
	}
}
