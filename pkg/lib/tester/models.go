package tester

import "github.com/authgear/authgear-server/pkg/util/pkce"

type TesterToken struct {
	TokenID      string         `json:"token_id"`
	ReturnURI    string         `json:"return_uri"`
	PKCEVerifier *pkce.Verifier `json:"pkce_verifier"`
}

func NewTesterToken(returnURI string) *TesterToken {
	return &TesterToken{
		TokenID:      newTesterTokenID(),
		ReturnURI:    returnURI,
		PKCEVerifier: pkce.GenerateS256Verifier(),
	}
}

type TesterResult struct {
	ID        string                 `json:"id"`
	ReturnURI string                 `json:"return_uri"`
	UserInfo  map[string]interface{} `json:"user_info"`
}

func NewTesterResultFromToken(token *TesterToken, userInfo map[string]interface{}) *TesterResult {
	return &TesterResult{
		ID:        newTesterResultID(),
		ReturnURI: token.ReturnURI,
		UserInfo:  userInfo,
	}
}
