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

type OAuthClientApplicationType string

const (
	OAuthClientApplicationTypeSPA            OAuthClientApplicationType = "spa"
	OAuthClientApplicationTypeTraditionalWeb OAuthClientApplicationType = "traditional_webapp"
	OAuthClientApplicationTypeNative         OAuthClientApplicationType = "native"
	OAuthClientApplicationTypeConfidential   OAuthClientApplicationType = "confidential"
	OAuthClientApplicationTypeThirdPartyApp  OAuthClientApplicationType = "third_party_app"
	OAuthClientApplicationTypeUnspecified    OAuthClientApplicationType = ""
)

func (t OAuthClientApplicationType) IsThirdParty() bool {
	switch t {
	case OAuthClientApplicationTypeSPA:
		return false
	case OAuthClientApplicationTypeTraditionalWeb:
		return false
	case OAuthClientApplicationTypeNative:
		return false
	case OAuthClientApplicationTypeConfidential:
		return false
	case OAuthClientApplicationTypeThirdPartyApp:
		return true
	default:
		return false
	}
}

func (t OAuthClientApplicationType) IsFirstParty() bool {
	return !t.IsThirdParty()
}

func (t OAuthClientApplicationType) IsConfidential() bool {
	switch t {
	case OAuthClientApplicationTypeSPA:
		return false
	case OAuthClientApplicationTypeTraditionalWeb:
		return false
	case OAuthClientApplicationTypeNative:
		return false
	case OAuthClientApplicationTypeConfidential:
		return true
	case OAuthClientApplicationTypeThirdPartyApp:
		return true
	default:
		return false
	}
}

func (t OAuthClientApplicationType) IsPublic() bool {
	return !t.IsConfidential()
}

func (t OAuthClientApplicationType) HasFullAccessScope() bool {
	switch t {
	case OAuthClientApplicationTypeSPA:
		return true
	case OAuthClientApplicationTypeTraditionalWeb:
		return true
	case OAuthClientApplicationTypeNative:
		return true
	case OAuthClientApplicationTypeConfidential:
		return false
	case OAuthClientApplicationTypeThirdPartyApp:
		return false
	default:
		return true
	}
}

func (t OAuthClientApplicationType) PIIAllowedInIDToken() bool {
	switch t {
	case OAuthClientApplicationTypeSPA:
		return false
	case OAuthClientApplicationTypeTraditionalWeb:
		return false
	case OAuthClientApplicationTypeNative:
		return false
	case OAuthClientApplicationTypeConfidential:
		return true
	case OAuthClientApplicationTypeThirdPartyApp:
		return true
	default:
		return false
	}
}

var _ = Schema.Add("OAuthClientConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"client_id": { "type": "string" },
		"client_uri": { "type": "string", "format": "uri" },
		"client_name": { "type": "string", "minLength": 1 },
		"name": { "type": "string" },
		"x_application_type": { "type": "string", "enum": ["spa", "traditional_webapp", "native", "confidential", "third_party_app"] },
		"x_max_concurrent_session": { "type": "integer", "enum": [0, 1] },
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
		"tos_uri": { "type": "string", "format": "uri" },
		"x_custom_ui_uri": { "type": "string", "format": "uri" },
		"x_app2app_enabled": { "type": "boolean" },
		"x_app2app_insecure_device_key_binding_enabled": { "type": "boolean" },
		"x_authentication_flow_allowlist": { "$ref": "#/$defs/AuthenticationFlowAllowlist" },
		"x_app_initiated_sso_to_web_enabled": { "type": "boolean" }
	},
	"required": ["name", "client_id", "redirect_uris"],
	"allOf": [
		{
			"if": {
				"properties": {
					"x_application_type": {
						"enum": ["traditional_webapp"]
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
					"x_application_type": { "enum": ["confidential", "third_party_app"] }
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
	ClientID                               string                       `json:"client_id,omitempty"`
	ClientURI                              string                       `json:"client_uri,omitempty"`
	ClientName                             string                       `json:"client_name,omitempty"`
	Name                                   string                       `json:"name,omitempty"`
	ApplicationType                        OAuthClientApplicationType   `json:"x_application_type,omitempty"`
	MaxConcurrentSession                   int                          `json:"x_max_concurrent_session,omitempty"`
	RedirectURIs                           []string                     `json:"redirect_uris,omitempty"`
	GrantTypes                             []string                     `json:"grant_types,omitempty"`
	ResponseTypes                          []string                     `json:"response_types,omitempty"`
	PostLogoutRedirectURIs                 []string                     `json:"post_logout_redirect_uris,omitempty"`
	AccessTokenLifetime                    DurationSeconds              `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime                   DurationSeconds              `json:"refresh_token_lifetime_seconds,omitempty"`
	RefreshTokenIdleTimeoutEnabled         *bool                        `json:"refresh_token_idle_timeout_enabled,omitempty"`
	RefreshTokenIdleTimeout                DurationSeconds              `json:"refresh_token_idle_timeout_seconds,omitempty"`
	IssueJWTAccessToken                    bool                         `json:"issue_jwt_access_token,omitempty"`
	PolicyURI                              string                       `json:"policy_uri,omitempty"`
	TOSURI                                 string                       `json:"tos_uri,omitempty"`
	CustomUIURI                            string                       `json:"x_custom_ui_uri,omitempty"`
	App2appEnabled                         bool                         `json:"x_app2app_enabled,omitempty"`
	App2appInsecureDeviceKeyBindingEnabled bool                         `json:"x_app2app_insecure_device_key_binding_enabled,omitempty"`
	AuthenticationFlowAllowlist            *AuthenticationFlowAllowlist `json:"x_authentication_flow_allowlist,omitempty"`
	AppInitiatedSSOToWebEnabled            bool                         `json:"x_app_initiated_sso_to_web_enabled,omitempty"`
}

var _ = Schema.Add("AuthenticationFlowAllowlist", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"groups": { "type": "array", "items": { "$ref": "#/$defs/AuthenticationFlowAllowlistGroup" } },
		"flows": { "type": "array", "items": { "$ref": "#/$defs/AuthenticationFlowAllowlistFlow" } }
	}
}
`)

var _ = Schema.Add("AuthenticationFlowAllowlistGroup", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "minLength": 1 }
	},
	"required": ["name"]
}
`)

var _ = Schema.Add("AuthenticationFlowAllowlistFlow", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": {
			"type": "string",
			"enum": [
				"signup",
				"promote",
				"login",
				"signup_login",
				"reauth",
				"account_recovery"
			]
		},
		"name": { "$ref": "#/$defs/AuthenticationFlowObjectName" }
	},
	"required": ["type", "name"]
}
`)

type AuthenticationFlowAllowlist struct {
	Groups []*AuthenticationFlowAllowlistGroup `json:"groups,omitempty"`
	Flows  []*AuthenticationFlowAllowlistFlow  `json:"flows,omitempty"`
}

type AuthenticationFlowAllowlistGroup struct {
	Name string `json:"name"`
}

type AuthenticationFlowAllowlistFlow struct {
	Type AuthenticationFlowType `json:"type"`
	Name string                 `json:"name"`
}

func (c *OAuthClientConfig) DefaultRedirectURI() string {
	if len(c.RedirectURIs) > 0 {
		return c.RedirectURIs[0]
	}

	return ""
}

func (c *OAuthClientConfig) IsThirdParty() bool {
	return c.ApplicationType.IsThirdParty()
}

func (c *OAuthClientConfig) IsFirstParty() bool {
	return c.ApplicationType.IsFirstParty()
}

func (c *OAuthClientConfig) IsConfidential() bool {
	return c.ApplicationType.IsConfidential()
}

func (c *OAuthClientConfig) IsPublic() bool {
	return c.ApplicationType.IsPublic()
}

func (c *OAuthClientConfig) HasFullAccessScope() bool {
	return c.ApplicationType.HasFullAccessScope()
}

func (c *OAuthClientConfig) PIIAllowedInIDToken() bool {
	return c.ApplicationType.PIIAllowedInIDToken()
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
