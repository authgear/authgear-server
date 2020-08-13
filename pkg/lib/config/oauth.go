package config

var _ = Schema.Add("OAuthConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"clients": { "type": "array", "items": { "$ref": "#/$defs/OAuthClientConfig" } }
	}
}
`)

type OAuthConfig struct {
	Clients []OAuthClientConfig `json:"clients,omitempty"`
}

func (c *OAuthConfig) GetClient(clientID string) (OAuthClientConfig, bool) {
	for _, c := range c.Clients {
		if c.ClientID() == clientID {
			return c, true
		}
	}
	return nil, false
}

var _ = Schema.Add("OAuthClientConfig", `
{
	"type": "object",
	"additionalProperties": false,
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
		"post_logout_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } },
		"access_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"refresh_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" }
	},
	"required": ["client_id", "redirect_uris"]
}
`)

type OAuthClientConfig map[string]interface{}

func (c OAuthClientConfig) SetDefaults() {
	if c.AccessTokenLifetime() == 0 {
		c.SetAccessTokenLifetime(1800)
	}
	if c.RefreshTokenLifetime() == 0 {
		if c.AccessTokenLifetime() > 86400 {
			c.SetRefreshTokenLifetime(c.AccessTokenLifetime())
		} else {
			c.SetRefreshTokenLifetime(86400)
		}
	}
}

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
func (c OAuthClientConfig) AccessTokenLifetime() DurationSeconds {
	if f64, ok := c["access_token_lifetime_seconds"].(float64); ok {
		return DurationSeconds(f64)
	}
	return 0
}

func (c OAuthClientConfig) SetAccessTokenLifetime(t DurationSeconds) {
	c["access_token_lifetime_seconds"] = float64(t)
}

func (c OAuthClientConfig) RefreshTokenLifetime() DurationSeconds {
	if f64, ok := c["refresh_token_lifetime_seconds"].(float64); ok {
		return DurationSeconds(f64)
	}
	return 0
}

func (c OAuthClientConfig) SetRefreshTokenLifetime(t DurationSeconds) {
	c["refresh_token_lifetime_seconds"] = float64(t)
}
