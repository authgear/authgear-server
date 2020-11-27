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

func (c *OAuthConfig) GetClient(clientID string) (*OAuthClientConfig, bool) {
	for _, c := range c.Clients {
		if c.ClientID == clientID {
			return &c, true
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
		"name": { "type": "string" },
		"redirect_uris": {
			"type": "array",
			"items": { "type": "string", "format": "uri" },
			"minItems": 1
		},
		"grant_types": { "type": "array", "items": { "type": "string" } },
		"response_types": { "type": "array", "items": { "type": "string" } },
		"post_logout_redirect_uris": { "type": "array", "items": { "type": "string", "format": "uri" } },
		"access_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds", "minimum": 300 },
		"refresh_token_lifetime_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"issue_jwt_access_token": { "type": "boolean" }
	},
	"required": ["name", "client_id", "redirect_uris"]
}
`)

type OAuthClientConfig struct {
	ClientID               string          `json:"client_id,omitempty"`
	ClientURI              string          `json:"client_uri,omitempty"`
	Name                   string          `json:"name,omitempty"`
	RedirectURIs           []string        `json:"redirect_uris,omitempty"`
	GrantTypes             []string        `json:"grant_types,omitempty"`
	ResponseTypes          []string        `json:"response_types,omitempty"`
	PostLogoutRedirectURIs []string        `json:"post_logout_redirect_uris,omitempty"`
	AccessTokenLifetime    DurationSeconds `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime   DurationSeconds `json:"refresh_token_lifetime_seconds,omitempty"`
	IssueJWTAccessToken    bool            `json:"issue_jwt_access_token,omitempty"`
}

func (c *OAuthClientConfig) SetDefaults() {
	if c.AccessTokenLifetime == 0 {
		c.AccessTokenLifetime = DefaultAccessTokenLifetime
	}
	if c.RefreshTokenLifetime == 0 {
		if c.AccessTokenLifetime > DefaultSessionLifetime {
			c.RefreshTokenLifetime = c.AccessTokenLifetime
		} else {
			c.RefreshTokenLifetime = DefaultSessionLifetime
		}
	}
}
