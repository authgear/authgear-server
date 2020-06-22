package config

import (
	"errors"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
)

func newBool(v bool) *bool { return &v }

func newInt(v int) *int { return &v }

// nonSecretJWKFields is a set of common non-sensitive field names in JWK
var nonSecretJWKFields = map[string]struct{}{
	"kid": {},
	"kty": {},
	"alg": {},
}

func extractJWKSecrets(keys []interface{}) []string {
	const minJWKSecretLength = 8

	var secrets []string
	var extractSecret func(v interface{})
	extractSecret = func(v interface{}) {
		switch v := v.(type) {
		case string:
			if len(v) >= minJWKSecretLength {
				secrets = append(secrets, v)
			}
		case []interface{}:
			for _, i := range v {
				extractSecret(i)
			}
		case map[string]interface{}:
			for _, i := range v {
				extractSecret(i)
			}
		}
	}

	for _, key := range keys {
		jwk := key.(map[string]interface{})
		for k, v := range jwk {
			if _, ok := nonSecretJWKFields[k]; ok {
				continue
			}
			extractSecret(v)
		}
	}
	return secrets
}

func ExtractOctetKey(set *jwk.Set, id string) ([]byte, error) {
	for _, key := range set.Keys {
		if key.KeyID() != id {
			continue
		}
		switch key.KeyType() {
		case jwa.OctetSeq:
			data, err := key.Materialize()
			if err != nil {
				return nil, err
			}
			return data.([]byte), nil
		default:
			return nil, errors.New("unexpected key type (key type should be octet)")
		}
	}

	return nil, errors.New("octet key not found")
}
