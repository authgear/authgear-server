package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/wechat"
)

type IdentificationData struct {
	TypedData
	Options []IdentificationOption `json:"options"`
}

func NewIdentificationData(d IdentificationData) IdentificationData {
	d.Type = DataTypeIdentificationData
	return d
}

var _ authflow.Data = IdentificationData{}

func (IdentificationData) Data() {}

type IdentificationOption struct {
	Identification config.AuthenticationFlowIdentification `json:"identification"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// ProviderType is specific to OAuth.
	ProviderType string `json:"provider_type,omitempty"`
	// Alias is specific to OAuth.
	Alias string `json:"alias,omitempty"`
	// WechatAppType is specific to OAuth.
	WechatAppType wechat.AppType `json:"wechat_app_type,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`
}

func NewIdentificationOptionIDToken(i config.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
	}
}

func NewIdentificationOptionLoginID(i config.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
	}
}

func NewIdentificationOptionsOAuth(oauthConfig *config.OAuthSSOConfig, oauthFeatureConfig *config.OAuthSSOProvidersFeatureConfig, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) []IdentificationOption {
	output := []IdentificationOption{}
	for _, p := range oauthConfig.Providers {
		if !identity.IsOAuthSSOProviderTypeDisabled(p.AsProviderConfig(), oauthFeatureConfig) {
			output = append(output, IdentificationOption{
				Identification: config.AuthenticationFlowIdentificationOAuth,
				BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
				ProviderType:   p.AsProviderConfig().Type(),
				Alias:          p.Alias(),
				WechatAppType:  wechat.ProviderConfig(p).AppType(),
			})
		}
	}
	return output
}

func NewIdentificationOptionPasskey(requestOptions *model.WebAuthnRequestOptions, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: config.AuthenticationFlowIdentificationPasskey,
		BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
		RequestOptions: requestOptions,
	}
}

func (i *IdentificationOption) isBotProtectionRequired() bool {
	if i.BotProtection == nil {
		return false
	}
	if i.BotProtection.Enabled != nil && *i.BotProtection.Enabled && i.BotProtection.Provider != nil && i.BotProtection.Provider.Type != "" {
		return true
	}

	return false
}
