package sso

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type authURLParams struct {
	oauthConfig    config.OAuthConfiguration
	providerConfig config.OAuthProviderConfiguration
	options        Options
	state          State
	baseURL        string
	nonce          string
	responseMode   string
}

func redirectURI(oauthConfig config.OAuthConfiguration, providerConfig config.OAuthProviderConfiguration) string {
	u, _ := url.Parse(oauthConfig.URLPrefix)
	orgPath := strings.TrimRight(u.Path, "/")
	path := fmt.Sprintf("%s/sso/%s/auth_handler", orgPath, providerConfig.ID)
	u.Path = path
	return u.String()
}

func authURL(params authURLParams) (string, error) {
	encodedState, err := EncodeState(params.oauthConfig.StateJWTSecret, params.state)
	if err != nil {
		return "", err
	}

	v := url.Values{}
	// Add user-specified options first
	// to avoid important params being overridden
	for k, o := range params.options {
		v.Add(k, fmt.Sprintf("%v", o))
	}

	v.Set("response_type", "code")
	v.Set("client_id", params.providerConfig.ClientID)
	v.Set("redirect_uri", redirectURI(params.oauthConfig, params.providerConfig))
	v.Set("state", encodedState)
	v.Set("scope", params.providerConfig.Scope)
	if params.nonce != "" {
		v.Set("nonce", params.nonce)
	}
	if params.responseMode != "" {
		v.Set("response_mode", params.responseMode)
	}

	return params.baseURL + "?" + v.Encode(), nil
}
