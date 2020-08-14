package sso

type oauthErrorResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

func (r oauthErrorResp) AsError() error {
	return NewSSOFailed(SSOUnauthorized, "oauth failed")
}
