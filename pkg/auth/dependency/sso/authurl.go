package sso

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type RedirectURLFunc func(urlPrefix *url.URL, providerConfig config.OAuthProviderConfiguration) string

type authURLParams struct {
	oauthConfig    *config.OAuthConfiguration
	redirectURI    string
	providerConfig config.OAuthProviderConfiguration
	encodedState   string
	baseURL        string
	nonce          string
	responseMode   string
	display        string
	accessType     string
	prompt         string
}

func authURL(params authURLParams) (string, error) {
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", params.providerConfig.ClientID)
	v.Add("redirect_uri", params.redirectURI)
	v.Add("scope", params.providerConfig.Scope)
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
