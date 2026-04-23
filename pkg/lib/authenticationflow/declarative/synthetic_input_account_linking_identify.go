package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// This input is for advancing the login flow with the conflicted existing identity
type SyntheticInputAccountLinkingIdentify struct {
	Identification model.AuthenticationFlowIdentification
	BotProtection  *InputTakeBotProtectionBody

	// For identification=email/phone/username
	LoginID string

	// For identification=oauth
	Alias        string
	RedirectURI  string
	ResponseMode string
}

// GetLoginID implements inputTakeLoginID.
func (i *SyntheticInputAccountLinkingIdentify) GetLoginID() string {
	return i.LoginID
}

// GetIdentificationMethod implements inputTakeIdentificationMethod.
func (i *SyntheticInputAccountLinkingIdentify) GetIdentificationMethod() model.AuthenticationFlowIdentification {
	return i.Identification
}

// GetOAuthAlias implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputAccountLinkingIdentify) GetOAuthAlias() string {
	return i.Alias
}

// GetOAuthRedirectURI implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputAccountLinkingIdentify) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

// GetOAuthResponseMode implements inputTakeOAuthAuthorizationRequest.
func (i *SyntheticInputAccountLinkingIdentify) GetOAuthResponseMode() string {
	return i.ResponseMode
}

func (*SyntheticInputAccountLinkingIdentify) Input() {}

var _ inputTakeIdentificationMethod = &SyntheticInputAccountLinkingIdentify{}
var _ inputTakeLoginID = &SyntheticInputAccountLinkingIdentify{}
var _ inputTakeOAuthAuthorizationRequest = &SyntheticInputAccountLinkingIdentify{}
var _ inputTakeBotProtection = &SyntheticInputAccountLinkingIdentify{}

func (i *SyntheticInputAccountLinkingIdentify) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *SyntheticInputAccountLinkingIdentify) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *SyntheticInputAccountLinkingIdentify) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
