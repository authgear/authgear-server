package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type LinkedInImpl struct {
	OAuthConfig    config.OAuthConfiguration
	ProviderConfig config.OAuthProviderConfiguration
}

func (f *LinkedInImpl) GetAuthURL(params GetURLParams) (string, error) {
	if f.ProviderConfig.ClientID == "" {
		skyErr := skyerr.NewError(skyerr.InvalidArgument, "ClientID is required")
		return "", skyErr
	}
	p := authURLParams{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		options:        params.Options,
		state:          NewState(params),
		baseURL:        BaseURL(f.ProviderConfig),
	}
	return authURL(p)
}

func (f *LinkedInImpl) DecodeState(encodedState string) (*State, error) {
	return DecodeState(f.OAuthConfig.StateJWTSecret, encodedState)
}

func (f *LinkedInImpl) GetAuthInfo(code string, scope string, encodedState string) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		code:           code,
		encodedState:   encodedState,
		accessTokenURL: AccessTokenURL(f.ProviderConfig),
		userProfileURL: UserProfileURL(f.ProviderConfig),
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfo()
}

func (f *LinkedInImpl) GetAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	h := getAuthInfoRequest{
		oauthConfig:    f.OAuthConfig,
		providerConfig: f.ProviderConfig,
		accessTokenURL: AccessTokenURL(f.ProviderConfig),
		userProfileURL: UserProfileURL(f.ProviderConfig),
		processor:      newDefaultAuthInfoProcessor(),
	}
	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

var (
	_ Provider = &LinkedInImpl{}
)
