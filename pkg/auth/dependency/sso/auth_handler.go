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

	processAccessToken func(a accessToken) accessToken
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

	return "", nil
}
