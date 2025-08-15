package config

type ExternalJWTConfig struct {
	Issuers []ExternalJWTIssuerConfig `json:"issuers,omitempty"`
}

type ExternalJWTIssuerConfig struct {
	Iss     string `json:"iss"`
	Aud     string `json:"aud,omitempty"`
	JWKSURI string `json:"jwks_uri,omitempty"`
}

var _ = Schema.Add("ExternalJWTConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"issuers": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/ExternalJWTIssuerConfig"
			}
		}
	}
}
`)

var _ = Schema.Add("ExternalJWTIssuerConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"iss": { "type": "string" },
		"aud": { "type": "string" },
		"jwks_uri": { "type": "string", "format": "uri" }
	},
	"required": ["iss", "jwks_uri"]
}
`)
