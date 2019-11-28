package sso

import (
	"fmt"
	"net/url"
	"path"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreUrl "github.com/skygeario/skygear-server/pkg/core/url"
)

type authURLParams struct {
	oauthConfig    *config.OAuthConfiguration
	urlPrefix      *url.URL
	providerConfig config.OAuthProviderConfiguration
	encodedState   string
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
	v.Add("state", params.encodedState)

	return params.baseURL + "?" + v.Encode(), nil
}
