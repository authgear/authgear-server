package config

var _ = Schema.Add("HTTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"public_origin": { "type": "string", "format": "http_origin" },
		"allowed_origins": { "type": "array", "items": { "type": "string", "minLength": 1 } },
		"cookie_prefix": { "type": "string" }
	},
	"required": [ "public_origin" ]
}
`)

type HTTPConfig struct {
	PublicOrigin   string   `json:"public_origin"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	CookiePrefix   string   `json:"cookie_prefix,omitempty"`
}
