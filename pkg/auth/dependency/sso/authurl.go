package sso

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type RedirectURLProvider interface {
	SSOCallbackURL(providerConfig config.OAuthSSOProviderConfig) *url.URL
}

type RedirectURLFunc func(urlPrefix *url.URL, providerConfig config.OAuthSSOProviderConfig) string

type authURLParams struct {
	redirectURI string
	clientID    string
	scope       string
	state       string
	baseURL     string
}

func authURL(params authURLParams) (string, error) {
	v := url.Values{}
	v.Add("response_type", "code")
	v.Add("client_id", params.clientID)
	v.Add("redirect_uri", params.redirectURI)
	v.Add("scope", params.scope)
	v.Add("state", params.state)
	return params.baseURL + "?" + v.Encode(), nil
}
