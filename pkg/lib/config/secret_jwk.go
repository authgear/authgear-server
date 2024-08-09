package config

import (
	"encoding/json"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

var _ = SecretConfigSchema.Add("JWK", `
{
	"type": "object",
	"properties": {
		"kid": { "type": "string" },
		"kty": { "type": "string" }
	},
	"required": ["kid", "kty"]
}
`)

type JWK struct {
	jwk.Key
}

var _ json.Marshaler = &JWK{}
var _ json.Unmarshaler = &JWK{}

func (c *JWK) MarshalJSON() ([]byte, error) {
	return c.Key.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *JWK) UnmarshalJSON(b []byte) error {
	key, err := jwk.ParseKey(b)
	if err != nil {
		return err
	}
	c.Key = key
	return nil
}

func (c *JWK) SensitiveStrings() []string {
	return nil
}
