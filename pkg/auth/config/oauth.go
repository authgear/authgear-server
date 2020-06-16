package config

var _ = Schema.Add("OAuthConfig", `
{
	"type": "object",
	"properties": {
		"access_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"refresh_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"clients": { "type": "array", "items": { "$ref": "#/$defs/OAuthClientConfig" } }
	}
}
`)

type OAuthConfig struct {
	AccessTokenLifetime  DurationSeconds     `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime DurationSeconds     `json:"refresh_token_lifetime_seconds,omitempty"`
	Clients              []OAuthClientConfig `json:"clients,omitempty"`
}

func (c *OAuthConfig) SetDefaults() {
	if c.AccessTokenLifetime == 0 {
		c.AccessTokenLifetime = 1800
	}
	if c.RefreshTokenLifetime == 0 {
		c.RefreshTokenLifetime = 86400
	}
	if c.AccessTokenLifetime > c.RefreshTokenLifetime {
		c.RefreshTokenLifetime = c.AccessTokenLifetime
	}
}

var _ = Schema.Add("OAuthClientConfig", `
{
	"type": "object",
	"properties": {
		"client_id": { "type": "string" },
		"client_uri": { "type": "string", "format": "uri" },
		"redirect_uris": {
			"type": "array",
			"items": { "type": "string", "format": "uri" },
			"minItems": 1
		},
		"grant_types": { "type": "array", "items": { "type": "string" } },
		"response_types": { "type": "array", "items": { "type": "string" } },
		"post_logout_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } }
	},
	"required": ["client_id", "redirect_uris"]
}
`)

type OAuthClientConfig map[string]interface{}

func (c OAuthClientConfig) ClientID() string {
	if s, ok := c["client_id"].(string); ok {
		return s
	}
	return ""
}

func (c OAuthClientConfig) ClientURI() string {
	if s, ok := c["client_uri"].(string); ok {
		return s
	}
	return ""
}

func (c OAuthClientConfig) RedirectURIs() (out []string) {
	if arr, ok := c["redirect_uris"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return
}

func (c OAuthClientConfig) GrantTypes() (out []string) {
	if arr, ok := c["grant_types"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func (c OAuthClientConfig) ResponseTypes() (out []string) {
	if arr, ok := c["response_types"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}

func (c OAuthClientConfig) PostLogoutRedirectURIs() (out []string) {
	if arr, ok := c["post_logout_redirect_uris"].([]interface{}); ok {
		for _, item := range arr {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}
