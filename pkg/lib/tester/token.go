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
