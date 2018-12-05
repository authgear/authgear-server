package sso

type authHandler struct {
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

	processAccessToken func(a accessToken) accessToken
	processPrincipalID func(p map[string]interface{}) string
}

func (h authHandler) handle() (string, error) {
	accessToken, err := fetchAccessToken(
		h.code,
		h.clientID,
		h.urlPrefix,
		h.providerName,
		h.clientSecret,
		h.accessTokenURL,
	)
	if err != nil {
		return "", err
	}

	if h.processAccessToken != nil {
		accessToken = h.processAccessToken(accessToken)
	}

	userProfile, err := fetchUserProfile(accessToken, h.userProfileURL)
	if err != nil {
		return "", err
	}

	principalID := ""
	if h.processPrincipalID != nil {
		principalID = h.processPrincipalID(userProfile)
	} else {
		principalID = processPrincipalID(userProfile)
	}

	return "", nil
}

func processPrincipalID(userProfile map[string]interface{}) string {
	id, ok := userProfile["id"].(string)
	if !ok {
		return ""
	}
	return id
}
