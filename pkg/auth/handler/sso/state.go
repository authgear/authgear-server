package sso

type AuthAPISSOState map[string]string

func (c AuthAPISSOState) CallbackURL() string {
	if s, ok := c["callback_url"]; ok {
		return s
	}
	return ""
}

// CodeChallenge is borrowed from PKCE.
// See https://www.oauth.com/oauth2-servers/pkce/authorization-request/
func (c AuthAPISSOState) CodeChallenge() string {
	if s, ok := c["code_challenge"]; ok {
		return s
	}
	return ""
}

func (c AuthAPISSOState) SetCallbackURL(s string) {
	c["callback_url"] = s
}

func (c AuthAPISSOState) SetCodeChallenge(s string) {
	c["code_challenge"] = s
}
