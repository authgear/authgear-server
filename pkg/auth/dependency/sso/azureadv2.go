package sso

import (
	"errors"
	"fmt"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const (
	azureadv2AuthorizationURLFormat string = "https://login.microsoftonline.com/%s/oauth2/authorize"
	azureadv2TokenURLFormat         string = "https://login.microsoftonline.com/%s/oauth2/token"
	azureadv2UserInfoURLFormat      string = "https://login.microsoftonline.com/%s/openid/userinfo"
)

type Azureadv2Impl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *Azureadv2Impl) GetAuthURL(params GetURLParams) (string, error) {
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        fmt.Sprintf(azureadv2AuthorizationURLFormat, f.ProviderConfig.Tenant),
	}
	return authURL(p)
}

func (f *Azureadv2Impl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *Azureadv2Impl) GetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	return f.OpenIDConnectGetAuthInfo(r)
}

func (f *Azureadv2Impl) OpenIDConnectGetAuthInfo(r OAuthAuthorizationResponse) (authInfo AuthInfo, err error) {
	err = errors.New("TODO")
	return
}

var (
	_ OAuthProvider         = &Azureadv2Impl{}
	_ OpenIDConnectProvider = &Azureadv2Impl{}
)
