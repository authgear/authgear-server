package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type getAuthInfoRequest struct {
	oauthConfig    config.OAuthConfiguration
	providerConfig config.OAuthProviderConfiguration
	code           string
	accessTokenURL string
	encodedState   string
	userProfileURL string
	processor      AuthInfoProcessor
}

func (h getAuthInfoRequest) getAuthInfo() (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig: h.providerConfig,
	}

	state, err := DecodeState(h.oauthConfig.StateJWTSecret, h.encodedState)
	if err != nil {
		return
	}
	authInfo.State = state

	r, err := fetchAccessTokenResp(
		h.code,
		h.accessTokenURL,
		h.oauthConfig,
		h.providerConfig,
	)
	if err != nil {
		return
	}

	accessTokenResp, err := h.processor.DecodeAccessTokenResp(r)
	if err != nil {
		return
	}
	authInfo.ProviderAccessTokenResp = accessTokenResp

	err = h.processor.ValidateAccessTokenResp(accessTokenResp)
	if err != nil {
		return
	}

	return h.getAuthInfoByAccessTokenResp(accessTokenResp)
}

func (h getAuthInfoRequest) getAuthInfoByAccessTokenResp(accessTokenResp AccessTokenResp) (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderConfig: h.providerConfig,
		// validated accessTokenResp
		ProviderAccessTokenResp: accessTokenResp,
	}

	var state State
	if h.encodedState != "" {
		state, err = DecodeState(h.oauthConfig.StateJWTSecret, h.encodedState)
		if err != nil {
			return
		}
	}
	authInfo.State = state

	userProfile, err := fetchUserProfile(accessTokenResp, h.userProfileURL)
	if err != nil {
		return
	}
	authInfo.ProviderRawProfile = userProfile
	authInfo.ProviderUserInfo = h.processor.DecodeUserInfo(userProfile)

	return
}
