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
	ClientID               string          `json:"client_id"`
	ClientURI              string          `json:"client_uri"`
	Name                   string          `json:"name"`
	RedirectURIs           []string        `json:"redirect_uris"`
	GrantTypes             []string        `json:"grant_types"`
	ResponseTypes          []string        `json:"response_types"`
	PostLogoutRedirectURIs []string        `json:"post_logout_redirect_uris"`
	AccessTokenLifetime    DurationSeconds `json:"access_token_lifetime_seconds"`
	RefreshTokenLifetime   DurationSeconds `json:"refresh_token_lifetime_seconds"`
	IssueJWTAccessToken    bool            `json:"issue_jwt_access_token"`
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
