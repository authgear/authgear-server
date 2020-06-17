package config

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
