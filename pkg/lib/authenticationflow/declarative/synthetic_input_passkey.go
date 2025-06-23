package declarative

import (
	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type SyntheticInputPasskey struct {
	Identification    model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	AssertionResponse *protocol.CredentialAssertionResponse  `json:"assertion_response,omitempty"`
	BotProtection     *InputTakeBotProtectionBody            `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &SyntheticInputPasskey{}
var _ inputTakeIdentificationMethod = &SyntheticInputPasskey{}
var _ inputTakePasskeyAssertionResponse = &SyntheticInputPasskey{}
var _ inputTakeBotProtection = &SyntheticInputPasskey{}

func (*SyntheticInputPasskey) Input() {}

func (i *SyntheticInputPasskey) GetIdentificationMethod() model.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *SyntheticInputPasskey) GetAssertionResponse() *protocol.CredentialAssertionResponse {
	return i.AssertionResponse
}

func (i *SyntheticInputPasskey) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *SyntheticInputPasskey) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *SyntheticInputPasskey) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
