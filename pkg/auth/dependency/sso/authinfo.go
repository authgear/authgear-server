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

	err = h.processor.ValidateAccessTokenResp(accessTokenResp)
	if err != nil {
		return
	}

	userProfile, err := fetchUserProfile(accessTokenResp, h.userProfileURL)
	if err != nil {
		return
	}

	userID := h.processor.DecodeUserID(userProfile)
	// TODO: process process_userinfo_hook
	authData := h.processor.DecodeAuthData(userProfile)

	state, err := DecodeState(h.stateJWTSecret, h.encodedState)
	if err != nil {
		return
	}

	authInfo = AuthInfo{
		ProviderName:            h.providerName,
		State:                   state,
		ProviderUserID:          userID,
		ProviderUserProfile:     userProfile,
		ProviderAccessTokenResp: accessTokenResp,
		ProviderAuthData:        authData,
	}

	return
}
