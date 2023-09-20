package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/go-webauthn/webauthn/protocol"
)

type SyntheticInputPasskey struct {
	Identification    config.AuthenticationFlowIdentification `json:"identification,omitempty"`
	AssertionResponse *protocol.CredentialAssertionResponse   `json:"assertion_response,omitempty"`
}

var _ authflow.Input = &SyntheticInputPasskey{}
var _ inputTakeIdentificationMethod = &SyntheticInputPasskey{}
var _ inputTakePasskeyAssertionResponse = &SyntheticInputPasskey{}

func (*SyntheticInputPasskey) Input() {}

func (i *SyntheticInputPasskey) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *SyntheticInputPasskey) GetAssertionResponse() *protocol.CredentialAssertionResponse {
	return i.AssertionResponse
}
