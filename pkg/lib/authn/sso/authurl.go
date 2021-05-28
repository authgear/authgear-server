package sso

import (
	"net/url"
	"strings"
)

type authURLParams struct {
	redirectURI string
	clientID    string
	scope       string
	state       string
	baseURL     string
	prompt      []string
}

func authURL(params authURLParams) (string, error) {
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", params.clientID)
	v.Add("redirect_uri", params.redirectURI)
	v.Add("scope", params.scope)
	v.Add("state", params.state)
	if len(params.prompt) > 0 {
		v.Add("prompt", strings.Join(params.prompt, " "))
	}
	return params.baseURL + "?" + v.Encode(), nil
}
