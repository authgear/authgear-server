package sso

import (
	"fmt"
	"net/url"
	"path"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreUrl "github.com/skygeario/skygear-server/pkg/core/url"
)

// GetURLParams is the argument of getAuthURL
type GetURLParams struct {
	State State
}

type authURLParams struct {
	oauthConfig    *config.OAuthConfiguration
	urlPrefix      *url.URL
	providerConfig config.OAuthProviderConfiguration
	state          State
	baseURL        string
	nonce          string
	responseMode   string
	display        string
	accessType     string
	prompt         string
}

func redirectURI(urlPrefix *url.URL, providerConfig config.OAuthProviderConfiguration) string {
	u := *urlPrefix
	u.Path = path.Join(u.Path, fmt.Sprintf("_auth/sso/%s/auth_handler", url.PathEscape(providerConfig.ID)))
	return u.String()
}

func authURL(params authURLParams) (string, error) {
	encodedState, err := EncodeState(params.oauthConfig.StateJWTSecret, params.state)
	if err != nil {
		return "", err
	}

	v := coreUrl.Query{}
	v.Add("response_type", "code")
	v.Add("client_id", params.providerConfig.ClientID)
	v.Add("redirect_uri", redirectURI(params.urlPrefix, params.providerConfig))
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
	// Instagram quirk
	// state must be the last parameter otherwise
	// it will be converted to lowercase when
	// redirecting user to login page if user has not logged in before
	v.Add("state", encodedState)

	return params.baseURL + "?" + v.Encode(), nil
}
