package config

var _ = Schema.Add("HTTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"hosts": { "type": "array", "items": { "type": "string" } },
		"admin_hosts": { "type": "array", "items": { "type": "string" } },
		"allowed_origins": { "type": "array", "items": { "type": "string" } },
		"cookie_prefix": { "type": "string" }
	}
}
`)

type HTTPConfig struct {
	Hosts          []string `json:"hosts,omitempty"`
	AdminHosts     []string `json:"admin_hosts,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty"`
	CookiePrefix   string   `json:"cookie_prefix,omitempty"`
}
