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
	Identification model.AuthenticationFlowIdentification `json:"identification"`

	BotProtection *BotProtectionData `json:"bot_protection,omitempty"`
	// ProviderType is specific to OAuth.
	ProviderType string `json:"provider_type,omitempty"`
	// Alias is specific to OAuth.
	Alias string `json:"alias,omitempty"`
	// WechatAppType is specific to OAuth.
	WechatAppType wechat.AppType `json:"wechat_app_type,omitempty"`
	// ProviderStatus is specific to OAuth.
	ProviderStatus OAuthProviderStatus `json:"provider_status,omitempty"`

	// WebAuthnRequestOptions is specific to Passkey.
	RequestOptions *model.WebAuthnRequestOptions `json:"request_options,omitempty"`

	// Server is specific to LDAP
	ServerName string `json:"server_name,omitempty"`
}

func NewIdentificationOptionIDToken(flows authflow.Flows, i model.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
	}
}

func NewIdentificationOptionLoginID(flows authflow.Flows, i model.AuthenticationFlowIdentification, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: i,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
	}
}

func NewIdentificationOptionsOAuth(flows authflow.Flows, oauthConfig *config.OAuthSSOConfig, oauthFeatureConfig *config.OAuthSSOProvidersFeatureConfig, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig, demoCredentials *config.SSOOAuthDemoCredentials) []IdentificationOption {
	output := []IdentificationOption{}
	for _, p := range oauthConfig.Providers {
		if !identity.IsOAuthSSOProviderTypeDisabled(p.AsProviderConfig(), oauthFeatureConfig) {
			status := p.ComputeProviderStatus(demoCredentials)

			output = append(output, IdentificationOption{
				Identification: model.AuthenticationFlowIdentificationOAuth,
				BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
				ProviderType:   p.AsProviderConfig().Type(),
				Alias:          p.Alias(),
				WechatAppType:  wechat.ProviderConfig(p).AppType(),
				ProviderStatus: status,
			})
		}
	}
	return output
}

func NewIdentificationOptionPasskey(flows authflow.Flows, requestOptions *model.WebAuthnRequestOptions, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) IdentificationOption {
	return IdentificationOption{
		Identification: model.AuthenticationFlowIdentificationPasskey,
		BotProtection:  GetBotProtectionData(flows, authflowCfg, appCfg),
		RequestOptions: requestOptions,
	}
}

func NewIdentificationOptionLDAP(ldapConfig *config.LDAPConfig, authflowCfg *config.AuthenticationFlowBotProtection, appCfg *config.BotProtectionConfig) []IdentificationOption {
	output := []IdentificationOption{}
	for _, s := range ldapConfig.Servers {
		output = append(output, IdentificationOption{
			Identification: model.AuthenticationFlowIdentificationLDAP,
			ServerName:     s.Name,
			// TODO(DEV-1659): Support bot protection in LDAP
			// BotProtection:  GetBotProtectionData(authflowCfg, appCfg),
		})
	}
	return output
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
