package config

type OAuthConfig struct {
	AccessTokenLifetime  DurationSeconds     `json:"access_token_lifetime_seconds,omitempty"`
	RefreshTokenLifetime DurationSeconds     `json:"refresh_token_lifetime_seconds,omitempty"`
	Clients              []OAuthClientConfig `json:"clients,omitempty"`
}

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
