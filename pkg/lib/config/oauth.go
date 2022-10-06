package config

import "net/url"

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

type OAuthClientApplicationType string

const (
	OAuthClientApplicationTypeSPA            OAuthClientApplicationType = "spa"
	OAuthClientApplicationTypeTraditionalWeb OAuthClientApplicationType = "traditional_webapp"
	OAuthClientApplicationTypeNative         OAuthClientApplicationType = "native"
	OAuthClientApplicationTypeThirdPartyApp  OAuthClientApplicationType = "third_party_app"
	OAuthClientApplicationTypeUnspecified    OAuthClientApplicationType = ""
)

type ClientParty string

const (
	ClientPartyFirst ClientParty = "first_party"
	ClientPartyThird ClientParty = "third_party"
)

var _ = Schema.Add("OAuthClientConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client_id": { "type": "string" },
		"client_uri": { "type": "string", "format": "uri" },
		"client_name": { "type": "string", "minLength": 1 },
		"name": { "type": "string" },
		"x_application_type": { "type": "string", "enum": ["spa", "traditional_webapp", "native", "third_party_app"] },
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
		"refresh_token_idle_timeout_enabled": { "type": "boolean" },
		"refresh_token_idle_timeout_seconds": { "$ref": "#/$defs/DurationSeconds" },
		"issue_jwt_access_token": { "type": "boolean" },
		"policy_uri": { "type": "string", "format": "uri" },
		"tos_uri": { "type": "string", "format": "uri"  },
		"is_first_party": { "type": "boolean" }
	},
	"required": ["name", "client_id", "redirect_uris"],
	"allOf": [
		{
			"if": {
				"properties": {
					"x_application_type": {
						"enum": ["spa", "traditional_webapp"]
					}
				},
				"required": ["x_application_type"]
			},
			"then": {
				"properties": {
					"post_logout_redirect_uris": {
						"minItems": 1
					}
				},
				"required": ["post_logout_redirect_uris"]
			}
		},
		{
			"if": {
				"properties": {
					"x_application_type": { "enum": ["third_party_app"] }
				},
				"required": ["x_application_type"]
			},
			"then": {
				"required": ["client_name"]
			}
		}
	]
}
`)

type OAuthClientConfig struct {
	ClientID                       string                     `json:"client_id,omitempty"`
	ClientURI                      string                     `json:"client_uri,omitempty"`
	ClientName                     string                     `json:"client_name,omitempty"`
	Name                           string                     `json:"name,omitempty"`
	ApplicationType                OAuthClientApplicationType `json:"x_application_type,omitempty"`
	RedirectURIs                   []string                   `json:"redirect_uris,omitempty"`
	GrantTypes                     []string                   `json:"grant_types,omitempty"`
	ResponseTypes                  []string                   `json:"response_types,omitempty"`
	PostLogoutRedirectURIs         []string                   `json:"post_logout_redirect_uris,omitempty"`
	AccessTokenLifetime            DurationSeconds            `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime           DurationSeconds            `json:"refresh_token_lifetime_seconds,omitempty"`
	RefreshTokenIdleTimeoutEnabled *bool                      `json:"refresh_token_idle_timeout_enabled,omitempty"`
	RefreshTokenIdleTimeout        DurationSeconds            `json:"refresh_token_idle_timeout_seconds,omitempty"`
	IssueJWTAccessToken            bool                       `json:"issue_jwt_access_token,omitempty"`
	PolicyURI                      string                     `json:"policy_uri,omitempty"`
	TOSURI                         string                     `json:"tos_uri,omitempty"`
}

func (c *OAuthClientConfig) ClientParty() ClientParty {
	if c.ApplicationType == OAuthClientApplicationTypeThirdPartyApp {
		return ClientPartyThird
	}
	// Except OAuthClientApplicationTypeThirdPartyApp
	// All the other clients are first party client
	return ClientPartyFirst
}

func (c *OAuthClientConfig) SetDefaults() {
	if c.AccessTokenLifetime == 0 {
		c.AccessTokenLifetime = DefaultAccessTokenLifetime
	}

	if c.RefreshTokenLifetime == 0 {
		if c.AccessTokenLifetime > DefaultRefreshTokenLifetime {
			c.RefreshTokenLifetime = c.AccessTokenLifetime
		} else {
			c.RefreshTokenLifetime = DefaultRefreshTokenLifetime
		}
	}

	if c.RefreshTokenIdleTimeoutEnabled == nil {
		b := DefaultRefreshTokenIdleTimeoutEnabled
		c.RefreshTokenIdleTimeoutEnabled = &b
	}
	if c.RefreshTokenIdleTimeout == 0 {
		c.RefreshTokenIdleTimeout = DefaultRefreshTokenIdleTimeout
	}
}

// RedirectURIHosts derives the list of host from the RedirectURIs
// items may be duplicate
func (c *OAuthClientConfig) RedirectURIHosts() []string {
	result := []string{}
	for _, uri := range c.RedirectURIs {
		u, err := url.Parse(uri)
		if err == nil {
			if u.Scheme == "http" || u.Scheme == "https" {
				result = append(result, u.Host)
			}
		}
	}
	return result
}
