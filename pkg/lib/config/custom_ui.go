package config

var _ = Schema.Add("CustomUIEndpointsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"authentication": { "type": "string", "format": "uri" },
		"login_link": { "type": "string", "format": "uri" },
		"account_recovery": { "type": "string", "format": "uri" }
	}
}
`)

// CustomUIEndpointsConfig is the pages of a Custom UI.
// See docs/specs/custom-ui.md.
type CustomUIEndpointsConfig struct {
	Authentication  string `json:"authentication,omitempty"`
	LoginLink       string `json:"login_link,omitempty"`
	AccountRecovery string `json:"account_recovery,omitempty"`
}

func (c *CustomUIEndpointsConfig) IsEmpty() bool {
	return c == nil || (c.Authentication == "" && c.LoginLink == "" && c.AccountRecovery == "")
}

func (c *UIConfig) customUIEndpoints() CustomUIEndpointsConfig {
	if c == nil || c.CustomUIEndpoints == nil {
		return CustomUIEndpointsConfig{}
	}
	return *c.CustomUIEndpoints
}

// CustomUIAuthenticationURI returns the authentication page the client itself
// sets, ignoring the project default.
func (c *OAuthClientConfig) CustomUIAuthenticationURI() string {
	if c == nil {
		return ""
	}
	return c.CustomUIURI
}

// ResolveCustomUIAuthenticationURI returns "" when Auth UI should be used.
// client may be nil.
func (c *UIConfig) ResolveCustomUIAuthenticationURI(client *OAuthClientConfig) string {
	if uri := client.CustomUIAuthenticationURI(); uri != "" {
		return uri
	}
	return c.customUIEndpoints().Authentication
}

// ResolveCustomUILoginLinkURI returns "" when Auth UI should be used.
func (c *UIConfig) ResolveCustomUILoginLinkURI() string {
	return c.customUIEndpoints().LoginLink
}

// ResolveCustomUIAccountRecoveryURI returns "" when Auth UI should be used.
func (c *UIConfig) ResolveCustomUIAccountRecoveryURI() string {
	return c.customUIEndpoints().AccountRecovery
}

func appendNonEmpty(out []string, values ...string) []string {
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

// CustomUIAllURIs returns every configured Custom UI URI, for CORS.
func CustomUIAllURIs(ui *UIConfig, oauth *OAuthConfig) []string {
	project := ui.customUIEndpoints()
	out := appendNonEmpty(nil, project.Authentication, project.LoginLink, project.AccountRecovery)
	return appendClientCustomUIURIs(out, oauth)
}

// CustomUIAuthenticationURIs returns every URI that can resolve to the
// authentication page, for passkey origins and frame-ancestors.
func CustomUIAuthenticationURIs(ui *UIConfig, oauth *OAuthConfig) []string {
	out := appendNonEmpty(nil, ui.customUIEndpoints().Authentication)
	return appendClientCustomUIURIs(out, oauth)
}

func appendClientCustomUIURIs(out []string, oauth *OAuthConfig) []string {
	if oauth == nil {
		return out
	}
	for _, client := range oauth.Clients {
		out = appendNonEmpty(out, client.CustomUIURI)
	}
	return out
}
