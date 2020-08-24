package config

var _ = Schema.Add("HTTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"public_origin": { "type": "string" },
		"allowed_origins": { "type": "array", "items": { "type": "string" } },
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
