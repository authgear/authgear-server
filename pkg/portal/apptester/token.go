package apptester

type AppTesterToken struct {
	TokenID   string `json:"token_id"`
	ReturnURI string `json:"return_uri"`
}

func NewTesterToken(returnURI string) *AppTesterToken {
	return &AppTesterToken{
		TokenID:   newTesterTokenID(),
		ReturnURI: returnURI,
	}
}
