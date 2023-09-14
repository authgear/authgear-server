package tester

type TesterToken struct {
	TokenID   string `json:"token_id"`
	ReturnURI string `json:"return_uri"`
}

func NewTesterToken(returnURI string) *TesterToken {
	return &TesterToken{
		TokenID:   newTesterTokenID(),
		ReturnURI: returnURI,
	}
}
