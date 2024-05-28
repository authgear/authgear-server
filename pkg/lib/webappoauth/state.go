package webappoauth

import (
	"encoding/base64"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type WebappOAuthState struct {
	UIImplementation config.UIImplementation `json:"ui_implementation"`
	WebSessionID     string                  `json:"web_session_id"`

	// authflow, authflowv2 specific fields
	XStep            string `json:"x_step"`
	ErrorRedirectURI string `json:"error_redirect_uri"`
}

func (s WebappOAuthState) Encode() string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodeWebappOAuthState(stateStr string) (*WebappOAuthState, error) {
	b, err := base64.RawURLEncoding.DecodeString(stateStr)
	if err != nil {
		return nil, err
	}

	var state WebappOAuthState
	err = json.Unmarshal(b, &state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}
