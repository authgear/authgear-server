package sso

type getAuthInfoRequest struct {
	providerName   string
	clientID       string
	clientSecret   string
	urlPrefix      string
	code           string
	scope          Scope
	stateJWTSecret string
	encodedState   string
	accessTokenURL string
	userProfileURL string
	processor      AuthInfoProcessor
}

func (h getAuthInfoRequest) getAuthInfo() (authInfo AuthInfo, err error) {
	authInfo = AuthInfo{
		ProviderName: h.providerName,
	}

	state, err := DecodeState(h.stateJWTSecret, h.encodedState)
	if err != nil {
		return
	}
	authInfo.State = state

	r, err := fetchAccessTokenResp(
		h.code,
		h.clientID,
		h.urlPrefix,
		h.providerName,
		h.clientSecret,
		h.accessTokenURL,
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
		ProviderName: h.providerName,
		// validated accessTokenResp
		ProviderAccessTokenResp: accessTokenResp,
	}

	var state State
	if h.encodedState != "" {
		state, err = DecodeState(h.stateJWTSecret, h.encodedState)
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
