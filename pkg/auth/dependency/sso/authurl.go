package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type RedirectURLFunc func(urlPrefix *url.URL, providerConfig config.OAuthSSOProviderConfig) string

type authURLParams struct {
	redirectURI  string
	clientID     string
	scope        string
	encodedState string
	baseURL      string
	nonce        string
	responseMode string
	display      string
	accessType   string
	prompt       string
}

func authURL(params authURLParams) (string, error) {
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", params.clientID)
	v.Add("redirect_uri", params.redirectURI)
	v.Add("scope", params.scope)
	if params.nonce != "" {
		v.Add("nonce", params.nonce)
	}
	if params.responseMode != "" {
		v.Add("response_mode", params.responseMode)
	}
	if params.display != "" {
		v.Add("display", params.display)
	}
	if params.accessType != "" {
		v.Add("access_type", params.accessType)
	}
	if params.prompt != "" {
		v.Add("prompt", params.prompt)
	}
	v.Add("state", params.encodedState)

	return params.baseURL + "?" + v.Encode(), nil
}
